package handlers

import (
	"net/http"
	"strconv"

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
	// Получаем последние проекты для главной страницы
	var featuredProjects []models.Project
	h.db.Where("featured = ?", true).
		Preload("Categories").
		Preload("Images").
		Limit(6).
		Find(&featuredProjects)

	// Получаем основные услуги
	var services []models.Service
	h.db.Where("featured = ?", true).
		Order("sort_order").
		Find(&services)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":    "LED экраны в СПб | Service 'n' Repair",
		"projects": featuredProjects,
		"services": services,
	})
}

// ProjectsPage - страница портфолио
func (h *Handlers) ProjectsPage(c *gin.Context) {
	// Получаем параметры фильтрации
	categorySlug := c.Query("category")

	query := h.db.Preload("Categories").Preload("Images")

	if categorySlug != "" {
		query = query.Joins("JOIN project_categories pc ON projects.id = pc.project_id").
			Joins("JOIN categories cat ON pc.category_id = cat.id").
			Where("cat.slug = ?", categorySlug)
	}

	var projects []models.Project
	query.Order("created_at DESC").Find(&projects)

	// Получаем все категории для фильтров
	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "projects.html", gin.H{
		"title":          "Портфолио | LED экраны",
		"projects":       projects,
		"categories":     categories,
		"activeCategory": categorySlug,
	})
}

// ProjectDetail - детальная страница проекта
func (h *Handlers) ProjectDetail(c *gin.Context) {
	slug := c.Param("slug")

	var project models.Project
	if err := h.db.Where("slug = ?", slug).
		Preload("Categories").
		Preload("Images").
		First(&project).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"title": "Проект не найден",
		})
		return
	}

	// Увеличиваем счетчик просмотров
	h.db.Model(&project).Update("view_count", project.ViewCount+1)

	c.HTML(http.StatusOK, "project_detail.html", gin.H{
		"title":   project.Title + " | Портфолио",
		"project": project,
	})
}

// ServicesPage - страница услуг
func (h *Handlers) ServicesPage(c *gin.Context) {
	var services []models.Service
	h.db.Order("sort_order").Find(&services)

	c.HTML(http.StatusOK, "services.html", gin.H{
		"title":    "Услуги | LED экраны",
		"services": services,
	})
}

// ContactPage - страница контактов
func (h *Handlers) ContactPage(c *gin.Context) {
	c.HTML(http.StatusOK, "contact.html", gin.H{
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

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные формы",
		})
		return
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
