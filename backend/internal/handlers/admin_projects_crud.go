package handlers

import (
	"net/http"
	"strconv"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

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
		if err := deleteImageFile(image.FilePath); err != nil {
			// Логируем ошибку, но продолжаем
			logError("Ошибка удаления файла", image.FilePath, err)
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
