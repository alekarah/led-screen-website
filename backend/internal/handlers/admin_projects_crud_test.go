package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledsite/internal/models"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// CreateProject Tests
// ============================================================================

func TestCreateProject_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects", h.CreateProject)

	// Создаем тестовую категорию
	category := models.Category{Name: "LED экраны", Slug: "led-screens"}
	h.db.Create(&category)

	// Подготовка multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Тестовый проект")
	writer.WriteField("description", "Описание проекта")
	writer.WriteField("location", "Москва")
	writer.WriteField("size", "3x2 м")
	writer.WriteField("pixel_pitch", "P10")
	writer.WriteField("categories", fmt.Sprintf("%d", category.ID))
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/projects", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "project_id")

	// Проверяем, что проект создан в БД
	var project models.Project
	h.db.Preload("Categories").First(&project, uint(response["project_id"].(float64)))
	assert.Equal(t, "Тестовый проект", project.Title)
	assert.Contains(t, project.Slug, "testovyy-proekt") // Проверка slug генерации (может быть с суффиксом)
	assert.Equal(t, "Описание проекта", project.Description)
	assert.Equal(t, 1, len(project.Categories))
}

func TestCreateProject_SlugGeneration(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects", h.CreateProject)

	// Создаем первый проект
	body1 := &bytes.Buffer{}
	writer1 := multipart.NewWriter(body1)
	writer1.WriteField("title", "LED экран")
	writer1.WriteField("description", "Первый проект")
	writer1.Close()

	req1, _ := http.NewRequest("POST", "/admin/projects", body1)
	req1.Header.Set("Content-Type", writer1.FormDataContentType())
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	projectID1 := uint(response1["project_id"].(float64))

	var project1 models.Project
	h.db.First(&project1, projectID1)
	slug1 := project1.Slug
	assert.Contains(t, slug1, "led-ekran") // Slug содержит базовое имя

	// Создаем второй проект с таким же названием
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	writer2.WriteField("title", "LED экран")
	writer2.WriteField("description", "Второй проект")
	writer2.Close()

	req2, _ := http.NewRequest("POST", "/admin/projects", body2)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	projectID2 := uint(response2["project_id"].(float64))

	var project2 models.Project
	h.db.First(&project2, projectID2)
	slug2 := project2.Slug
	// Slug должен быть уникальным (отличаться от первого)
	assert.Contains(t, slug2, "led-ekran")
	assert.NotEqual(t, slug1, slug2, "Slugs должны быть уникальными")
}

func TestCreateProject_Featured(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects", h.CreateProject)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Избранный проект")
	writer.WriteField("description", "Описание")
	writer.WriteField("featured", "on") // checkbox value
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/projects", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	var project models.Project
	h.db.First(&project, uint(response["project_id"].(float64)))
	assert.True(t, project.Featured)
}

// ============================================================================
// GetProject Tests
// ============================================================================

func TestGetProject_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/projects/:id", h.GetProject)

	// Создаем проект с категориями
	category := models.Category{Name: "Тест", Slug: "test"}
	h.db.Create(&category)

	project := models.Project{
		Title:       "Проект для получения",
		Description: "Описание",
		Slug:        "proekt-dlya-polucheniya",
	}
	h.db.Create(&project)
	h.db.Model(&project).Association("Categories").Append(&category)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/projects/%d", project.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	projectData := response["project"].(map[string]interface{})
	assert.Equal(t, "Проект для получения", projectData["title"])

	categories := projectData["categories"].([]interface{})
	assert.Equal(t, 1, len(categories))
}

func TestGetProject_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/projects/:id", h.GetProject)

	req, _ := http.NewRequest("GET", "/admin/projects/99999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProject_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/projects/:id", h.GetProject)

	req, _ := http.NewRequest("GET", "/admin/projects/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetProject_EmptyCategoriesAndImages(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/admin/projects/:id", h.GetProject)

	project := models.Project{
		Title:       "Проект без категорий",
		Description: "Описание",
		Slug:        "proekt-bez-kategoriy",
	}
	h.db.Create(&project)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/projects/%d", project.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	projectData := response["project"].(map[string]interface{})

	// Категории и изображения могут быть nil или пустым массивом
	if categories, ok := projectData["categories"].([]interface{}); ok {
		assert.Equal(t, 0, len(categories))
	}
	if images, ok := projectData["images"].([]interface{}); ok {
		assert.Equal(t, 0, len(images))
	}
}

// ============================================================================
// UpdateProject Tests
// ============================================================================

func TestUpdateProject_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/update", h.UpdateProject)

	project := models.Project{
		Title:       "Старое название",
		Description: "Старое описание",
		Slug:        "staroe-nazvanie",
		Featured:    false,
	}
	h.db.Create(&project)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Новое название")
	writer.WriteField("description", "Новое описание")
	writer.WriteField("location", "Санкт-Петербург")
	writer.WriteField("size", "5x3 м")
	writer.WriteField("pixel_pitch", "P5")
	writer.WriteField("featured", "on")
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/projects/%d/update", project.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Project
	h.db.First(&updated, project.ID)
	assert.Equal(t, "Новое название", updated.Title)
	assert.Equal(t, "Новое описание", updated.Description)
	assert.Equal(t, "Санкт-Петербург", updated.Location)
	assert.Equal(t, "5x3 м", updated.Size)
	assert.Equal(t, "P5", updated.PixelPitch)
	assert.True(t, updated.Featured)
}

func TestUpdateProject_UpdateCategories(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/update", h.UpdateProject)

	// Создаем категории
	category1 := models.Category{Name: "Категория 1", Slug: "category-1"}
	category2 := models.Category{Name: "Категория 2", Slug: "category-2"}
	h.db.Create(&category1)
	h.db.Create(&category2)

	// Создаем проект с первой категорией
	project := models.Project{
		Title:       "Проект",
		Description: "Описание",
		Slug:        "proekt",
	}
	h.db.Create(&project)
	h.db.Model(&project).Association("Categories").Append(&category1)

	// Обновляем проект, меняя категорию на вторую
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Проект")
	writer.WriteField("description", "Описание")
	writer.WriteField("categories", fmt.Sprintf("%d", category2.ID))
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/projects/%d/update", project.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что категории обновлены
	var updated models.Project
	h.db.Preload("Categories").First(&updated, project.ID)
	assert.Equal(t, 1, len(updated.Categories))
	assert.Equal(t, category2.ID, updated.Categories[0].ID)
}

func TestUpdateProject_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/update", h.UpdateProject)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Новое название")
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/projects/99999/update", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateProject_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/projects/:id/update", h.UpdateProject)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Новое название")
	writer.Close()

	req, _ := http.NewRequest("POST", "/admin/projects/invalid/update", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ============================================================================
// DeleteProject Tests
// ============================================================================

func TestDeleteProject_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/projects/:id", h.DeleteProject)

	// Создаем проект с категориями
	category := models.Category{Name: "Тест", Slug: "test"}
	h.db.Create(&category)

	project := models.Project{
		Title:       "Проект для удаления",
		Description: "Описание",
		Slug:        "proekt-dlya-udaleniya",
	}
	h.db.Create(&project)
	h.db.Model(&project).Association("Categories").Append(&category)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/projects/%d", project.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что проект удален
	var count int64
	h.db.Model(&models.Project{}).Where("id = ?", project.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	// Проверяем, что связь с категориями удалена
	var categoryCount int64
	h.db.Table("project_categories").Where("project_id = ?", project.ID).Count(&categoryCount)
	assert.Equal(t, int64(0), categoryCount)
}

func TestDeleteProject_WithImages(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/projects/:id", h.DeleteProject)

	project := models.Project{
		Title:       "Проект с изображениями",
		Description: "Описание",
		Slug:        "proekt-s-izobrazheniyami",
	}
	h.db.Create(&project)

	// Создаем тестовое изображение (без реального файла)
	projectID := project.ID
	image := models.Image{
		ProjectID: &projectID,
		FilePath:  "/fake/path/image.jpg",
		Filename:  "image.jpg",
		IsPrimary: true,
		SortOrder: 1,
		Caption:   "Тестовое изображение",
	}
	h.db.Create(&image)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/projects/%d", project.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что проект и изображения удалены
	var projectCount int64
	h.db.Model(&models.Project{}).Where("id = ?", project.ID).Count(&projectCount)
	assert.Equal(t, int64(0), projectCount)

	var imageCount int64
	h.db.Model(&models.Image{}).Where("project_id = ?", project.ID).Count(&imageCount)
	assert.Equal(t, int64(0), imageCount)
}

func TestDeleteProject_WithViewStats(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/projects/:id", h.DeleteProject)

	project := models.Project{
		Title:       "Проект со статистикой",
		Description: "Описание",
		Slug:        "proekt-so-statistikoy",
	}
	h.db.Create(&project)

	// Создаем статистику просмотров
	viewStat := models.ProjectViewDaily{
		ProjectID: project.ID,
		Views:     100,
	}
	h.db.Create(&viewStat)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/projects/%d", project.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что статистика тоже удалена
	var statCount int64
	h.db.Model(&models.ProjectViewDaily{}).Where("project_id = ?", project.ID).Count(&statCount)
	assert.Equal(t, int64(0), statCount)
}

func TestDeleteProject_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/projects/:id", h.DeleteProject)

	req, _ := http.NewRequest("DELETE", "/admin/projects/99999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteProject_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/projects/:id", h.DeleteProject)

	req, _ := http.NewRequest("DELETE", "/admin/projects/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
