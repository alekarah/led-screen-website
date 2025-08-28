package handlers

import (
	"net/http"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// Смена статуса заявки (new | processed | archived)
func (h *Handlers) UpdateContactStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if body.Status != "new" && body.Status != "processed" && body.Status != "archived" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый статус"})
		return
	}
	now := nowMoscow().UTC()
	var archivedAt *time.Time
	if body.Status == "archived" {
		archivedAt = &now
	} else {
		archivedAt = nil
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": body.Status, "archived_at": archivedAt}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заявку"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Статус изменён", "status": body.Status})
}

// Массовая смена статуса
func (h *Handlers) BulkUpdateContacts(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		IDs    []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Неверные данные"})
		return
	}
	switch req.Action {
	case "processed", "new", "archived":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Недопустимое действие"})
		return
	}
	now := nowMoscow().UTC()
	var archivedAt *time.Time
	if req.Action == "archived" {
		archivedAt = &now
	} else {
		archivedAt = nil
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{"status": req.Action, "archived_at": archivedAt}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Не удалось обновить заявки"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "action": req.Action, "ids": req.IDs})
}

// Архивировать одну заявку
func (h *Handlers) ArchiveContact(c *gin.Context) {
	id := c.Param("id")
	now := nowMoscow().UTC()
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": "archived", "archived_at": &now}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось архивировать"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "В архиве", "status": "archived"})
}

// Восстановить из архива
func (h *Handlers) RestoreContact(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		To string `json:"to"`
	} // "new" | "processed"
	_ = c.ShouldBindJSON(&body)
	if body.To != "processed" {
		body.To = "new"
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": body.To, "archived_at": nil}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось восстановить"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Восстановлено", "status": body.To})
}

// Удалить заявку (hard=true — жёстко)
func (h *Handlers) DeleteContact(c *gin.Context) {
	id := c.Param("id")
	hard := c.Query("hard")
	if hard == "true" || hard == "1" || hard == "yes" {
		if err := h.db.Delete(&models.ContactForm{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
		return
	}
	// fallback: мягкий архив
	now := nowMoscow().UTC()
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": "archived", "archived_at": &now}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось архивировать"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "В архиве", "status": "archived"})
}
