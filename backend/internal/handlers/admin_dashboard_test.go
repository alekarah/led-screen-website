package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestAdminContacts7Days_EmptyDatabase проверяет возврат пустой статистики
// ПРИМЕЧАНИЕ: Тест пропускается для SQLite, так как используется PostgreSQL-специфичный SQL
func TestAdminContacts7Days_EmptyDatabase(t *testing.T) {
	t.Skip("Пропуск: AdminContacts7Days использует PostgreSQL-специфичный SQL (generate_series, INTERVAL)")

	router, h := setupTestRouter(t)
	router.GET("/api/admin/contacts-7d", h.AdminContacts7Days)

	req, _ := http.NewRequest("GET", "/api/admin/contacts-7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Должно вернуться 7 дней (с сегодня по -6 дней)
	assert.Equal(t, 7, len(response))

	// Все счётчики должны быть 0
	for _, day := range response {
		count := int(day["count"].(float64))
		assert.Equal(t, 0, count)
	}
}

// TestAdminContacts7Days_WithData проверяет подсчёт заявок за последние 7 дней
// ПРИМЕЧАНИЕ: Тест пропускается для SQLite, так как используется PostgreSQL-специфичный SQL
func TestAdminContacts7Days_WithData(t *testing.T) {
	t.Skip("Пропуск: AdminContacts7Days использует PostgreSQL-специфичный SQL (generate_series, INTERVAL)")

	router, h := setupTestRouter(t)
	router.GET("/api/admin/contacts-7d", h.AdminContacts7Days)

	// Создаем заявки за разные дни
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)

	// 3 заявки сегодня
	for i := 0; i < 3; i++ {
		contact := models.ContactForm{
			Name:      "Контакт сегодня",
			Phone:     "+7 123",
			CreatedAt: today,
		}
		h.db.Create(&contact)
	}

	// 2 заявки вчера
	yesterday := today.AddDate(0, 0, -1)
	for i := 0; i < 2; i++ {
		contact := models.ContactForm{
			Name:      "Контакт вчера",
			Phone:     "+7 123",
			CreatedAt: yesterday,
		}
		h.db.Create(&contact)
	}

	// 1 заявка 3 дня назад
	threeDaysAgo := today.AddDate(0, 0, -3)
	contact := models.ContactForm{
		Name:      "Контакт 3 дня назад",
		Phone:     "+7 123",
		CreatedAt: threeDaysAgo,
	}
	h.db.Create(&contact)

	// 1 заявка 10 дней назад (не должна попасть в выборку)
	tenDaysAgo := today.AddDate(0, 0, -10)
	oldContact := models.ContactForm{
		Name:      "Старый контакт",
		Phone:     "+7 123",
		CreatedAt: tenDaysAgo,
	}
	h.db.Create(&oldContact)

	req, _ := http.NewRequest("GET", "/api/admin/contacts-7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Должно быть 7 дней
	assert.Equal(t, 7, len(response))

	// Суммарно должно быть 6 заявок (3 + 2 + 1, без старой заявки 10 дней назад)
	totalCount := 0
	for _, day := range response {
		count := int(day["count"].(float64))
		totalCount += count
	}
	assert.Equal(t, 6, totalCount)
}
