package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
