package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminPricesPage - страница управления ценами
func (h *Handlers) AdminPricesPage(c *gin.Context) {
	var priceItems []models.PriceItem
	if err := h.db.Order("sort_order ASC, id DESC").Find(&priceItems).Error; err != nil {
		log.Printf("Ошибка загрузки позиций прайса: %v", err)
		c.HTML(http.StatusInternalServerError, "admin_base.html", gin.H{
			"PageID": "admin-error",
			"error":  "Ошибка загрузки данных",
		})
		return
	}

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":      "Управление ценами",
		"PageID":     "admin-prices",
		"PriceItems": priceItems,
	})
}

// CreatePriceItem - создание новой позиции прайса
func (h *Handlers) CreatePriceItem(c *gin.Context) {
	var priceItem models.PriceItem

	// Получаем данные из формы
	priceItem.Title = strings.TrimSpace(c.PostForm("title"))
	priceItem.Description = strings.TrimSpace(c.PostForm("description"))
	priceFromStr := c.PostForm("price_from")
	priceItem.HasSpecifications = c.PostForm("has_specifications") == "on"
	priceItem.IsActive = c.PostForm("is_active") == "on"

	// Валидация
	if priceItem.Title == "" {
		jsonErr(c, http.StatusBadRequest, "Название не может быть пустым")
		return
	}

	// Парсим цену
	if priceFrom, err := strconv.Atoi(priceFromStr); err == nil && priceFrom >= 0 {
		priceItem.PriceFrom = priceFrom
	} else {
		jsonErr(c, http.StatusBadRequest, "Некорректная цена")
		return
	}

	// Обрабатываем загрузку изображения
	if file, err := c.FormFile("image"); err == nil {
		// Генерируем имя файла: price_{timestamp}_{original}.ext
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("price_%d_%s%s", time.Now().Unix(), generateSlug(priceItem.Title), ext)
		uploadPath := filepath.Join(h.uploadPath, filename)

		// Создаем директорию если её нет
		if err := os.MkdirAll(filepath.Dir(uploadPath), 0755); err != nil {
			log.Printf("Ошибка создания директории: %v", err)
			jsonErr(c, http.StatusInternalServerError, "Ошибка сохранения файла")
			return
		}

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			log.Printf("Ошибка сохранения файла: %v", err)
			jsonErr(c, http.StatusInternalServerError, "Ошибка загрузки изображения")
			return
		}

		// Сохраняем относительный веб-путь
		priceItem.ImagePath = "/static/uploads/" + filename

		// Генерируем миниатюры с дефолтным кроппингом
		cropParams := CropParams{
			X:     50,
			Y:     50,
			Scale: 1.0,
		}
		thumbnails, err := GenerateThumbnails(uploadPath, cropParams)
		if err != nil {
			log.Printf("Ошибка создания миниатюр для %s: %v", filename, err)
		} else {
			// Конвертируем пути миниатюр в веб-пути
			if path, ok := thumbnails[ThumbnailSmall.Suffix]; ok {
				priceItem.ThumbnailSmallPath = convertToWebPath(path)
			}
			if path, ok := thumbnails[ThumbnailMedium.Suffix]; ok {
				priceItem.ThumbnailMediumPath = convertToWebPath(path)
			}
			log.Printf("Миниатюры созданы для %s", filename)
		}
	}

	// Сохраняем позицию прайса
	if err := h.db.Create(&priceItem).Error; err != nil {
		log.Printf("Ошибка создания позиции прайса '%s': %v", priceItem.Title, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка создания позиции")
		return
	}

	// Если есть характеристики, обрабатываем их
	if priceItem.HasSpecifications {
		specificationsJSON := c.PostForm("specifications")
		if specificationsJSON != "" {
			var specs []struct {
				Group string `json:"group"`
				Key   string `json:"key"`
				Value string `json:"value"`
				Order int    `json:"order"`
			}

			if err := json.Unmarshal([]byte(specificationsJSON), &specs); err == nil {
				for _, spec := range specs {
					specification := models.PriceSpecification{
						PriceItemID: priceItem.ID,
						SpecGroup:   spec.Group,
						SpecKey:     spec.Key,
						SpecValue:   spec.Value,
						SortOrder:   spec.Order,
					}
					if err := h.db.Create(&specification).Error; err != nil {
						log.Printf("Ошибка создания характеристики: %v", err)
					}
				}
			}
		}
	}

	jsonOK(c, gin.H{
		"message":  "Позиция прайса успешно создана",
		"price_id": priceItem.ID,
	})
}

// GetPriceItem - получение позиции прайса для редактирования
func (h *Handlers) GetPriceItem(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var priceItem models.PriceItem
	if err := h.db.Preload("Specifications", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).First(&priceItem, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Позиция прайса не найдена")
		return
	}

	if priceItem.Specifications == nil {
		priceItem.Specifications = []models.PriceSpecification{}
	}

	jsonOK(c, gin.H{"price_item": priceItem})
}

// UpdatePriceItem - обновление позиции прайса
func (h *Handlers) UpdatePriceItem(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var priceItem models.PriceItem
	if err := h.db.First(&priceItem, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Позиция прайса не найдена")
		return
	}

	// Обновляем данные
	priceItem.Title = strings.TrimSpace(c.PostForm("title"))
	priceItem.Description = strings.TrimSpace(c.PostForm("description"))
	priceFromStr := c.PostForm("price_from")
	priceItem.HasSpecifications = c.PostForm("has_specifications") == "on"
	priceItem.IsActive = c.PostForm("is_active") == "on"

	// Валидация
	if priceItem.Title == "" {
		jsonErr(c, http.StatusBadRequest, "Название не может быть пустым")
		return
	}

	// Парсим цену
	if priceFrom, err := strconv.Atoi(priceFromStr); err == nil && priceFrom >= 0 {
		priceItem.PriceFrom = priceFrom
	} else {
		jsonErr(c, http.StatusBadRequest, "Некорректная цена")
		return
	}

	// Обрабатываем загрузку нового изображения
	if file, err := c.FormFile("image"); err == nil {
		// Удаляем старое изображение и миниатюры
		if priceItem.ImagePath != "" {
			oldFilename := filepath.Base(priceItem.ImagePath)
			oldFilePath := filepath.Join(h.uploadPath, oldFilename)
			os.Remove(oldFilePath)
			DeleteThumbnails(oldFilePath)
		}

		// Загружаем новое изображение
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("price_%d_%s%s", time.Now().Unix(), generateSlug(priceItem.Title), ext)
		uploadPath := filepath.Join(h.uploadPath, filename)

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			log.Printf("Ошибка сохранения файла: %v", err)
			jsonErr(c, http.StatusInternalServerError, "Ошибка загрузки изображения")
			return
		}

		// Сохраняем относительный веб-путь
		priceItem.ImagePath = "/static/uploads/" + filename

		// Генерируем миниатюры с сохраненными параметрами кроппинга или дефолтными
		cropParams := CropParams{
			X:     priceItem.CropX,
			Y:     priceItem.CropY,
			Scale: priceItem.CropScale,
		}
		if cropParams.X == 0 && cropParams.Y == 0 && cropParams.Scale == 0 {
			cropParams = CropParams{X: 50, Y: 50, Scale: 1.0}
		}

		thumbnails, err := GenerateThumbnails(uploadPath, cropParams)
		if err != nil {
			log.Printf("Ошибка создания миниатюр для %s: %v", filename, err)
		} else {
			if path, ok := thumbnails[ThumbnailSmall.Suffix]; ok {
				priceItem.ThumbnailSmallPath = convertToWebPath(path)
			}
			if path, ok := thumbnails[ThumbnailMedium.Suffix]; ok {
				priceItem.ThumbnailMediumPath = convertToWebPath(path)
			}
			log.Printf("Миниатюры обновлены для %s", filename)
		}
	}

	// Сохраняем изменения
	if err := h.db.Save(&priceItem).Error; err != nil {
		log.Printf("Ошибка обновления позиции прайса ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления позиции")
		return
	}

	// Удаляем старые характеристики
	h.db.Where("price_item_id = ?", priceItem.ID).Delete(&models.PriceSpecification{})

	// Если есть характеристики, сохраняем новые
	if priceItem.HasSpecifications {
		specificationsJSON := c.PostForm("specifications")
		if specificationsJSON != "" {
			var specs []struct {
				Group string `json:"group"`
				Key   string `json:"key"`
				Value string `json:"value"`
				Order int    `json:"order"`
			}

			if err := json.Unmarshal([]byte(specificationsJSON), &specs); err == nil {
				for _, spec := range specs {
					specification := models.PriceSpecification{
						PriceItemID: priceItem.ID,
						SpecGroup:   spec.Group,
						SpecKey:     spec.Key,
						SpecValue:   spec.Value,
						SortOrder:   spec.Order,
					}
					if err := h.db.Create(&specification).Error; err != nil {
						log.Printf("Ошибка создания характеристики: %v", err)
					}
				}
			}
		}
	}

	jsonOK(c, gin.H{"message": "Позиция прайса успешно обновлена"})
}

// DeletePriceItem - удаление позиции прайса
func (h *Handlers) DeletePriceItem(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var priceItem models.PriceItem
	if err := h.db.First(&priceItem, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Позиция прайса не найдена")
		return
	}

	// Удаляем изображение и миниатюры
	if priceItem.ImagePath != "" {
		// Извлекаем имя файла из ImagePath (/static/uploads/file.png -> file.png)
		filename := filepath.Base(priceItem.ImagePath)
		// Строим полный путь к файлу
		filePath := filepath.Join(h.uploadPath, filename)
		if err := os.Remove(filePath); err != nil {
			log.Printf("Ошибка удаления изображения %s: %v", filePath, err)
		}
		DeleteThumbnails(filePath)
	}

	// Удаляем позицию (характеристики удалятся автоматически через CASCADE)
	if err := h.db.Delete(&priceItem).Error; err != nil {
		log.Printf("Ошибка удаления позиции прайса ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления позиции")
		return
	}

	jsonOK(c, gin.H{"message": "Позиция прайса успешно удалена"})
}

// UpdatePriceItemsSorting - обновление порядка отображения позиций прайса
func (h *Handlers) UpdatePriceItemsSorting(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Некорректные данные")
		return
	}

	// Обновляем sort_order для каждой позиции
	for i, id := range req.IDs {
		if err := h.db.Model(&models.PriceItem{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
			log.Printf("Ошибка обновления порядка для позиции ID=%d: %v", id, err)
		}
	}

	jsonOK(c, gin.H{"message": "Порядок успешно обновлен"})
}

// UpdatePriceImageCrop - обновление настроек кроппинга изображения цены
func (h *Handlers) UpdatePriceImageCrop(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var priceItem models.PriceItem
	if err := h.db.First(&priceItem, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Позиция прайса не найдена")
		return
	}

	// Проверяем что изображение есть
	if priceItem.ImagePath == "" {
		jsonErr(c, http.StatusBadRequest, "У позиции нет изображения")
		return
	}

	// Парсим параметры кроппинга
	var req struct {
		CropX     float64 `json:"crop_x"`
		CropY     float64 `json:"crop_y"`
		CropScale float64 `json:"crop_scale"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Неверные параметры кроппинга для позиции %d: %v", id, err)
		jsonErr(c, http.StatusBadRequest, "Неверные параметры кроппинга")
		return
	}

	// Валидация
	if req.CropX < 0 || req.CropX > 100 {
		req.CropX = 50
	}
	if req.CropY < 0 || req.CropY > 100 {
		req.CropY = 50
	}
	if req.CropScale < 1.0 || req.CropScale > 3.0 {
		req.CropScale = 1.0
	}

	// Обновляем параметры кроппинга
	priceItem.CropX = req.CropX
	priceItem.CropY = req.CropY
	priceItem.CropScale = req.CropScale

	// Регенерируем миниатюры с новыми параметрами
	cropParams := CropParams{
		X:     req.CropX,
		Y:     req.CropY,
		Scale: req.CropScale,
	}

	// Извлекаем имя файла из ImagePath (/static/uploads/file.png -> file.png)
	filename := filepath.Base(priceItem.ImagePath)
	// Строим полный путь к файлу
	filePath := filepath.Join(h.uploadPath, filename)

	log.Printf("[DEBUG] UpdatePriceImageCrop: ID=%d, FilePath=%s, Crop=(%.1f, %.1f, %.1fx)",
		id, filePath, cropParams.X, cropParams.Y, cropParams.Scale)

	thumbnails, err := GenerateThumbnails(filePath, cropParams)
	if err != nil {
		log.Printf("[ERROR] Не удалось регенерировать миниатюры для позиции %d: %v", id, err)
	} else {
		if path, ok := thumbnails[ThumbnailSmall.Suffix]; ok {
			priceItem.ThumbnailSmallPath = convertToWebPath(path)
		}
		if path, ok := thumbnails[ThumbnailMedium.Suffix]; ok {
			priceItem.ThumbnailMediumPath = convertToWebPath(path)
		}
		log.Printf("[DEBUG] Миниатюры регенерированы для позиции %d с кроппингом (%.1f, %.1f, %.1fx)",
			id, req.CropX, req.CropY, req.CropScale)
	}

	if err := h.db.Save(&priceItem).Error; err != nil {
		log.Printf("Ошибка сохранения настроек кроппинга для позиции %d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка сохранения настроек")
		return
	}

	jsonOK(c, gin.H{"price_item": priceItem})
}

// DeletePriceImage - удаление изображения из позиции прайса
func (h *Handlers) DeletePriceImage(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var priceItem models.PriceItem
	if err := h.db.First(&priceItem, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Позиция прайса не найдена")
		return
	}

	// Проверяем что изображение есть
	if priceItem.ImagePath == "" {
		jsonErr(c, http.StatusBadRequest, "У позиции нет изображения")
		return
	}

	// Удаляем файлы
	filename := filepath.Base(priceItem.ImagePath)
	filePath := filepath.Join(h.uploadPath, filename)
	if err := os.Remove(filePath); err != nil {
		log.Printf("Ошибка удаления изображения %s: %v", filePath, err)
	}
	DeleteThumbnails(filePath)

	// Очищаем пути в БД
	priceItem.ImagePath = ""
	priceItem.ThumbnailSmallPath = ""
	priceItem.ThumbnailMediumPath = ""
	priceItem.CropX = 50
	priceItem.CropY = 50
	priceItem.CropScale = 1.0

	if err := h.db.Save(&priceItem).Error; err != nil {
		log.Printf("Ошибка обновления позиции %d после удаления изображения: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления позиции")
		return
	}

	jsonOK(c, gin.H{"message": "Изображение удалено"})
}

// convertToWebPath - конвертирует полный путь файловой системы в веб-путь
// frontend/static/uploads/file.png -> /static/uploads/file.png
// frontend\static\uploads\file.png -> /static/uploads/file.png
func convertToWebPath(fsPath string) string {
	// Нормализуем разделители
	normalized := filepath.ToSlash(fsPath)

	log.Printf("[DEBUG] convertToWebPath: input='%s' normalized='%s'", fsPath, normalized)

	// Ищем "/static/uploads/" в пути
	if idx := strings.Index(normalized, "/static/uploads/"); idx != -1 {
		result := normalized[idx:] // Возвращаем с /static/uploads/...
		log.Printf("[DEBUG] convertToWebPath: found at idx=%d result='%s'", idx, result)
		return result
	}

	// Если не нашли, возвращаем как есть (fallback)
	log.Printf("[DEBUG] convertToWebPath: not found, returning '%s'", normalized)
	return normalized
}

// convertToFilePath - конвертирует веб-путь в файловый путь
// /static/uploads/file.png -> frontend/static/uploads/file.png
func convertToFilePath(webPath string) string {
	// Убираем ведущий слеш если есть
	webPath = strings.TrimPrefix(webPath, "/")

	// Если путь начинается с static/uploads/, добавляем frontend/
	if strings.HasPrefix(webPath, "static/uploads/") {
		return filepath.Join("frontend", webPath)
	}

	// Если уже полный путь, возвращаем как есть
	return webPath
}
