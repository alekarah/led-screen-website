package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestTelegramUpdateStatus_Success проверяет успешное изменение статуса
func TestTelegramUpdateStatus_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/update-status", h.TelegramUpdateStatus)

	// Создаем тестовый контакт
	contact := models.ContactForm{
		Name:   "Тест Контакт",
		Phone:  "+7 123 456-78-90",
		Status: "new",
	}
	h.db.Create(&contact)

	// Отправляем запрос на изменение статуса
	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"status":     "processed",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/update-status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что статус изменился в БД
	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.Equal(t, "processed", updatedContact.Status)
}

// TestTelegramUpdateStatus_InvalidStatus проверяет обработку неверного статуса
func TestTelegramUpdateStatus_InvalidStatus(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/update-status", h.TelegramUpdateStatus)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"status":     "invalid_status",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/update-status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Недопустимый статус")
}

// TestTelegramUpdateStatus_ArchiveContact проверяет архивирование контакта
func TestTelegramUpdateStatus_ArchiveContact(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/update-status", h.TelegramUpdateStatus)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "processed",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"status":     "archived",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/update-status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что archived_at установлен
	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.Equal(t, "archived", updatedContact.Status)
	assert.NotNil(t, updatedContact.ArchivedAt)
}

// TestTelegramAddNote_Success проверяет успешное добавление заметки
func TestTelegramAddNote_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/add-note", h.TelegramAddNote)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"text":       "Тестовая заметка",
		"author":     "Telegram Bot",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/add-note", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что заметка создалась
	var note models.ContactNote
	h.db.Where("contact_id = ?", contact.ID).First(&note)
	assert.Equal(t, "Тестовая заметка", note.Text)
	assert.Equal(t, "Telegram Bot", note.Author)
}

// TestTelegramAddNote_MissingFields проверяет обработку отсутствующих полей
func TestTelegramAddNote_MissingFields(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/add-note", h.TelegramAddNote)

	payload := map[string]interface{}{
		"contact_id": 999,
		// text отсутствует
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/add-note", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTelegramSetReminder_Success проверяет успешную установку напоминания
func TestTelegramSetReminder_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/set-reminder", h.TelegramSetReminder)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	// Устанавливаем напоминание на завтра 9:00
	tomorrow := time.Now().Add(24*time.Hour).Format("2006-01-02") + " 09:00"
	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"remind_at":  tomorrow,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/set-reminder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что remind_flag установлен
	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.True(t, updatedContact.RemindFlag)
	assert.NotNil(t, updatedContact.RemindAt)
}

// TestTelegramSetReminder_InvalidFormat проверяет обработку неверного формата даты
func TestTelegramSetReminder_InvalidFormat(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/set-reminder", h.TelegramSetReminder)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"contact_id": contact.ID,
		"remind_at":  "invalid-date-format",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/set-reminder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTelegramGetDueReminders_NoReminders проверяет когда нет напоминаний
func TestTelegramGetDueReminders_NoReminders(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/telegram/due-reminders", h.TelegramGetDueReminders)

	req, _ := http.NewRequest("GET", "/api/telegram/due-reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	reminders := response["reminders"].([]interface{})
	assert.Equal(t, 0, len(reminders))
}

// TestTelegramGetDueReminders_WithDueReminder проверяет возврат напоминаний
func TestTelegramGetDueReminders_WithDueReminder(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/telegram/due-reminders", h.TelegramGetDueReminders)

	// Создаем контакт с напоминанием в прошлом (используем UTC как в БД)
	pastTime := time.Now().UTC().Add(-1 * time.Hour)
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Email:      "test@example.com",
		Status:     "new",
		RemindFlag: true,
		RemindAt:   &pastTime,
	}
	h.db.Create(&contact)

	req, _ := http.NewRequest("GET", "/api/telegram/due-reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	reminders := response["reminders"].([]interface{})

	// Проверяем что есть хотя бы одно напоминание
	if assert.Greater(t, len(reminders), 0, "Должно быть хотя бы одно напоминание") {
		reminder := reminders[0].(map[string]interface{})
		assert.Equal(t, "Тест", reminder["name"])
		assert.Equal(t, "+7 123", reminder["phone"])
		assert.Equal(t, "test@example.com", reminder["email"])
	}
}

// TestTelegramGetDueReminders_OnlyFutureReminders проверяет что будущие напоминания не возвращаются
func TestTelegramGetDueReminders_OnlyFutureReminders(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/telegram/due-reminders", h.TelegramGetDueReminders)

	// Создаем контакт с напоминанием в будущем
	futureTime := time.Now().Add(24 * time.Hour)
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "new",
		RemindFlag: true,
		RemindAt:   &futureTime,
	}
	h.db.Create(&contact)

	req, _ := http.NewRequest("GET", "/api/telegram/due-reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	reminders := response["reminders"].([]interface{})
	// Напоминание в будущем не должно возвращаться
	assert.Equal(t, 0, len(reminders))
}

// TestTelegramMarkReminderSent_Success проверяет успешную пометку напоминания
func TestTelegramMarkReminderSent_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/mark-reminder-sent", h.TelegramMarkReminderSent)

	pastTime := time.Now().Add(-1 * time.Hour)
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "new",
		RemindFlag: true,
		RemindAt:   &pastTime,
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"contact_id": contact.ID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/mark-reminder-sent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что remind_flag сброшен
	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.False(t, updatedContact.RemindFlag)
}

// TestTelegramMarkReminderSent_MissingContactID проверяет обработку отсутствующего contact_id
func TestTelegramMarkReminderSent_MissingContactID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/telegram/mark-reminder-sent", h.TelegramMarkReminderSent)

	payload := map[string]interface{}{
		// contact_id отсутствует
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/telegram/mark-reminder-sent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
