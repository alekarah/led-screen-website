package handlers

import (
	"net/http"
	"strconv"
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

	now := NowMSKUTC()
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

	now := NowMSKUTC()
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

	now := NowMSKUTC()
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
	// Игнорируем ошибку binding - используем значение по умолчанию если не передано
	if err := c.ShouldBindJSON(&body); err != nil {
		body.To = "new" // default value
	}
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
	now := NowMSKUTC()
	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": "archived", "archived_at": &now}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось архивировать")
		return
	}
	jsonOK(c, gin.H{"message": "В архиве", "status": "archived"})
}

// Получить список заметок по заявке
func (h *Handlers) GetContactNotes(c *gin.Context) {
	contactID, ok := mustID(c)
	if !ok {
		return
	}

	var notes []models.ContactNote
	if err := h.db.Where("contact_id = ?", contactID).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось получить заметки")
		return
	}
	jsonOK(c, gin.H{"notes": notes})
}

// Создать заметку
func (h *Handlers) CreateContactNote(c *gin.Context) {
	contactID, ok := mustID(c)
	if !ok {
		return
	}

	var body struct {
		Text   string `json:"text"`
		Author string `json:"author"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Text) == 0 {
		jsonErr(c, http.StatusBadRequest, "Текст заметки обязателен")
		return
	}

	note := models.ContactNote{
		ContactID: contactID,
		Text:      body.Text,
		Author:    body.Author,
		CreatedAt: NowMSKUTC(),
	}
	if err := h.db.Create(&note).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось создать заметку")
		return
	}
	jsonOK(c, gin.H{"note": note})
}

// Удалить заметку
func (h *Handlers) DeleteContactNote(c *gin.Context) {
	contactID, ok := mustID(c)
	if !ok {
		return
	}

	noteIDStr := c.Param("note_id")
	noteID64, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil || noteID64 == 0 {
		jsonErr(c, http.StatusBadRequest, "Некорректный id заметки")
		return
	}
	noteID := uint(noteID64)

	// Удаляем только если заметка принадлежит этой заявке
	if err := h.db.Where("id = ? AND contact_id = ?", noteID, contactID).
		Delete(&models.ContactNote{}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось удалить заметку")
		return
	}
	jsonOK(c, gin.H{"message": "Заметка удалена", "id": noteID})
}

// Установить/снять напоминание
func (h *Handlers) UpdateContactReminder(c *gin.Context) {
	contactID, ok := mustID(c)
	if !ok {
		return
	}

	var body struct {
		RemindAt   string `json:"remind_at"`   // RFC3339 или "2006-01-02 15:04"
		RemindFlag *bool  `json:"remind_flag"` // если не передали — вычислим сами
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}

	var remindAtPtr *time.Time
	var flag bool

	if body.RemindAt == "" {
		// очистка напоминания
		flag = false
	} else {
		// пробуем RFC3339, затем "2006-01-02 15:04" в МСК
		if t, err := time.Parse(time.RFC3339, body.RemindAt); err == nil {
			tt := t.UTC()
			remindAtPtr = &tt
		} else if t2, err2 := time.ParseInLocation("2006-01-02 15:04", body.RemindAt, moscowLoc); err2 == nil {
			tt := t2.UTC()
			remindAtPtr = &tt
		} else {
			jsonErr(c, http.StatusBadRequest, "Неверный формат даты напоминания")
			return
		}
		flag = true
	}

	// если явно передали RemindFlag — уважим
	if body.RemindFlag != nil {
		flag = *body.RemindFlag && remindAtPtr != nil
	}

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", contactID).
		Updates(map[string]any{
			"remind_at":   remindAtPtr,
			"remind_flag": flag,
		}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось сохранить напоминание")
		return
	}

	jsonOK(c, gin.H{
		"message":     "Напоминание сохранено",
		"remind_at":   remindAtPtr,
		"remind_flag": flag,
	})
}
