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

	"github.com/stretchr/testify/assert"
)

// ---------- ReorderProject ----------

func TestReorderProject_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/reorder", h.ReorderProject)

	project := createTestProject(t, h, "Тестовый проект")
	assert.Equal(t, 0, project.SortOrder, "изначально sort_order должен быть 0")

	form := url.Values{}
	form.Set("position", "5")

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/projects/%d/reorder", project.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что порядок обновился
	var updated models.Project
	h.db.First(&updated, project.ID)
	assert.Equal(t, 5, updated.SortOrder)
}

func TestReorderProject_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/reorder", h.ReorderProject)

	form := url.Values{}
	form.Set("position", "5")

	req, _ := http.NewRequest("POST", "/admin/projects/999/reorder", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReorderProject_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/reorder", h.ReorderProject)

	form := url.Values{}
	form.Set("position", "5")

	req, _ := http.NewRequest("POST", "/admin/projects/abc/reorder", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReorderProject_InvalidPosition(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/reorder", h.ReorderProject)

	project := createTestProject(t, h, "Тестовый проект")

	form := url.Values{}
	form.Set("position", "invalid")

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/projects/%d/reorder", project.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "позиция")
}

func TestReorderProject_NegativePosition(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/reorder", h.ReorderProject)

	project := createTestProject(t, h, "Тестовый проект")

	form := url.Values{}
	form.Set("position", "-5")

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/projects/%d/reorder", project.ID), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Негативная позиция технически валидна (может быть для сортировки в обратном порядке)
	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Project
	h.db.First(&updated, project.ID)
	assert.Equal(t, -5, updated.SortOrder)
}

// ---------- BulkReorderProjects ----------

func TestBulkReorderProjects_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	// Создаём 3 проекта
	p1 := createTestProject(t, h, "Проект 1")
	p2 := createTestProject(t, h, "Проект 2")
	p3 := createTestProject(t, h, "Проект 3")

	// Меняем порядок: 3, 1, 2
	reqData := BulkReorderRequest{
		Projects: []ProjectOrderData{
			{ID: int(p3.ID), SortOrder: 0},
			{ID: int(p1.ID), SortOrder: 1},
			{ID: int(p2.ID), SortOrder: 2},
		},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что порядок обновился
	var updated1, updated2, updated3 models.Project
	h.db.First(&updated1, p1.ID)
	h.db.First(&updated2, p2.ID)
	h.db.First(&updated3, p3.ID)

	assert.Equal(t, 1, updated1.SortOrder)
	assert.Equal(t, 2, updated2.SortOrder)
	assert.Equal(t, 0, updated3.SortOrder)
}

func TestBulkReorderProjects_EmptyList(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	reqData := BulkReorderRequest{
		Projects: []ProjectOrderData{},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Пустой список")
}

func TestBulkReorderProjects_InvalidJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBulkReorderProjects_WithNonExistentID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	p1 := createTestProject(t, h, "Проект 1")

	// GORM Update не выдаёт ошибку для несуществующего ID, просто игнорирует его
	reqData := BulkReorderRequest{
		Projects: []ProjectOrderData{
			{ID: int(p1.ID), SortOrder: 5},
			{ID: 99999, SortOrder: 10}, // несуществующий проект (будет проигнорирован)
		},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// GORM не считает это ошибкой
	assert.Equal(t, http.StatusOK, w.Code)

	// p1 должен обновиться
	var updated models.Project
	h.db.First(&updated, p1.ID)
	assert.Equal(t, 5, updated.SortOrder)
}

func TestBulkReorderProjects_SingleProject(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	p1 := createTestProject(t, h, "Единственный проект")

	reqData := BulkReorderRequest{
		Projects: []ProjectOrderData{
			{ID: int(p1.ID), SortOrder: 42},
		},
	}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Project
	h.db.First(&updated, p1.ID)
	assert.Equal(t, 42, updated.SortOrder)
}

func TestBulkReorderProjects_ManyProjects(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/bulk-reorder", h.BulkReorderProjects)

	// Создаём 10 проектов
	var projects []models.Project
	for i := 0; i < 10; i++ {
		p := createTestProject(t, h, fmt.Sprintf("Проект %d", i))
		projects = append(projects, p)
	}

	// Обратный порядок
	var orderData []ProjectOrderData
	for i, p := range projects {
		orderData = append(orderData, ProjectOrderData{
			ID:        int(p.ID),
			SortOrder: 9 - i, // обратный порядок
		})
	}

	reqData := BulkReorderRequest{Projects: orderData}
	jsonData, _ := json.Marshal(reqData)

	req, _ := http.NewRequest("POST", "/admin/projects/bulk-reorder", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что все обновились
	for i, p := range projects {
		var updated models.Project
		h.db.First(&updated, p.ID)
		assert.Equal(t, 9-i, updated.SortOrder)
	}
}
