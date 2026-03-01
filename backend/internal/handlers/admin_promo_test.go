package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// GetActivePromo Tests
// ============================================================================

func TestGetActivePromo_NoPromoExists(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	req, _ := http.NewRequest("GET", "/api/promo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["active"].(bool))
}

func TestGetActivePromo_InactivePromo(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	// Создаем неактивное промо
	promo := models.PromoPopup{
		Title:    "Скидка 20%",
		Content:  "Успейте заказать со скидкой!",
		Pages:    `["home","projects"]`,
		IsActive: false,
	}
	h.db.Create(&promo)

	req, _ := http.NewRequest("GET", "/api/promo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["active"].(bool))
}

func TestGetActivePromo_ActiveOnHomePage(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	// Создаем активное промо для главной страницы
	promo := models.PromoPopup{
		Title:    "Скидка 20%",
		Content:  "Успейте заказать со скидкой!",
		Pages:    `["home","projects"]`,
		IsActive: true,
	}
	h.db.Create(&promo)

	req, _ := http.NewRequest("GET", "/api/promo?page=home", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["active"].(bool))
	assert.Equal(t, "Скидка 20%", response["title"])
	assert.Equal(t, "Успейте заказать со скидкой!", response["content"])
}

func TestGetActivePromo_DefaultPageIsHome(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	// Создаем активное промо для главной страницы
	promo := models.PromoPopup{
		Title:    "Акция",
		Content:  "Специальное предложение",
		Pages:    `["home"]`,
		IsActive: true,
	}
	h.db.Create(&promo)

	// Запрос без параметра page (по умолчанию home)
	req, _ := http.NewRequest("GET", "/api/promo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["active"].(bool))
}

func TestGetActivePromo_PageNotEnabled(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	// Создаем активное промо только для главной страницы
	promo := models.PromoPopup{
		Title:    "Скидка 20%",
		Content:  "Успейте заказать!",
		Pages:    `["home"]`,
		IsActive: true,
	}
	h.db.Create(&promo)

	// Запрашиваем для страницы projects (не включена)
	req, _ := http.NewRequest("GET", "/api/promo?page=projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["active"].(bool))
}

func TestGetActivePromo_InvalidPagesJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/promo", h.GetActivePromo)

	// Создаем промо с невалидным JSON в Pages
	promo := models.PromoPopup{
		Title:    "Скидка 20%",
		Content:  "Успейте заказать!",
		Pages:    `invalid json`,
		IsActive: true,
	}
	h.db.Create(&promo)

	req, _ := http.NewRequest("GET", "/api/promo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["active"].(bool))
}

// ============================================================================
// AdminPromoUpdate Tests
// ============================================================================

func TestAdminPromoUpdate_SavesAllFields(t *testing.T) {
	router, h := setupTestRouter(t)
	// Создаём начальную запись (AdminPromoUpdate требует существующей)
	h.db.Create(&models.PromoPopup{Title: "old", Content: "old", IsActive: false, Pages: `["home"]`, TTLHours: 24})
	router.POST("/admin/promo", h.AdminPromoUpdate)

	form := url.Values{}
	form.Set("title", "Летняя акция")
	form.Set("content", "Скидка 15% до конца июля")
	form.Set("is_active", "on")
	form.Set("ttl_hours", "48")
	form.Set("show_delay_seconds", "5")
	form.Add("pages[]", "home")
	form.Add("pages[]", "prices")

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/promo?success=1", w.Header().Get("Location"))

	var p models.PromoPopup
	h.db.First(&p)
	assert.Equal(t, "Летняя акция", p.Title)
	assert.Equal(t, "Скидка 15% до конца июля", p.Content)
	assert.True(t, p.IsActive)
	assert.Equal(t, 48, p.TTLHours)
	assert.Equal(t, 5, p.ShowDelaySeconds)

	var pages []string
	json.Unmarshal([]byte(p.Pages), &pages)
	assert.Equal(t, []string{"home", "prices"}, pages)
}

func TestAdminPromoUpdate_IsActiveFalse(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.Create(&models.PromoPopup{IsActive: true, Pages: `["home"]`, TTLHours: 24})
	router.POST("/admin/promo", h.AdminPromoUpdate)

	form := url.Values{}
	form.Set("title", "Акция")
	// is_active не передан — должен стать false

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var p models.PromoPopup
	h.db.First(&p)
	assert.False(t, p.IsActive)
}

func TestAdminPromoUpdate_InvalidTTL_DefaultsTo24(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.Create(&models.PromoPopup{Pages: `["home"]`, TTLHours: 10})
	router.POST("/admin/promo", h.AdminPromoUpdate)

	form := url.Values{}
	form.Set("ttl_hours", "не_число")

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var p models.PromoPopup
	h.db.First(&p)
	assert.Equal(t, 24, p.TTLHours)
}

func TestAdminPromoUpdate_InvalidDelay_DefaultsTo0(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.Create(&models.PromoPopup{Pages: `["home"]`, ShowDelaySeconds: 10})
	router.POST("/admin/promo", h.AdminPromoUpdate)

	form := url.Values{}
	form.Set("show_delay_seconds", "abc")

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var p models.PromoPopup
	h.db.First(&p)
	assert.Equal(t, 0, p.ShowDelaySeconds)
}

func TestAdminPromoUpdate_EmptyPages_DefaultsToHome(t *testing.T) {
	router, h := setupTestRouter(t)
	h.db.Create(&models.PromoPopup{Pages: `["prices","projects"]`, TTLHours: 24})
	router.POST("/admin/promo", h.AdminPromoUpdate)

	// pages[] не передан
	form := url.Values{}
	form.Set("title", "Тест")

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	var p models.PromoPopup
	h.db.First(&p)
	var pages []string
	json.Unmarshal([]byte(p.Pages), &pages)
	assert.Equal(t, []string{"home"}, pages)
}

func TestAdminPromoUpdate_NoRecord_Returns404(t *testing.T) {
	router, h := setupTestRouter(t)
	// БД пустая — записи нет
	router.POST("/admin/promo", h.AdminPromoUpdate)

	form := url.Values{}
	form.Set("title", "Акция")

	req, _ := http.NewRequest("POST", "/admin/promo", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ============================================================================
// getAllPages Tests
// ============================================================================

func TestGetAllPages_ReturnsCorrectCount(t *testing.T) {
	pages := getAllPages()
	assert.Len(t, pages, 6)
}

func TestGetAllPages_ContainsExpectedIDs(t *testing.T) {
	pages := getAllPages()
	ids := make([]string, len(pages))
	for i, p := range pages {
		ids[i] = p.ID
	}
	assert.Contains(t, ids, "home")
	assert.Contains(t, ids, "prices")
	assert.Contains(t, ids, "projects")
	assert.Contains(t, ids, "services")
	assert.Contains(t, ids, "led-guide")
	assert.Contains(t, ids, "contact")
}

func TestGetAllPages_AllHaveNames(t *testing.T) {
	for _, p := range getAllPages() {
		assert.NotEmpty(t, p.Name, "страница %q должна иметь Name", p.ID)
	}
}
