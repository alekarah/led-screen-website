package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// formatPhone Tests
// ============================================================================

func TestFormatPhone_Standard11Digits(t *testing.T) {
	assert.Equal(t, "+7 967 560 88 58", formatPhone("+79675608858"))
}

func TestFormatPhone_WithDashes(t *testing.T) {
	assert.Equal(t, "+7 967 560 88 58", formatPhone("+7-967-560-88-58"))
}

func TestFormatPhone_WithSpaces(t *testing.T) {
	assert.Equal(t, "+7 967 560 88 58", formatPhone("+7 967 560 88 58"))
}

func TestFormatPhone_WithParentheses(t *testing.T) {
	assert.Equal(t, "+7 999 123 45 67", formatPhone("+7 (999) 123-45-67"))
}

func TestFormatPhone_Short_ReturnsOriginal(t *testing.T) {
	assert.Equal(t, "123", formatPhone("123"))
}

func TestFormatPhone_Empty_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", formatPhone(""))
}

func TestFormatPhone_WrongCountryCode_ReturnsOriginal(t *testing.T) {
	// 11 цифр но не начинается с 7
	assert.Equal(t, "+38099123456", formatPhone("+38099123456"))
}

func TestFormatPhone_TooLong_ReturnsOriginal(t *testing.T) {
	assert.Equal(t, "+799999999999", formatPhone("+799999999999")) // 12 цифр
}

// ============================================================================
// AdminSettingsUpdate Tests
// ============================================================================

func TestAdminSettingsUpdate_SavesAllFields(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.AutoMigrate(&models.SiteSettings{})
	router.POST("/admin/settings", h.AdminSettingsUpdate)

	form := url.Values{}
	form.Set("phone", "+79991234567")
	form.Set("phone_note", "Звонки с 9 до 21")
	form.Set("email", "test@example.com")
	form.Set("email_note", "Ответим за 2 часа")
	form.Set("address", "Санкт-Петербург")
	form.Set("address_note", "Выезд бесплатно")
	form.Set("work_hours", "Пн-Пт: 9-18")
	form.Set("work_hours_note", "Аварийные вызовы 24/7")
	form.Set("stats_projects", "200")
	form.Set("stats_years", "10")

	req, _ := http.NewRequest("POST", "/admin/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/settings?success=1", w.Header().Get("Location"))

	var s models.SiteSettings
	h.db.First(&s)
	assert.Equal(t, "+79991234567", s.Phone)
	assert.Equal(t, "+7 999 123 45 67", s.PhoneDisplay) // formatPhone применён
	assert.Equal(t, "Звонки с 9 до 21", s.PhoneNote)
	assert.Equal(t, "test@example.com", s.Email)
	assert.Equal(t, "Ответим за 2 часа", s.EmailNote)
	assert.Equal(t, "Санкт-Петербург", s.Address)
	assert.Equal(t, "Выезд бесплатно", s.AddressNote)
	assert.Equal(t, "Пн-Пт: 9-18", s.WorkHours)
	assert.Equal(t, "Аварийные вызовы 24/7", s.WorkHoursNote)
	assert.Equal(t, 200, s.StatsProjects)
	assert.Equal(t, 10, s.StatsYears)
}

func TestAdminSettingsUpdate_InvalidStats_KeepsOldValue(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.AutoMigrate(&models.SiteSettings{})

	// Создаём начальные настройки
	h.db.Create(&models.SiteSettings{
		Phone:         "+79991234567",
		StatsProjects: 150,
		StatsYears:    5,
	})

	router.POST("/admin/settings", h.AdminSettingsUpdate)

	form := url.Values{}
	form.Set("phone", "+79991234567")
	form.Set("stats_projects", "не_число") // невалидное значение
	form.Set("stats_years", "abc")         // невалидное значение

	req, _ := http.NewRequest("POST", "/admin/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var s models.SiteSettings
	h.db.First(&s)
	// Невалидные числа — значения не меняются
	assert.Equal(t, 150, s.StatsProjects)
	assert.Equal(t, 5, s.StatsYears)
}

func TestAdminSettingsUpdate_PhoneDisplayFormatted(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.AutoMigrate(&models.SiteSettings{})
	router.POST("/admin/settings", h.AdminSettingsUpdate)

	form := url.Values{}
	form.Set("phone", "89001234567") // 8 вместо +7
	req, _ := http.NewRequest("POST", "/admin/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var s models.SiteSettings
	h.db.First(&s)
	// 8... → первая цифра не 7, форматирование не применяется
	assert.Equal(t, "89001234567", s.PhoneDisplay)
}

func TestAdminSettingsUpdate_ZeroStats(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.AutoMigrate(&models.SiteSettings{})
	router.POST("/admin/settings", h.AdminSettingsUpdate)

	form := url.Values{}
	form.Set("stats_projects", "0")
	form.Set("stats_years", "0")

	req, _ := http.NewRequest("POST", "/admin/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var s models.SiteSettings
	h.db.First(&s)
	assert.Equal(t, 0, s.StatsProjects)
	assert.Equal(t, 0, s.StatsYears)
}
