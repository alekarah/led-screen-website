package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminDashboard - главная страница админки
func (h *Handlers) AdminDashboard(c *gin.Context) {
	var stats struct {
		ProjectsCount int64
		ImagesCount   int64
		ContactsCount int64
	}

	h.db.Model(&models.Project{}).Count(&stats.ProjectsCount)
	h.db.Model(&models.Image{}).Count(&stats.ImagesCount)
	h.db.Model(&models.ContactForm{}).Count(&stats.ContactsCount)

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title": "Админ панель",
		"stats": stats,
	})
}

// AdminProjects - управление проектами
func (h *Handlers) AdminProjects(c *gin.Context) {
	var projects []models.Project
	h.db.Preload("Categories").Preload("Images").Order("created_at DESC").Find(&projects)

	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "admin_projects.html", gin.H{
		"title":      "Управление проектами",
		"projects":   projects,
		"categories": categories,
	})
}

// CreateProject - создание нового проекта
func (h *Handlers) CreateProject(c *gin.Context) {
	var project models.Project

	// Получаем данные из формы
	project.Title = c.PostForm("title")
	project.Description = c.PostForm("description")
	project.Location = c.PostForm("location")
	project.Size = c.PostForm("size")
	project.PixelPitch = c.PostForm("pixel_pitch")
	project.Featured = c.PostForm("featured") == "on"

	// Генерируем slug из заголовка
	project.Slug = generateSlug(project.Title)

	// Сохраняем проект
	if err := h.db.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка создания проекта: " + err.Error(),
		})
		return
	}

	// Обрабатываем категории
	categoryIDs := c.PostFormArray("categories")
	for _, idStr := range categoryIDs {
		if id, err := strconv.Atoi(idStr); err == nil {
			var category models.Category
			if h.db.First(&category, id).Error == nil {
				h.db.Model(&project).Association("Categories").Append(&category)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Проект успешно создан",
		"project_id": project.ID,
	})
}

// GetProject - получение проекта для редактирования
func (h *Handlers) GetProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := h.db.Preload("Categories").Preload("Images").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Проект не найден",
		})
		return
	}

	var allCategories []models.Category
	h.db.Find(&allCategories)

	// Убеждаемся что все поля инициализированы
	if project.Categories == nil {
		project.Categories = []models.Category{}
	}
	if project.Images == nil {
		project.Images = []models.Image{}
	}

	c.JSON(http.StatusOK, gin.H{
		"project":    project,
		"categories": allCategories,
	})
}

// UpdateProject - редактирование проекта
func (h *Handlers) UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Проект не найден",
		})
		return
	}

	// Обновляем данные проекта
	project.Title = c.PostForm("title")
	project.Description = c.PostForm("description")
	project.Location = c.PostForm("location")
	project.Size = c.PostForm("size")
	project.PixelPitch = c.PostForm("pixel_pitch")
	project.Featured = c.PostForm("featured") == "on"

	// НЕ генерируем новый slug при редактировании, оставляем старый

	// Сохраняем изменения
	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка обновления проекта: " + err.Error(),
		})
		return
	}

	// Обновляем категории
	h.db.Model(&project).Association("Categories").Clear()
	categoryIDs := c.PostFormArray("categories")
	for _, idStr := range categoryIDs {
		if categoryId, err := strconv.Atoi(idStr); err == nil {
			var category models.Category
			if h.db.First(&category, categoryId).Error == nil {
				h.db.Model(&project).Association("Categories").Append(&category)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Проект успешно обновлен",
	})
}

// DeleteProject - удаление проекта
func (h *Handlers) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := h.db.Preload("Images").Preload("Categories").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Проект не найден",
		})
		return
	}

	// Удаляем файлы изображений
	for _, image := range project.Images {
		if err := os.Remove(image.FilePath); err != nil {
			// Логируем ошибку, но продолжаем
			fmt.Printf("Ошибка удаления файла %s: %v\n", image.FilePath, err)
		}
	}

	// Сначала удаляем связи с категориями
	if err := h.db.Model(&project).Association("Categories").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка удаления связей с категориями: " + err.Error(),
		})
		return
	}

	// Удаляем изображения из БД
	if err := h.db.Where("project_id = ?", project.ID).Delete(&models.Image{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка удаления изображений: " + err.Error(),
		})
		return
	}

	// Теперь удаляем сам проект
	if err := h.db.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка удаления проекта: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Проект успешно удален",
	})
}

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
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("project_%d_%d_%d%s", projectID, time.Now().Unix(), i, ext)

		// Путь для сохранения
		uploadPath := "../frontend/static/uploads"
		if err := os.MkdirAll(uploadPath, 0755); err != nil {
			continue
		}

		filePath := filepath.Join(uploadPath, filename)

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			continue
		}

		// Создаем запись в базе
		projectIDUint := uint(projectID)
		image := models.Image{
			ProjectID:    &projectIDUint,
			Filename:     filename,
			OriginalName: file.Filename,
			FilePath:     filePath,
			FileSize:     file.Size,
			MimeType:     file.Header.Get("Content-Type"),
			SortOrder:    i,
		}

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
	if err := os.Remove(image.FilePath); err != nil {
		// Логируем ошибку, но продолжаем удаление из БД
		fmt.Printf("Ошибка удаления файла %s: %v\n", image.FilePath, err)
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

	// Получаем новые параметры кроппинга
	cropXStr := c.PostForm("crop_x")
	cropYStr := c.PostForm("crop_y")
	cropScaleStr := c.PostForm("crop_scale")

	fmt.Printf("Получены параметры: cropX=%s, cropY=%s, cropScale=%s\n", cropXStr, cropYStr, cropScaleStr)

	cropX, err := strconv.ParseFloat(cropXStr, 64)
	if err != nil {
		fmt.Printf("Ошибка парсинга cropX: %v\n", err)
		cropX = 50
	}

	cropY, err := strconv.ParseFloat(cropYStr, 64)
	if err != nil {
		fmt.Printf("Ошибка парсинга cropY: %v\n", err)
		cropY = 50
	}

	cropScale, err := strconv.ParseFloat(cropScaleStr, 64)
	if err != nil {
		fmt.Printf("Ошибка парсинга cropScale: %v\n", err)
		cropScale = 1.0
	}

	// Преобразуем значения для translate (50% = центр = 0 смещения)
	translateX := cropX - 50 // -50 до +50
	translateY := cropY - 50 // -50 до +50

	fmt.Printf("Преобразованные значения: translateX=%.2f, translateY=%.2f, scale=%.2f\n", translateX, translateY, cropScale)

	// Валидация значений
	if cropX < 0 || cropX > 100 {
		cropX = 50
	}
	if cropY < 0 || cropY > 100 {
		cropY = 50
	}
	if cropScale < 0.5 || cropScale > 3.0 {
		cropScale = 1.0
	}

	fmt.Printf("Обновляем изображение %s: cropX=%.2f, cropY=%.2f, cropScale=%.2f\n", id, cropX, cropY, cropScale)

	// Обновляем настройки
	image.CropX = cropX - 50
	image.CropY = cropY - 50
	image.CropScale = cropScale

	if err := h.db.Save(&image).Error; err != nil {
		fmt.Printf("Ошибка сохранения: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения настроек: " + err.Error(),
		})
		return
	}

	fmt.Printf("Настройки успешно сохранены для изображения %s\n", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Настройки кроппинга обновлены",
		"image":   image,
	})
}
