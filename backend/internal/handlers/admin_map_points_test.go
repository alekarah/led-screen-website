package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------- CreateMapPoint ----------

func TestCreateMapPoint_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	form := url.Values{
		"title":        {"Варшавская 26"},
		"latitude":     {"59.8683"},
		"longitude":    {"30.3138"},
		"description":  {"Тестовая точка"},
		"panorama_url": {"https://yandex.ru/maps/panorama"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])

	// Проверяем в БД
	var point models.MapPoint
	h.db.First(&point)
	assert.Equal(t, "Варшавская 26", point.Title)
	assert.Equal(t, 59.8683, point.Latitude)
	assert.Equal(t, 30.3138, point.Longitude)
	assert.Equal(t, "Тестовая точка", point.Description)
	assert.Equal(t, "https://yandex.ru/maps/panorama", point.PanoramaURL)
	assert.True(t, point.IsActive)
}

func TestCreateMapPoint_EmptyTitle(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	form := url.Values{
		"title":     {""},
		"latitude":  {"59.8683"},
		"longitude": {"30.3138"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMapPoint_InvalidLatitude(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	form := url.Values{
		"title":     {"Точка"},
		"latitude":  {"abc"},
		"longitude": {"30.3138"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "широта")
}

func TestCreateMapPoint_InvalidLongitude(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	form := url.Values{
		"title":     {"Точка"},
		"latitude":  {"59.8683"},
		"longitude": {"xyz"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "долгота")
}

func TestCreateMapPoint_DefaultIsActive(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	// Без явного is_active — по умолчанию true
	form := url.Values{
		"title":     {"Точка по умолчанию"},
		"latitude":  {"59.0"},
		"longitude": {"30.0"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var point models.MapPoint
	h.db.First(&point)
	assert.True(t, point.IsActive)
}

// ---------- Дубликаты ----------

func TestCreateMapPoint_Duplicate(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	// Создаём первую точку
	h.db.Create(&models.MapPoint{Title: "Существующая", Latitude: 59.8683, Longitude: 30.3138, IsActive: true})

	// Пытаемся создать точку с теми же координатами
	form := url.Values{
		"title":     {"Дубликат"},
		"latitude":  {"59.8683"},
		"longitude": {"30.3138"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "уже существует")

	// В БД должна быть только 1 точка
	var count int64
	h.db.Model(&models.MapPoint{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCreateMapPoint_NearbyNotDuplicate(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points", h.CreateMapPoint)

	// Создаём первую точку
	h.db.Create(&models.MapPoint{Title: "Точка A", Latitude: 59.8683, Longitude: 30.3138, IsActive: true})

	// Точка на расстоянии >11м (delta > 0.0001)
	form := url.Values{
		"title":     {"Точка B"},
		"latitude":  {"59.870"},
		"longitude": {"30.315"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	h.db.Model(&models.MapPoint{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestBulkImport_SkipsDuplicates(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/bulk-import", h.BulkImportMapPoints)

	// Создаём существующую точку
	h.db.Create(&models.MapPoint{Title: "Существующая", Latitude: 59.8683, Longitude: 30.3138, IsActive: true})

	body, _ := json.Marshal(gin.H{
		"links": []string{
			"https://yandex.ru/maps/?ll=30.3138,59.8683", // дубликат
			"https://yandex.ru/maps/?ll=30.5000,60.0000", // новая
		},
	})

	req, _ := http.NewRequest("POST", "/admin/map-points/bulk-import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["created"])
	errors := resp["errors"].([]interface{})
	assert.Equal(t, 1, len(errors))
	assert.Contains(t, errors[0], "уже существует")
}

// ---------- GetMapPoint ----------

func TestGetMapPoint_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/map-points/:id", h.GetMapPoint)

	point := models.MapPoint{Title: "Тестовая", Latitude: 59.0, Longitude: 30.0}
	h.db.Create(&point)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/map-points/%d", point.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	mp := resp["map_point"].(map[string]interface{})
	assert.Equal(t, "Тестовая", mp["title"])
}

func TestGetMapPoint_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/map-points/:id", h.GetMapPoint)

	req, _ := http.NewRequest("GET", "/admin/map-points/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMapPoint_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/map-points/:id", h.GetMapPoint)

	req, _ := http.NewRequest("GET", "/admin/map-points/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- UpdateMapPoint ----------

func TestUpdateMapPoint_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/:id/update", h.UpdateMapPoint)

	point := models.MapPoint{Title: "Старое", Latitude: 59.0, Longitude: 30.0}
	h.db.Create(&point)

	form := url.Values{
		"title":        {"Новое название"},
		"latitude":     {"60.0"},
		"longitude":    {"31.0"},
		"description":  {"Обновлено"},
		"panorama_url": {"https://panorama.test"},
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/map-points/%d/update", point.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.MapPoint
	h.db.First(&updated, point.ID)
	assert.Equal(t, "Новое название", updated.Title)
	assert.Equal(t, 60.0, updated.Latitude)
	assert.Equal(t, 31.0, updated.Longitude)
	assert.Equal(t, "Обновлено", updated.Description)
	assert.Equal(t, "https://panorama.test", updated.PanoramaURL)
}

func TestUpdateMapPoint_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/:id/update", h.UpdateMapPoint)

	form := url.Values{
		"title":     {"Тест"},
		"latitude":  {"59.0"},
		"longitude": {"30.0"},
	}

	req, _ := http.NewRequest("POST", "/admin/map-points/999/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMapPoint_EmptyTitle(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/:id/update", h.UpdateMapPoint)

	point := models.MapPoint{Title: "Точка", Latitude: 59.0, Longitude: 30.0}
	h.db.Create(&point)

	form := url.Values{
		"title":     {"  "},
		"latitude":  {"59.0"},
		"longitude": {"30.0"},
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/map-points/%d/update", point.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- DeleteMapPoint ----------

func TestDeleteMapPoint_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/map-points/:id", h.DeleteMapPoint)

	point := models.MapPoint{Title: "Удалить", Latitude: 59.0, Longitude: 30.0}
	h.db.Create(&point)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/map-points/%d", point.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	h.db.Model(&models.MapPoint{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteMapPoint_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/map-points/:id", h.DeleteMapPoint)

	req, _ := http.NewRequest("DELETE", "/admin/map-points/0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- UpdateMapPointsSorting ----------

func TestUpdateMapPointsSorting_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/sort", h.UpdateMapPointsSorting)

	p1 := models.MapPoint{Title: "Первая", Latitude: 59.0, Longitude: 30.0}
	p2 := models.MapPoint{Title: "Вторая", Latitude: 60.0, Longitude: 31.0}
	h.db.Create(&p1)
	h.db.Create(&p2)

	body, _ := json.Marshal(gin.H{"ids": []uint{p2.ID, p1.ID}})
	req, _ := http.NewRequest("POST", "/admin/map-points/sort", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем порядок
	var point1, point2 models.MapPoint
	h.db.First(&point1, p1.ID)
	h.db.First(&point2, p2.ID)
	assert.Equal(t, 1, point1.SortOrder) // p1 стала второй
	assert.Equal(t, 0, point2.SortOrder) // p2 стала первой
}

func TestUpdateMapPointsSorting_BadJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/sort", h.UpdateMapPointsSorting)

	req, _ := http.NewRequest("POST", "/admin/map-points/sort", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- BulkImportMapPoints ----------

func TestBulkImport_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/bulk-import", h.BulkImportMapPoints)

	body, _ := json.Marshal(gin.H{
		"links": []string{
			"https://yandex.ru/maps/2/saint-petersburg/house/varshavskaya_ulitsa_26/?ll=30.3138,59.8683&z=17",
			"https://yandex.ru/maps/2/saint-petersburg/house/nevskiy_prospekt_1/?ll=30.3351,59.9386&z=17",
		},
	})

	req, _ := http.NewRequest("POST", "/admin/map-points/bulk-import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2), resp["created"])

	var count int64
	h.db.Model(&models.MapPoint{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// Проверяем извлечённое название
	var point models.MapPoint
	h.db.First(&point)
	assert.Contains(t, point.Title, "varshavskaya")
}

func TestBulkImport_MixedValidInvalid(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/bulk-import", h.BulkImportMapPoints)

	body, _ := json.Marshal(gin.H{
		"links": []string{
			"https://yandex.ru/maps/?ll=30.3138,59.8683",
			"https://example.com/no-coords",
			"",
		},
	})

	req, _ := http.NewRequest("POST", "/admin/map-points/bulk-import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["created"])
	errors := resp["errors"].([]interface{})
	assert.Equal(t, 1, len(errors)) // example.com без ll
}

func TestBulkImport_BadJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/map-points/bulk-import", h.BulkImportMapPoints)

	req, _ := http.NewRequest("POST", "/admin/map-points/bulk-import", strings.NewReader("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- parseCoordsFromYandexURL ----------

func TestParseCoordsFromYandexURL_Valid(t *testing.T) {
	lat, lng, err := parseCoordsFromYandexURL("https://yandex.ru/maps/?ll=30.3138,59.8683&z=17")
	assert.NoError(t, err)
	assert.InDelta(t, 59.8683, lat, 0.0001)
	assert.InDelta(t, 30.3138, lng, 0.0001)
}

func TestParseCoordsFromYandexURL_NoLLParam(t *testing.T) {
	_, _, err := parseCoordsFromYandexURL("https://yandex.ru/maps/?z=17")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ll не найден")
}

func TestParseCoordsFromYandexURL_MalformedLL(t *testing.T) {
	_, _, err := parseCoordsFromYandexURL("https://yandex.ru/maps/?ll=30.3138")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "формат ll")
}

func TestParseCoordsFromYandexURL_InvalidCoord(t *testing.T) {
	_, _, err := parseCoordsFromYandexURL("https://yandex.ru/maps/?ll=abc,def")
	assert.Error(t, err)
}

func TestParseCoordsFromYandexURL_BadURL(t *testing.T) {
	_, _, err := parseCoordsFromYandexURL("://bad-url")
	assert.Error(t, err)
}

// ---------- extractTitleFromURL ----------

func TestExtractTitleFromURL_WithHouse(t *testing.T) {
	title := extractTitleFromURL("https://yandex.ru/maps/2/saint-petersburg/house/varshavskaya_ulitsa_26/?ll=30.3138,59.8683", 1)
	assert.Equal(t, "varshavskaya ulitsa 26", title)
}

func TestExtractTitleFromURL_NoHouse(t *testing.T) {
	title := extractTitleFromURL("https://yandex.ru/maps/?ll=30.3138,59.8683", 5)
	assert.Equal(t, "Точка 5", title)
}

func TestExtractTitleFromURL_InvalidURL(t *testing.T) {
	title := extractTitleFromURL("://bad", 3)
	assert.Equal(t, "Точка 3", title)
}

func TestExtractTitleFromURL_HouseAtEndOfPath(t *testing.T) {
	// /house/ в конце URL — segments[i+1] = "" (пустая строка после trailing slash)
	// Текущая реализация вернёт пустую строку — фиксируем поведение
	title := extractTitleFromURL("https://yandex.ru/maps/house/", 2)
	assert.Equal(t, "", title)
}
