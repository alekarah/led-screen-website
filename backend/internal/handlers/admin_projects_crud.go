package handlers

import (
	"log"
	"net/http"
	"strconv"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// Обеспечиваем уникальность slug (разрешаем одинаковые названия)
	baseSlug := project.Slug
	suffix := 1
	for {
		var count int64
		h.db.Model(&models.Project{}).Where("slug = ?", project.Slug).Count(&count)
		if count == 0 {
			break
		}
		suffix++
		project.Slug = baseSlug + "-" + strconv.Itoa(suffix)
	}

	// Сохраняем проект
	if err := h.db.Create(&project).Error; err != nil {
		log.Printf("Ошибка создания проекта '%s': %v", project.Title, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка создания проекта",
		})
		return
	}

	// Обрабатываем категории
	categoryIDs := c.PostFormArray("categories")
	for _, idStr := range categoryIDs {
		if id, err := strconv.Atoi(idStr); err == nil {
			var category models.Category
			if h.db.First(&category, id).Error == nil {
				if err := h.db.Model(&project).Association("Categories").Append(&category); err != nil {
					log.Printf("Ошибка добавления категории ID=%d к проекту ID=%d: %v", id, project.ID, err)
				}
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
	id, ok := mustID(c)
	if !ok {
		return
	}

	var project models.Project
	if err := h.db.Preload("Categories").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		First(&project, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Проект не найден")
		return
	}
	var allCategories []models.Category
	h.db.Find(&allCategories)

	if project.Categories == nil {
		project.Categories = []models.Category{}
	}
	if project.Images == nil {
		project.Images = []models.Image{}
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	jsonOK(c, gin.H{"project": project, "categories": allCategories})
}

// UpdateProject - редактирование проекта
func (h *Handlers) UpdateProject(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var project models.Project
	if err := h.db.First(&project, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Проект не найден")
		return
	}

	project.Title = c.PostForm("title")
	project.Description = c.PostForm("description")
	project.Location = c.PostForm("location")
	project.Size = c.PostForm("size")
	project.PixelPitch = c.PostForm("pixel_pitch")
	project.Featured = c.PostForm("featured") == "on"

	if err := h.db.Save(&project).Error; err != nil {
		log.Printf("Ошибка обновления проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка обновления проекта")
		return
	}

	if err := h.db.Model(&project).Association("Categories").Clear(); err != nil {
		log.Printf("Ошибка очистки категорий для проекта ID=%d: %v", project.ID, err)
	}
	for _, idStr := range c.PostFormArray("categories") {
		if categoryId, err := strconv.Atoi(idStr); err == nil {
			var category models.Category
			if h.db.First(&category, categoryId).Error == nil {
				if err := h.db.Model(&project).Association("Categories").Append(&category); err != nil {
					log.Printf("Ошибка добавления категории ID=%d к проекту ID=%d: %v", categoryId, project.ID, err)
				}
			}
		}
	}
	jsonOK(c, gin.H{"message": "Проект успешно обновлен"})
}

// DeleteProject - удаление проекта + всего связанного (изображения, категории, просмотры)
func (h *Handlers) DeleteProject(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var project models.Project
	if err := h.db.Preload("Images").Preload("Categories").First(&project, id).Error; err != nil {
		jsonErr(c, http.StatusNotFound, "Проект не найден")
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&project).Association("Categories").Clear(); err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления связей с категориями для проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления связей с категориями")
		return
	}

	if err := tx.Where("project_id = ?", project.ID).Delete(&models.Image{}).Error; err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления изображений для проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления изображений")
		return
	}

	if err := tx.Where("project_id = ?", project.ID).Delete(&models.ProjectViewDaily{}).Error; err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления статистики просмотров для проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления статистики просмотров")
		return
	}

	if err := tx.Delete(&project).Error; err != nil {
		tx.Rollback()
		log.Printf("Ошибка удаления проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка удаления проекта")
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Ошибка завершения транзакции при удалении проекта ID=%d: %v", id, err)
		jsonErr(c, http.StatusInternalServerError, "Ошибка завершения транзакции")
		return
	}

	for _, image := range project.Images {
		if err := deleteImageFile(image.FilePath); err != nil {
			logError("Ошибка удаления файла", image.FilePath, err)
		}
	}

	jsonOK(c, gin.H{"message": "Проект успешно удален"})
}
