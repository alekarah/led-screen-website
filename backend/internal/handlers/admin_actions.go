package handlers

import (
	"net/http"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// Смена статуса заявки (new | processed | archived)
func (h *Handlers) UpdateContactStatus(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}
	status, valid := parseStatus(body.Status)
	if !valid {
		jsonErr(c, http.StatusBadRequest, "Недопустимый статус")
		return
	}

	now := NowMSK().UTC()
	var archivedAt *time.Time
	if status == "archived" {
		archivedAt = &now
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "archived_at": archivedAt}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось обновить заявку")
		return
	}
	jsonOK(c, gin.H{"message": "Статус изменён", "status": status})
}

// Массовая смена статуса
func (h *Handlers) BulkUpdateContacts(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		IDs    []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}
	action, valid := parseStatus(req.Action)
	if !valid {
		jsonErr(c, http.StatusBadRequest, "Недопустимое действие")
		return
	}

	now := NowMSK().UTC()
	var archivedAt *time.Time
	if action == "archived" {
		archivedAt = &now
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]any{"status": action, "archived_at": archivedAt}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось обновить заявки")
		return
	}
	jsonOK(c, gin.H{"success": true, "action": action, "ids": req.IDs})
}

// Архивировать одну заявку
func (h *Handlers) ArchiveContact(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	now := NowMSK().UTC()
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "archived", "archived_at": &now}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось архивировать")
		return
	}
	jsonOK(c, gin.H{"message": "В архиве", "status": "archived"})
}

// Восстановить из архива
func (h *Handlers) RestoreContact(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	var body struct {
		To string `json:"to"`
	} // допустимо: "new" | "processed"
	_ = c.ShouldBindJSON(&body)
	if body.To != "processed" {
		body.To = "new"
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": body.To, "archived_at": nil}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось восстановить")
		return
	}
	jsonOK(c, gin.H{"message": "Восстановлено", "status": body.To})
}

// Удалить заявку (hard=true — жёстко)
func (h *Handlers) DeleteContact(c *gin.Context) {
	id, ok := mustID(c)
	if !ok {
		return
	}

	hard := c.Query("hard")
	if hard == "true" || hard == "1" || hard == "yes" {
		if err := h.db.Delete(&models.ContactForm{}, id).Error; err != nil {
			jsonErr(c, http.StatusInternalServerError, "Не удалось удалить")
			return
		}
		jsonOK(c, gin.H{"message": "Удалено"})
		return
	}
	// мягкий архив
	now := NowMSK().UTC()
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "archived", "archived_at": &now}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось архивировать")
		return
	}
	jsonOK(c, gin.H{"message": "В архиве", "status": "archived"})
}
