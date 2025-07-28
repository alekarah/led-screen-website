package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ledsite/internal/models"
)

type Handlers struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handlers {
	return &Handlers{db: db}
}

// HomePage - главная страница
func (h *Handlers) HomePage(c *gin.Context) {
	// Получаем проекты для главной страницы (только те, что помечены для показа)
	var featuredProjects []models.Project
	h.db.Where("featured = ?", true).
		Preload("Categories").
		Preload("Images").
		Order("sort_order ASC, created_at DESC").
		Limit(3).
		Find(&featuredProjects)

	// Если нет рекомендуемых проектов, показываем последние
	if len(featuredProjects) == 0 {
		h.db.Preload("Categories").
			Preload("Images").
			Order("sort_order ASC, created_at DESC").
			Limit(3).
			Find(&featuredProjects)
	}

	// Получаем основные услуги
	var services []models.Service
	h.db.Where("featured = ?", true).
		Order("sort_order").
		Find(&services)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":    "LED экраны в СПб | Service 'n' Repair",
		"projects": featuredProjects,
		"services": services,
	})
}

// ProjectsPage отображает страницу портфолио
func (h *Handlers) ProjectsPage(c *gin.Context) {
	// Получаем параметр фильтрации
	categorySlug := c.Query("category")

	query := h.db.Preload("Categories").Preload("Images")

	if categorySlug != "" {
		query = query.Joins("JOIN project_categories pc ON projects.id = pc.project_id").
			Joins("JOIN categories cat ON pc.category_id = cat.id").
			Where("cat.slug = ?", categorySlug)
	}

	var projects []models.Project
	query.Order("sort_order ASC, created_at DESC").Find(&projects)

	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":          "Портфолио - LED Display",
		"projects":       projects,
		"categories":     categories,
		"activeCategory": categorySlug,
	})
}

// ServicesPage - страница услуг
func (h *Handlers) ServicesPage(c *gin.Context) {
	var services []models.Service
	h.db.Order("sort_order").Find(&services)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":    "Услуги | LED экраны",
		"services": services,
	})
}

// ContactPage - страница контактов
func (h *Handlers) ContactPage(c *gin.Context) {
	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title": "Контакты | LED экраны",
	})
}

// API Handlers

// GetProjects - API для получения проектов
func (h *Handlers) GetProjects(c *gin.Context) {
	var projects []models.Project

	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	offset := (page - 1) * limit

	// Фильтрация по категории
	category := c.Query("category")

	query := h.db.Preload("Categories").Preload("Images")

	if category != "" {
		query = query.Joins("JOIN project_categories pc ON projects.id = pc.project_id").
			Joins("JOIN categories cat ON pc.category_id = cat.id").
			Where("cat.slug = ?", category)
	}

	var total int64
	query.Model(&models.Project{}).Count(&total)

	query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects)

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// SubmitContact - обработка формы обратной связи
func (h *Handlers) SubmitContact(c *gin.Context) {
	var form models.ContactForm

	// Проверяем Content-Type и парсим данные соответственно
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверные данные формы",
			})
			return
		}
	} else {
		// Парсим данные формы
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверные данные формы",
			})
			return
		}
	}

	// Простая валидация
	if form.Name == "" || form.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Имя и телефон обязательны для заполнения",
		})
		return
	}

	// Сохраняем в базу
	if err := h.db.Create(&form).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения заявки",
		})
		return
	}

	// TODO: Отправка email уведомления

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно отправлена! Мы свяжемся с вами в ближайшее время.",
	})
}

// PrivacyPage - страница обработки персональных данных
func (h *Handlers) PrivacyPage(c *gin.Context) {
	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title": "Обработка персональных данных",
	})
}

// Утилитные функции
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "ё", "e")
	slug = strings.ReplaceAll(slug, "а", "a")
	slug = strings.ReplaceAll(slug, "б", "b")
	slug = strings.ReplaceAll(slug, "в", "v")
	slug = strings.ReplaceAll(slug, "г", "g")
	slug = strings.ReplaceAll(slug, "д", "d")
	slug = strings.ReplaceAll(slug, "е", "e")
	slug = strings.ReplaceAll(slug, "ж", "zh")
	slug = strings.ReplaceAll(slug, "з", "z")
	slug = strings.ReplaceAll(slug, "и", "i")
	slug = strings.ReplaceAll(slug, "й", "y")
	slug = strings.ReplaceAll(slug, "к", "k")
	slug = strings.ReplaceAll(slug, "л", "l")
	slug = strings.ReplaceAll(slug, "м", "m")
	slug = strings.ReplaceAll(slug, "н", "n")
	slug = strings.ReplaceAll(slug, "о", "o")
	slug = strings.ReplaceAll(slug, "п", "p")
	slug = strings.ReplaceAll(slug, "р", "r")
	slug = strings.ReplaceAll(slug, "с", "s")
	slug = strings.ReplaceAll(slug, "т", "t")
	slug = strings.ReplaceAll(slug, "у", "u")
	slug = strings.ReplaceAll(slug, "ф", "f")
	slug = strings.ReplaceAll(slug, "х", "h")
	slug = strings.ReplaceAll(slug, "ц", "ts")
	slug = strings.ReplaceAll(slug, "ч", "ch")
	slug = strings.ReplaceAll(slug, "ш", "sh")
	slug = strings.ReplaceAll(slug, "щ", "sch")
	slug = strings.ReplaceAll(slug, "ъ", "")
	slug = strings.ReplaceAll(slug, "ы", "y")
	slug = strings.ReplaceAll(slug, "ь", "")
	slug = strings.ReplaceAll(slug, "э", "e")
	slug = strings.ReplaceAll(slug, "ю", "yu")
	slug = strings.ReplaceAll(slug, "я", "ya")

	// Добавляем timestamp для уникальности
	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}
