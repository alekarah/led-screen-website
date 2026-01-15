package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// UploadImages - загрузка изображений для проекта
func (h *Handlers) UploadImages(c *gin.Context) {
	projectIDStr := c.PostForm("project_id")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		log.Printf("[ERROR] Неверный project_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID проекта",
		})
		return
	}

	// Проверяем что проект существует
	var project models.Project
	if dbErr := h.db.First(&project, projectID).Error; dbErr != nil {
		log.Printf("[ERROR] Проект %d не найден: %v", projectID, dbErr)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Проект не найден",
		})
		return
	}

	// Получаем загруженные файлы
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("[ERROR] Ошибка обработки multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ошибка обработки файлов",
		})
		return
	}

	files := form.File["images"]

	var uploadedImages []models.Image

	// Проверяем, есть ли уже изображения у проекта
	var existingImagesCount int64
	h.db.Model(&models.Image{}).Where("project_id = ?", projectID).Count(&existingImagesCount)

	// Проверяем, есть ли уже главное изображение
	var hasPrimaryImage bool
	if existingImagesCount > 0 {
		var primaryCount int64
		h.db.Model(&models.Image{}).Where("project_id = ? AND is_primary = ?", projectID, true).Count(&primaryCount)
		hasPrimaryImage = primaryCount > 0
	}

	for i, file := range files {

		// Проверяем тип файла
		if !isImageFile(file.Filename) {
			log.Printf("[WARN] Файл %s пропущен: неподдерживаемый тип", file.Filename)
			continue
		}

		// Проверяем размер файла
		if file.Size > h.maxUploadSize {
			log.Printf("[WARN] Файл %s пропущен: слишком большой размер %d байт (максимум: %d)", file.Filename, file.Size, h.maxUploadSize)
			continue
		}

		// Генерируем уникальное имя файла
		filename := generateImageFilename(projectID, i, file.Filename)

		// Сохраняем файл
		filePath, err := h.saveUploadedFile(c, file, filename)
		if err != nil {
			log.Printf("[ERROR] Ошибка сохранения файла %s: %v", filename, err)
			continue
		}

		// Создаем запись в базе
		image := createImageRecord(projectID, filename, file, filePath, i)

		// Первое изображение проекта автоматически становится главным
		if !hasPrimaryImage && i == 0 {
			image.IsPrimary = true
			hasPrimaryImage = true // Чтобы остальные загружаемые изображения не стали главными
		}

		// Генерируем thumbnails с дефолтным кроппингом (без трансформаций)
		cropParams := CropParams{
			X:     50,
			Y:     50,
			Scale: 1.0,
		}
		thumbnails, err := GenerateThumbnails(filePath, cropParams)
		if err != nil {
			log.Printf("[WARN] Не удалось создать thumbnails для %s: %v", filename, err)
		} else {
			// Сохраняем пути к thumbnails в модель
			if path, ok := thumbnails[ThumbnailSmall.Suffix]; ok {
				image.ThumbnailSmallPath = path
			}
			if path, ok := thumbnails[ThumbnailMedium.Suffix]; ok {
				image.ThumbnailMediumPath = path
			}
		}

		if err := h.db.Create(&image).Error; err == nil {
			uploadedImages = append(uploadedImages, image)
		} else {
			log.Printf("[ERROR] Ошибка добавления изображения %s в БД: %v", filename, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Загружено %d изображений", len(uploadedImages)),
		"images":  uploadedImages,
	})
}

// DeleteImage - удаление изображения
func (h *Handlers) DeleteImage(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var image models.Image
	if err := h.db.First(&image, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Изображение не найдено")
		return
	}

	// Удаляем оригинал
	if err := deleteImageFile(image.FilePath); err != nil {
		logError("Ошибка удаления файла", image.FilePath, err)
	}

	// Удаляем все thumbnails
	DeleteThumbnails(image.FilePath)

	h.db.Delete(&image)
	jsonOK(c, gin.H{"message": "Изображение удалено"})
}

// UpdateImageCrop - обновление настроек кроппинга изображения
func (h *Handlers) UpdateImageCrop(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var image models.Image
	if err := h.db.First(&image, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Изображение не найдено")
		return
	}

	cropData, err := parseCropParameters(c)
	if err != nil {
		log.Printf("Неверные параметры кроппинга для изображения %d: %v", id, err)
		jsonErr(c, http.StatusBadRequest, "Неверные параметры кроппинга")
		return
	}
	cropData = validateCropData(cropData)

	// Обновляем параметры кроппинга
	image.CropX, image.CropY, image.CropScale = cropData.X, cropData.Y, cropData.Scale

	// Регенерируем thumbnails с новыми настройками кроппинга
	cropParams := CropParams(cropData)

	thumbnails, err := GenerateThumbnails(image.FilePath, cropParams)
	if err != nil {
		log.Printf("[ERROR] Не удалось регенерировать thumbnails для изображения %d: %v", id, err)
	} else {
		// Обновляем пути к thumbnails в модели
		if path, ok := thumbnails[ThumbnailSmall.Suffix]; ok {
			image.ThumbnailSmallPath = path
		}
		if path, ok := thumbnails[ThumbnailMedium.Suffix]; ok {
			image.ThumbnailMediumPath = path
		}
	}

	if err := h.db.Save(&image).Error; err != nil {
		log.Printf("Ошибка сохранения настроек кроппинга для изображения %d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка сохранения настроек")
		return
	}
	jsonOK(c, gin.H{"image": image})
}

// SetPrimaryImage - установка изображения как главного для проекта
func (h *Handlers) SetPrimaryImage(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	// Находим изображение
	var image models.Image
	if err := h.db.First(&image, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Изображение не найдено")
		return
	}

	// Проверяем что изображение привязано к проекту
	if image.ProjectID == nil {
		jsonErr(c, http.StatusBadRequest, "Изображение не привязано к проекту")
		return
	}

	// Начинаем транзакцию
	tx := h.db.Begin()

	// Сбрасываем флаг is_primary у всех изображений проекта
	if err := tx.Model(&models.Image{}).
		Where("project_id = ?", *image.ProjectID).
		Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		log.Printf("Ошибка сброса флага is_primary для проекта %d: %v", *image.ProjectID, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка установки главного изображения")
		return
	}

	// Устанавливаем флаг is_primary для выбранного изображения
	image.IsPrimary = true
	if err := tx.Save(&image).Error; err != nil {
		tx.Rollback()
		log.Printf("Ошибка установки флага is_primary для изображения %d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка установки главного изображения")
		return
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("Ошибка подтверждения транзакции для изображения %d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка установки главного изображения")
		return
	}

	jsonOK(c, gin.H{
		"message": "Главное изображение установлено",
		"image":   image,
	})
}

// Вспомогательные функции

// CropData структура для параметров кроппинга
type CropData struct {
	X     float64
	Y     float64
	Scale float64
}

// generateImageFilename генерирует уникальное имя файла
func generateImageFilename(projectID, index int, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return fmt.Sprintf("project_%d_%d_%d%s", projectID, time.Now().Unix(), index, ext)
}

// saveUploadedFile сохраняет загруженный файл
func (h *Handlers) saveUploadedFile(c *gin.Context, file *multipart.FileHeader, filename string) (string, error) {
	if err := os.MkdirAll(h.uploadPath, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(h.uploadPath, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

// createImageRecord создает запись изображения для БД
func createImageRecord(projectID int, filename string, file *multipart.FileHeader, filePath string, sortOrder int) models.Image {
	projectIDUint := uint(projectID)
	return models.Image{
		ProjectID:    &projectIDUint,
		Filename:     filename,
		OriginalName: file.Filename,
		FilePath:     filePath,
		FileSize:     file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		SortOrder:    sortOrder,
	}
}

// parseCropParameters парсит параметры кроппинга из JSON запроса
func parseCropParameters(c *gin.Context) (CropData, error) {
	var req struct {
		CropX     float64 `json:"crop_x"`
		CropY     float64 `json:"crop_y"`
		CropScale float64 `json:"crop_scale"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return CropData{}, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	return CropData{
		X:     req.CropX,
		Y:     req.CropY,
		Scale: req.CropScale,
	}, nil
}

// validateCropData валидирует параметры кроппинга
func validateCropData(data CropData) CropData {
	if data.X < 0 || data.X > 100 {
		data.X = 50
	}
	if data.Y < 0 || data.Y > 100 {
		data.Y = 50
	}
	if data.Scale < 1.0 || data.Scale > 3.0 {
		data.Scale = 1.0
	}
	return data
}

// deleteImageFile удаляет файл изображения
func deleteImageFile(filePath string) error {
	return os.Remove(filePath)
}

// logError логирует ошибки
func logError(message, context string, err error) {
	fmt.Printf("%s %s: %v\n", message, context, err)
}
