package handlers

import (
	"net/http"
	"os"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
)

// TelegramUpdateStatusRequest - запрос на изменение статуса из Telegram бота
type TelegramUpdateStatusRequest struct {
	ContactID uint   `json:"contact_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

// TelegramAddNoteRequest - запрос на добавление заметки из Telegram бота
type TelegramAddNoteRequest struct {
	ContactID uint   `json:"contact_id" binding:"required"`
	Text      string `json:"text" binding:"required"`
	Author    string `json:"author" binding:"required"`
}

// TelegramSetReminderRequest - запрос на установку напоминания из Telegram бота
type TelegramSetReminderRequest struct {
	ContactID uint   `json:"contact_id" binding:"required"`
	RemindAt  string `json:"remind_at" binding:"required"` // формат: "2006-01-02 15:04"
}

// telegramAuthMiddleware проверяет что запрос пришел от Telegram бота (localhost)
// В production можно добавить проверку секретного токена
func telegramAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем что запрос с localhost
		clientIP := c.ClientIP()
		if clientIP != "127.0.0.1" && clientIP != "::1" {
			jsonErr(c, http.StatusForbidden, "Доступ запрещен")
			c.Abort()
			return
		}

		// Опционально: проверка секретного токена
		expectedToken := os.Getenv("TELEGRAM_API_SECRET")
		if expectedToken != "" {
			token := c.GetHeader("X-Telegram-Token")
			if token != expectedToken {
				jsonErr(c, http.StatusUnauthorized, "Неверный токен")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// TelegramUpdateStatus обновляет статус контакта по запросу от Telegram бота
func (h *Handlers) TelegramUpdateStatus(c *gin.Context) {
	var req TelegramUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}

	status, valid := parseStatus(req.Status)
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
		Where("id = ?", req.ContactID).
		Updates(map[string]any{"status": status, "archived_at": archivedAt}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось обновить заявку")
		return
	}

	jsonOK(c, gin.H{"message": "Статус изменён", "status": status})
}

// TelegramAddNote добавляет заметку к контакту по запросу от Telegram бота
func (h *Handlers) TelegramAddNote(c *gin.Context) {
	var req TelegramAddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}

	note := models.ContactNote{
		ContactID: req.ContactID,
		Text:      req.Text,
		Author:    req.Author,
		CreatedAt: NowMSKUTC(),
	}

	if err := h.db.Create(&note).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось создать заметку")
		return
	}

	jsonOK(c, gin.H{"note": note})
}

// TelegramSetReminder устанавливает напоминание по запросу от Telegram бота
func (h *Handlers) TelegramSetReminder(c *gin.Context) {
	var req TelegramSetReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверные данные")
		return
	}

	// Парсим дату напоминания в формате "2006-01-02 15:04" (МСК)
	t, err := time.ParseInLocation("2006-01-02 15:04", req.RemindAt, moscowLoc)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, "Неверный формат даты напоминания")
		return
	}

	remindAtUTC := t.UTC()

	if err := h.db.Model(&models.ContactForm{}).
		Where("id = ?", req.ContactID).
		Updates(map[string]any{
			"remind_at":   &remindAtUTC,
			"remind_flag": true,
		}).Error; err != nil {
		jsonErr(c, http.StatusInternalServerError, "Не удалось сохранить напоминание")
		return
	}

	jsonOK(c, gin.H{
		"message":     "Напоминание сохранено",
		"remind_at":   remindAtUTC,
		"remind_flag": true,
	})
}
