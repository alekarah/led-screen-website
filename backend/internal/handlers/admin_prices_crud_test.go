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

// createTestPriceItem создаёт тестовую позицию прайса в БД
func createTestPriceItem(t *testing.T, h *Handlers, title string, priceFrom int, isActive bool) models.PriceItem {
	item := models.PriceItem{
		Title:       title,
		Description: "Описание " + title,
		PriceFrom:   priceFrom,
		IsActive:    isActive,
		SortOrder:   0,
	}
	err := h.db.Create(&item).Error
	assert.NoError(t, err)
	return item
}

// postPriceForm отправляет POST с form-data для создания/обновления позиции прайса
func postPriceForm(router *gin.Engine, method, path string, data map[string]string) *httptest.ResponseRecorder {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}

	req, _ := http.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------- CreatePriceItem ----------

func TestCreatePriceItem_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	data := map[string]string{
		"title":      "Билборд 6x3",
		"description": "Большой билборд",
		"price_from":  "150000",
		"is_active":   "on",
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["price_id"])

	// Проверяем что позиция создалась в БД
	var item models.PriceItem
	h.db.First(&item)
	assert.Equal(t, "Билборд 6x3", item.Title)
	assert.Equal(t, 150000, item.PriceFrom)
	assert.True(t, item.IsActive)
}

func TestCreatePriceItem_EmptyTitle(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	data := map[string]string{
		"title":      "",
		"price_from": "100000",
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "пустым")
}

func TestCreatePriceItem_InvalidPrice(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	data := map[string]string{
		"title":      "Тестовая позиция",
		"price_from": "invalid",
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "цена")
}

func TestCreatePriceItem_NegativePrice(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	data := map[string]string{
		"title":      "Тестовая позиция",
		"price_from": "-1000",
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePriceItem_WithSpecifications(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	specs := []map[string]interface{}{
		{"group": "Размеры", "key": "Ширина", "value": "6 м", "order": 0},
		{"group": "Размеры", "key": "Высота", "value": "3 м", "order": 1},
	}
	specsJSON, _ := json.Marshal(specs)

	data := map[string]string{
		"title":              "Билборд с характеристиками",
		"price_from":         "200000",
		"has_specifications": "on",
		"specifications":     string(specsJSON),
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что характеристики создались
	var item models.PriceItem
	h.db.Preload("Specifications").First(&item)
	assert.True(t, item.HasSpecifications)
	assert.Equal(t, 2, len(item.Specifications))
	assert.Equal(t, "Ширина", item.Specifications[0].SpecKey)
}

func TestCreatePriceItem_WithIsActiveChecked(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices", h.CreatePriceItem)

	data := map[string]string{
		"title":     "Активная позиция",
		"price_from": "100000",
		"is_active":  "on", // чекбокс отмечен
	}

	w := postPriceForm(router, "POST", "/admin/prices", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var item models.PriceItem
	h.db.First(&item)
	assert.True(t, item.IsActive)
}

// ---------- GetPriceItem ----------

func TestGetPriceItem_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/prices/:id", h.GetPriceItem)

	item := createTestPriceItem(t, h, "Билборд", 150000, true)

	req, _ := http.NewRequest("GET", "/admin/prices/"+fmt.Sprintf("%d", item.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	priceItem := response["price_item"].(map[string]interface{})
	assert.Equal(t, "Билборд", priceItem["title"])
}

func TestGetPriceItem_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/prices/:id", h.GetPriceItem)

	req, _ := http.NewRequest("GET", "/admin/prices/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPriceItem_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/prices/:id", h.GetPriceItem)

	req, _ := http.NewRequest("GET", "/admin/prices/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPriceItem_WithSpecifications(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/prices/:id", h.GetPriceItem)

	// Создаём позицию с характеристиками
	item := createTestPriceItem(t, h, "Билборд", 150000, true)
	item.HasSpecifications = true
	h.db.Save(&item)

	spec := models.PriceSpecification{
		PriceItemID: item.ID,
		SpecGroup:   "Размеры",
		SpecKey:     "Ширина",
		SpecValue:   "6 м",
		SortOrder:   0,
	}
	h.db.Create(&spec)

	req, _ := http.NewRequest("GET", "/admin/prices/"+fmt.Sprintf("%d", item.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	priceItem := response["price_item"].(map[string]interface{})
	specs := priceItem["specifications"].([]interface{})
	assert.Equal(t, 1, len(specs))
}

// ---------- UpdatePriceItem ----------

func TestUpdatePriceItem_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/update", h.UpdatePriceItem)

	item := createTestPriceItem(t, h, "Старое название", 100000, true)

	data := map[string]string{
		"title":      "Новое название",
		"description": "Новое описание",
		"price_from":  "200000",
		"is_active":   "on",
	}

	w := postPriceForm(router, "POST", "/admin/prices/"+fmt.Sprintf("%d", item.ID)+"/update", data)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что данные обновились
	var updated models.PriceItem
	h.db.First(&updated, item.ID)
	assert.Equal(t, "Новое название", updated.Title)
	assert.Equal(t, 200000, updated.PriceFrom)
}

func TestUpdatePriceItem_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/update", h.UpdatePriceItem)

	data := map[string]string{
		"title":      "Название",
		"price_from": "100000",
	}

	w := postPriceForm(router, "POST", "/admin/prices/999/update", data)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdatePriceItem_EmptyTitle(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/update", h.UpdatePriceItem)

	item := createTestPriceItem(t, h, "Билборд", 100000, true)

	data := map[string]string{
		"title":      "",
		"price_from": "100000",
	}

	w := postPriceForm(router, "POST", "/admin/prices/"+fmt.Sprintf("%d", item.ID)+"/update", data)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePriceItem_UpdatesSpecifications(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/update", h.UpdatePriceItem)

	// Создаём позицию с характеристиками
	item := createTestPriceItem(t, h, "Билборд", 150000, true)
	item.HasSpecifications = true
	h.db.Save(&item)

	oldSpec := models.PriceSpecification{
		PriceItemID: item.ID,
		SpecGroup:   "Старая группа",
		SpecKey:     "Старый ключ",
		SpecValue:   "Старое значение",
	}
	h.db.Create(&oldSpec)

	// Обновляем с новыми характеристиками
	newSpecs := []map[string]interface{}{
		{"group": "Новая группа", "key": "Новый ключ", "value": "Новое значение", "order": 0},
	}
	specsJSON, _ := json.Marshal(newSpecs)

	data := map[string]string{
		"title":              "Билборд",
		"price_from":         "150000",
		"has_specifications": "on",
		"specifications":     string(specsJSON),
	}

	w := postPriceForm(router, "POST", "/admin/prices/"+fmt.Sprintf("%d", item.ID)+"/update", data)

	assert.Equal(t, http.StatusOK, w.Code)

	// Старые характеристики должны быть удалены
	var specs []models.PriceSpecification
	h.db.Where("price_item_id = ?", item.ID).Find(&specs)
	assert.Equal(t, 1, len(specs))
	assert.Equal(t, "Новый ключ", specs[0].SpecKey)
}

// ---------- DeletePriceItem ----------

func TestDeletePriceItem_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/prices/:id", h.DeletePriceItem)

	item := createTestPriceItem(t, h, "Билборд", 150000, true)

	req, _ := http.NewRequest("DELETE", "/admin/prices/"+fmt.Sprintf("%d", item.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что позиция удалена
	var deleted models.PriceItem
	err := h.db.First(&deleted, item.ID).Error
	assert.Error(t, err, "позиция должна быть удалена")
}

func TestDeletePriceItem_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/prices/:id", h.DeletePriceItem)

	req, _ := http.NewRequest("DELETE", "/admin/prices/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeletePriceItem_WithSpecifications(t *testing.T) {
	t.Skip("Пропуск: SQLite in-memory может не поддерживать CASCADE delete для спецификаций")

	router, h := setupTestRouter(t)
	router.DELETE("/admin/prices/:id", h.DeletePriceItem)

	// Создаём позицию с характеристиками
	item := createTestPriceItem(t, h, "Билборд", 150000, true)
	spec := models.PriceSpecification{
		PriceItemID: item.ID,
		SpecGroup:   "Размеры",
		SpecKey:     "Ширина",
		SpecValue:   "6 м",
	}
	h.db.Create(&spec)

	req, _ := http.NewRequest("DELETE", "/admin/prices/"+fmt.Sprintf("%d", item.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Характеристики тоже должны быть удалены (CASCADE)
	var specs []models.PriceSpecification
	h.db.Where("price_item_id = ?", item.ID).Find(&specs)
	assert.Equal(t, 0, len(specs))
}

// ---------- DuplicatePriceItem ----------

func TestDuplicatePriceItem_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/duplicate", h.DuplicatePriceItem)

	original := createTestPriceItem(t, h, "Оригинал", 100000, true)

	req, _ := http.NewRequest("POST", "/admin/prices/"+fmt.Sprintf("%d", original.ID)+"/duplicate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что создалась копия
	var items []models.PriceItem
	h.db.Find(&items)
	assert.Equal(t, 2, len(items))

	// Находим копию
	var duplicate models.PriceItem
	h.db.Where("title = ?", "Оригинал (копия)").First(&duplicate)
	assert.Equal(t, original.PriceFrom, duplicate.PriceFrom)
	assert.Equal(t, original.IsActive, duplicate.IsActive)
}

func TestDuplicatePriceItem_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/duplicate", h.DuplicatePriceItem)

	req, _ := http.NewRequest("POST", "/admin/prices/999/duplicate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDuplicatePriceItem_CopiesSpecifications(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/duplicate", h.DuplicatePriceItem)

	// Создаём позицию с характеристиками
	original := createTestPriceItem(t, h, "Оригинал", 100000, true)
	original.HasSpecifications = true
	h.db.Save(&original)

	spec := models.PriceSpecification{
		PriceItemID: original.ID,
		SpecGroup:   "Размеры",
		SpecKey:     "Ширина",
		SpecValue:   "6 м",
		SortOrder:   0,
	}
	h.db.Create(&spec)

	req, _ := http.NewRequest("POST", "/admin/prices/"+fmt.Sprintf("%d", original.ID)+"/duplicate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Находим копию
	var duplicate models.PriceItem
	h.db.Preload("Specifications").Where("title = ?", "Оригинал (копия)").First(&duplicate)
	assert.Equal(t, 1, len(duplicate.Specifications))
	assert.Equal(t, "Ширина", duplicate.Specifications[0].SpecKey)
}

func TestDuplicatePriceItem_SetsCorrectSortOrder(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/:id/duplicate", h.DuplicatePriceItem)

	// Создаём несколько позиций с разными sort_order
	item1 := createTestPriceItem(t, h, "Позиция 1", 100000, true)
	item1.SortOrder = 5
	h.db.Save(&item1)

	item2 := createTestPriceItem(t, h, "Позиция 2", 200000, true)
	item2.SortOrder = 10
	h.db.Save(&item2)

	req, _ := http.NewRequest("POST", "/admin/prices/"+fmt.Sprintf("%d", item1.ID)+"/duplicate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Копия должна получить sort_order = max + 1 = 11
	var duplicate models.PriceItem
	h.db.Where("title = ?", "Позиция 1 (копия)").First(&duplicate)
	assert.Equal(t, 11, duplicate.SortOrder)
}

// ---------- UpdatePriceItemsSorting ----------

func TestUpdatePriceItemsSorting_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/sort", h.UpdatePriceItemsSorting)

	item1 := createTestPriceItem(t, h, "Позиция 1", 100000, true)
	item2 := createTestPriceItem(t, h, "Позиция 2", 200000, true)
	item3 := createTestPriceItem(t, h, "Позиция 3", 300000, true)

	// Меняем порядок: 3, 1, 2
	reqData := map[string][]uint{
		"ids": {item3.ID, item1.ID, item2.ID},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/prices/sort", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что порядок обновился
	var updated1, updated2, updated3 models.PriceItem
	h.db.First(&updated1, item1.ID)
	h.db.First(&updated2, item2.ID)
	h.db.First(&updated3, item3.ID)

	assert.Equal(t, 1, updated1.SortOrder) // было 0, стало 1 (второй в списке)
	assert.Equal(t, 2, updated2.SortOrder) // было 0, стало 2 (третий)
	assert.Equal(t, 0, updated3.SortOrder) // было 0, стало 0 (первый)
}

func TestUpdatePriceItemsSorting_BadJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/sort", h.UpdatePriceItemsSorting)

	req, _ := http.NewRequest("POST", "/admin/prices/sort", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePriceItemsSorting_EmptyIDs(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/prices/sort", h.UpdatePriceItemsSorting)

	reqData := map[string][]uint{
		"ids": {},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/prices/sort", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Пустой массив - валидный запрос, просто ничего не обновляется
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------- convertToWebPath ----------

func TestConvertToWebPath_Unix(t *testing.T) {
	path := "/home/user/frontend/static/uploads/file.png"
	result := convertToWebPath(path)
	assert.Equal(t, "/static/uploads/file.png", result)
}

func TestConvertToWebPath_Windows(t *testing.T) {
	path := "C:\\projects\\frontend\\static\\uploads\\file.png"
	result := convertToWebPath(path)
	assert.Equal(t, "/static/uploads/file.png", result)
}

func TestConvertToWebPath_NoMatch(t *testing.T) {
	path := "/some/other/path/file.png"
	result := convertToWebPath(path)
	assert.Equal(t, "/some/other/path/file.png", result)
}
