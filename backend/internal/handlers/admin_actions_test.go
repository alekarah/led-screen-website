package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// UpdateContactStatus Tests
// ============================================================================

func TestUpdateContactStatus_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/status", h.UpdateContactStatus)

	contact := models.ContactForm{
		Name:   "Тест Контакт",
		Phone:  "+7 123 456-78-90",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"status": "processed",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/contacts/%d/status", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.Equal(t, "processed", updatedContact.Status)
	assert.Nil(t, updatedContact.ArchivedAt)
}

func TestUpdateContactStatus_ToArchived(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/status", h.UpdateContactStatus)

	contact := models.ContactForm{
		Name:   "Тест Контакт",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"status": "archived",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/contacts/%d/status", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedContact models.ContactForm
	h.db.First(&updatedContact, contact.ID)
	assert.Equal(t, "archived", updatedContact.Status)
	assert.NotNil(t, updatedContact.ArchivedAt)
}

func TestUpdateContactStatus_InvalidStatus(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/status", h.UpdateContactStatus)

	contact := models.ContactForm{
		Name:   "Тест",
		Phone:  "+7 123",
		Status: "new",
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"status": "invalid_status",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/contacts/%d/status", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateContactStatus_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/status", h.UpdateContactStatus)

	payload := map[string]interface{}{
		"status": "processed",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/invalid/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// BulkUpdateContacts Tests
// ============================================================================

func TestBulkUpdateContacts_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/bulk", h.BulkUpdateContacts)

	contact1 := models.ContactForm{Name: "Контакт 1", Phone: "+7 111", Status: "new"}
	contact2 := models.ContactForm{Name: "Контакт 2", Phone: "+7 222", Status: "new"}
	h.db.Create(&contact1)
	h.db.Create(&contact2)

	payload := map[string]interface{}{
		"action": "processed",
		"ids":    []uint{contact1.ID, contact2.ID},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated1, updated2 models.ContactForm
	h.db.First(&updated1, contact1.ID)
	h.db.First(&updated2, contact2.ID)
	assert.Equal(t, "processed", updated1.Status)
	assert.Equal(t, "processed", updated2.Status)
}

func TestBulkUpdateContacts_ToArchived(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/bulk", h.BulkUpdateContacts)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"action": "archived",
		"ids":    []uint{contact.ID},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "archived", updated.Status)
	assert.NotNil(t, updated.ArchivedAt)
}

func TestBulkUpdateContacts_EmptyIDs(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/bulk", h.BulkUpdateContacts)

	payload := map[string]interface{}{
		"action": "processed",
		"ids":    []uint{},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBulkUpdateContacts_InvalidAction(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/bulk", h.BulkUpdateContacts)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"action": "invalid",
		"ids":    []uint{contact.ID},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/bulk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// ArchiveContact Tests
// ============================================================================

func TestArchiveContact_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/archive", h.ArchiveContact)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/archive", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "archived", updated.Status)
	assert.NotNil(t, updated.ArchivedAt)
}

func TestArchiveContact_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/archive", h.ArchiveContact)

	req, _ := http.NewRequest("PATCH", "/admin/contacts/invalid/archive", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// RestoreContact Tests
// ============================================================================

func TestRestoreContact_ToNew(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/restore", h.RestoreContact)

	now := time.Now()
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "archived",
		ArchivedAt: &now,
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"to": "new",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/restore", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "new", updated.Status)
	assert.Nil(t, updated.ArchivedAt)
}

func TestRestoreContact_ToProcessed(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/restore", h.RestoreContact)

	now := time.Now()
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "archived",
		ArchivedAt: &now,
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"to": "processed",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/restore", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "processed", updated.Status)
	assert.Nil(t, updated.ArchivedAt)
}

func TestRestoreContact_DefaultToNew(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/restore", h.RestoreContact)

	now := time.Now()
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "archived",
		ArchivedAt: &now,
	}
	h.db.Create(&contact)

	// Не передаем "to" - должно быть "new" по умолчанию
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/restore", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "new", updated.Status)
	assert.Nil(t, updated.ArchivedAt)
}

// ============================================================================
// DeleteContact Tests
// ============================================================================

func TestDeleteContact_SoftDelete(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id", h.DeleteContact)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/contacts/%d", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Мягкое удаление - просто архивация
	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.Equal(t, "archived", updated.Status)
	assert.NotNil(t, updated.ArchivedAt)
}

func TestDeleteContact_HardDelete(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id", h.DeleteContact)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/contacts/%d?hard=true", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Жёсткое удаление - запись удалена из БД
	var count int64
	h.db.Model(&models.ContactForm{}).Where("id = ?", contact.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteContact_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id", h.DeleteContact)

	req, _ := http.NewRequest("DELETE", "/admin/contacts/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// GetContactNotes Tests
// ============================================================================

func TestGetContactNotes_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/contacts/:id/notes", h.GetContactNotes)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	note1 := models.ContactNote{ContactID: contact.ID, Text: "Заметка 1", Author: "Админ"}
	note2 := models.ContactNote{ContactID: contact.ID, Text: "Заметка 2", Author: "Админ"}
	h.db.Create(&note1)
	h.db.Create(&note2)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/contacts/%d/notes", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	notes := response["notes"].([]interface{})
	assert.Equal(t, 2, len(notes))
}

func TestGetContactNotes_EmptyList(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/contacts/:id/notes", h.GetContactNotes)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/contacts/%d/notes", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	notes := response["notes"].([]interface{})
	assert.Equal(t, 0, len(notes))
}

func TestGetContactNotes_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/contacts/:id/notes", h.GetContactNotes)

	req, _ := http.NewRequest("GET", "/admin/contacts/invalid/notes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// CreateContactNote Tests
// ============================================================================

func TestCreateContactNote_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/notes", h.CreateContactNote)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"text":   "Новая заметка",
		"author": "Тест Админ",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/contacts/%d/notes", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var note models.ContactNote
	h.db.Where("contact_id = ?", contact.ID).First(&note)
	assert.Equal(t, "Новая заметка", note.Text)
	assert.Equal(t, "Тест Админ", note.Author)
}

func TestCreateContactNote_MissingText(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/notes", h.CreateContactNote)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"text":   "",
		"author": "Админ",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/contacts/%d/notes", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateContactNote_InvalidContactID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/contacts/:id/notes", h.CreateContactNote)

	payload := map[string]interface{}{
		"text":   "Заметка",
		"author": "Админ",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/admin/contacts/invalid/notes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// DeleteContactNote Tests
// ============================================================================

func TestDeleteContactNote_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id/notes/:note_id", h.DeleteContactNote)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	note := models.ContactNote{ContactID: contact.ID, Text: "Заметка для удаления", Author: "Админ"}
	h.db.Create(&note)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/contacts/%d/notes/%d", contact.ID, note.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	h.db.Model(&models.ContactNote{}).Where("id = ?", note.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteContactNote_InvalidNoteID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id/notes/:note_id", h.DeleteContactNote)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/contacts/%d/notes/invalid", contact.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteContactNote_WrongContact(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/contacts/:id/notes/:note_id", h.DeleteContactNote)

	contact1 := models.ContactForm{Name: "Контакт 1", Phone: "+7 111", Status: "new"}
	contact2 := models.ContactForm{Name: "Контакт 2", Phone: "+7 222", Status: "new"}
	h.db.Create(&contact1)
	h.db.Create(&contact2)

	note := models.ContactNote{ContactID: contact1.ID, Text: "Заметка контакта 1", Author: "Админ"}
	h.db.Create(&note)

	// Пытаемся удалить заметку контакта 1 через endpoint контакта 2 (security test)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/contacts/%d/notes/%d", contact2.ID, note.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Заметка не должна быть удалена (принадлежит другому контакту)
	var count int64
	h.db.Model(&models.ContactNote{}).Where("id = ?", note.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// ============================================================================
// UpdateContactReminder Tests
// ============================================================================

func TestUpdateContactReminder_SetReminderRFC3339(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/reminder", h.UpdateContactReminder)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	futureTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	payload := map[string]interface{}{
		"remind_at": futureTime,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/reminder", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.True(t, updated.RemindFlag)
	assert.NotNil(t, updated.RemindAt)
}

func TestUpdateContactReminder_SetReminderMSKFormat(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/reminder", h.UpdateContactReminder)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"remind_at": "2025-12-31 15:00",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/reminder", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.True(t, updated.RemindFlag)
	assert.NotNil(t, updated.RemindAt)
}

func TestUpdateContactReminder_ClearReminder(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/reminder", h.UpdateContactReminder)

	now := time.Now().UTC()
	contact := models.ContactForm{
		Name:       "Тест",
		Phone:      "+7 123",
		Status:     "new",
		RemindFlag: true,
		RemindAt:   &now,
	}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"remind_at": "",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/reminder", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.ContactForm
	h.db.First(&updated, contact.ID)
	assert.False(t, updated.RemindFlag)
	assert.Nil(t, updated.RemindAt)
}

func TestUpdateContactReminder_InvalidFormat(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/reminder", h.UpdateContactReminder)

	contact := models.ContactForm{Name: "Тест", Phone: "+7 123", Status: "new"}
	h.db.Create(&contact)

	payload := map[string]interface{}{
		"remind_at": "invalid-date-format",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/admin/contacts/%d/reminder", contact.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateContactReminder_InvalidContactID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.PATCH("/admin/contacts/:id/reminder", h.UpdateContactReminder)

	payload := map[string]interface{}{
		"remind_at": "2025-12-31 15:00",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", "/admin/contacts/invalid/reminder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
