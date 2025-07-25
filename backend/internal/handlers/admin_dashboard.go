package handlers

import (
	"net/http"

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
	h.db.Preload("Categories").Preload("Images").Order("sort_order ASC, created_at DESC").Find(&projects)

	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":      "Управление проектами",
		"projects":   projects,
		"categories": categories,
	})
}
