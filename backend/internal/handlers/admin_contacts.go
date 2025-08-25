package handlers

import (
	"net/http"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// /admin/contacts — страница всех заявок
func (h *Handlers) AdminContactsPage(c *gin.Context) {
	var contacts []models.ContactForm
	query := h.db.Model(&models.ContactForm{})

	// --- Поиск ---
	if search := c.Query("search"); search != "" {
		q := "%" + search + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR message ILIKE ?", q, q, q, q)
	}

	// --- Фильтр по статусу ---
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// --- Интервал дат ---
	switch c.Query("date") {
	case "today":
		query = query.Where("DATE(created_at) = CURRENT_DATE")
	case "7d":
		query = query.Where("created_at >= NOW() - INTERVAL '7 days'")
	case "month":
		query = query.Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE)")
	}

	query.Order("created_at DESC").Find(&contacts)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Заявки",
		"PageID":      "admin-contacts",
		"contactsAll": contacts,
		"search":      c.Query("search"),
		"status":      c.Query("status"),
		"dateRange":   c.Query("date"),
	})
}

// Смена статуса заявки
func (h *Handlers) UpdateContactStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Проверка допустимых статусов
	if body.Status != "new" && body.Status != "processed" && body.Status != "archived" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый статус"})
		return
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Update("status", body.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заявку"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус изменён", "status": body.Status})
}
