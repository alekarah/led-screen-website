package handlers

import (
	"fmt"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID проекта",
		})
		return
	}

	// Проверяем что проект существует
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Проект не найден",
		})
		return
	}

	// Получаем загруженные файлы
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ошибка обработки файлов",
		})
		return
	}

	files := form.File["images"]
	var uploadedImages []models.Image

	for i, file := range files {
		// Проверяем тип файла
		if !isImageFile(file.Filename) {
			continue
		}

		// Генерируем уникальное имя файла
		filename := generateImageFilename(projectID, i, file.Filename)

		// Сохраняем файл
		filePath, err := saveUploadedFile(c, file, filename)
		if err != nil {
			logError("Ошибка сохранения файла", filename, err)
			continue
		}

		// Создаем запись в базе
		image := createImageRecord(projectID, filename, file, filePath, i)

		if err := h.db.Create(&image).Error; err == nil {
			uploadedImages = append(uploadedImages, image)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Загружено %d изображений", len(uploadedImages)),
		"images":  uploadedImages,
	})
}

// DeleteImage - удаление изображения
func (h *Handlers) DeleteImage(c *gin.Context) {
	id := c.Param("id")

	var image models.Image
	if err := h.db.First(&image, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Изображение не найдено",
		})
		return
	}

	// Удаляем файл с диска
	if err := deleteImageFile(image.FilePath); err != nil {
		// Логируем ошибку, но продолжаем удаление из БД
		logError("Ошибка удаления файла", image.FilePath, err)
	}

	// Удаляем запись из БД
	h.db.Delete(&image)

	c.JSON(http.StatusOK, gin.H{
		"message": "Изображение удалено",
	})
}

// UpdateImageCrop - обновление настроек кроппинга изображения
func (h *Handlers) UpdateImageCrop(c *gin.Context) {
	id := c.Param("id")

	var image models.Image
	if err := h.db.First(&image, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Изображение не найдено",
		})
		return
	}

	// Парсим параметры кроппинга
	cropData, err := parseCropParameters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные параметры кроппинга: " + err.Error(),
		})
		return
	}

	// Валидируем значения
	cropData = validateCropData(cropData)

	// Обновляем настройки изображения
	image.CropX = cropData.X - 50 // Сохраняем готовое смещение
	image.CropY = cropData.Y - 50 // Сохраняем готовое смещение
	image.CropScale = cropData.Scale

	if err := h.db.Save(&image).Error; err != nil {
		logError("Ошибка сохранения настроек кроппинга", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения настроек: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Настройки кроппинга обновлены",
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
func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, filename string) (string, error) {
	uploadPath := "../frontend/static/uploads"
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadPath, filename)
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
	if data.Scale < 0.5 || data.Scale > 3.0 {
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
