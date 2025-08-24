package handlers

import (
	"net/http"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// /admin/contacts — страница всех заявок
func (h *Handlers) AdminContactsPage(c *gin.Context) {
	var contacts []models.ContactForm
	h.db.Order("created_at DESC").Find(&contacts)

	c.HTML(http.StatusOK, "admin_base.html", gin.H{
		"title":       "Заявки",
		"PageID":      "admin-contacts",
		"contactsAll": contacts,
	})
}

// Пометить заявку как обработанную
func (h *Handlers) MarkContactDone(c *gin.Context) {
	id := c.Param("id")

	// обновляем поле Processed = true
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Update("status", "processed").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заявку"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заявка помечена как обработанная"})
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
