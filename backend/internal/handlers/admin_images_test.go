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

// createTestProject создаёт тестовый проект в БД
func createTestProject(t *testing.T, h *Handlers, title string) models.Project {
	project := models.Project{
		Title: title,
		Slug:  generateSlug(title),
	}
	err := h.db.Create(&project).Error
	assert.NoError(t, err)
	return project
}

// createTestImage создаёт тестовое изображение в БД (без реального файла)
func createTestImage(t *testing.T, h *Handlers, projectID *uint, isPrimary bool) models.Image {
	image := models.Image{
		ProjectID:    projectID,
		Filename:     "test.jpg",
		OriginalName: "test.jpg",
		FilePath:     "/tmp/test.jpg",
		FileSize:     1024,
		MimeType:     "image/jpeg",
		IsPrimary:    isPrimary,
		SortOrder:    0,
	}
	err := h.db.Create(&image).Error
	assert.NoError(t, err)
	return image
}

// ---------- DeleteImage ----------

func TestDeleteImage_Success(t *testing.T) {
	t.Skip("Пропуск: требует реального файла для удаления")

	router, h := setupTestRouter(t)
	router.DELETE("/admin/images/:id", h.DeleteImage)

	project := createTestProject(t, h, "Тестовый проект")
	image := createTestImage(t, h, &project.ID, false)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/admin/images/%d", image.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что изображение удалено из БД
	var deleted models.Image
	err := h.db.First(&deleted, image.ID).Error
	assert.Error(t, err)
}

func TestDeleteImage_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/images/:id", h.DeleteImage)

	req, _ := http.NewRequest("DELETE", "/admin/images/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteImage_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.DELETE("/admin/images/:id", h.DeleteImage)

	req, _ := http.NewRequest("DELETE", "/admin/images/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- SetPrimaryImage ----------

func TestSetPrimaryImage_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/set-primary", h.SetPrimaryImage)

	project := createTestProject(t, h, "Тестовый проект")
	image1 := createTestImage(t, h, &project.ID, true)  // текущее главное
	image2 := createTestImage(t, h, &project.ID, false) // делаем главным

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/images/%d/set-primary", image2.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что image2 стало главным, а image1 нет
	var updated1, updated2 models.Image
	h.db.First(&updated1, image1.ID)
	h.db.First(&updated2, image2.ID)

	assert.False(t, updated1.IsPrimary, "старое главное изображение должно потерять флаг")
	assert.True(t, updated2.IsPrimary, "новое изображение должно стать главным")
}

func TestSetPrimaryImage_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/set-primary", h.SetPrimaryImage)

	req, _ := http.NewRequest("POST", "/admin/images/999/set-primary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetPrimaryImage_NoProjectID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/set-primary", h.SetPrimaryImage)

	// Изображение без project_id
	image := createTestImage(t, h, nil, false)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/images/%d/set-primary", image.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не привязано к проекту")
}

func TestSetPrimaryImage_OnlyOnePrimary(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/set-primary", h.SetPrimaryImage)

	project := createTestProject(t, h, "Тестовый проект")
	createTestImage(t, h, &project.ID, true)
	image2 := createTestImage(t, h, &project.ID, false)

	// Устанавливаем image2 как главное
	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/images/%d/set-primary", image2.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что только одно изображение главное
	var primaryCount int64
	h.db.Model(&models.Image{}).Where("project_id = ? AND is_primary = ?", project.ID, true).Count(&primaryCount)
	assert.Equal(t, int64(1), primaryCount, "должно быть только одно главное изображение")
}

// ---------- UpdateImageCrop ----------

func TestUpdateImageCrop_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/crop", h.UpdateImageCrop)

	cropData := map[string]float64{
		"crop_x":     25.5,
		"crop_y":     75.0,
		"crop_scale": 1.5,
	}
	jsonData, _ := json.Marshal(cropData)

	req, _ := http.NewRequest("POST", "/admin/images/999/crop", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateImageCrop_InvalidJSON(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/crop", h.UpdateImageCrop)

	project := createTestProject(t, h, "Тестовый проект")
	image := createTestImage(t, h, &project.ID, false)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/admin/images/%d/crop", image.ID), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateImageCrop_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/admin/images/:id/crop", h.UpdateImageCrop)

	req, _ := http.NewRequest("POST", "/admin/images/abc/crop", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------- Вспомогательные функции ----------

func TestGenerateImageFilename(t *testing.T) {
	filename := generateImageFilename(123, 0, "photo.jpg")
	assert.Contains(t, filename, "project_123_")
	assert.Contains(t, filename, ".jpg")
}

func TestValidateCropData_ValidValues(t *testing.T) {
	data := CropData{X: 25.5, Y: 75.0, Scale: 1.5}
	validated := validateCropData(data)

	assert.Equal(t, 25.5, validated.X)
	assert.Equal(t, 75.0, validated.Y)
	assert.Equal(t, 1.5, validated.Scale)
}

func TestValidateCropData_XOutOfRange(t *testing.T) {
	// X < 0
	data := CropData{X: -10, Y: 50, Scale: 1.0}
	validated := validateCropData(data)
	assert.Equal(t, 50.0, validated.X, "X должен быть 50 при значении < 0")

	// X > 100
	data = CropData{X: 150, Y: 50, Scale: 1.0}
	validated = validateCropData(data)
	assert.Equal(t, 50.0, validated.X, "X должен быть 50 при значении > 100")
}

func TestValidateCropData_YOutOfRange(t *testing.T) {
	// Y < 0
	data := CropData{X: 50, Y: -5, Scale: 1.0}
	validated := validateCropData(data)
	assert.Equal(t, 50.0, validated.Y, "Y должен быть 50 при значении < 0")

	// Y > 100
	data = CropData{X: 50, Y: 200, Scale: 1.0}
	validated = validateCropData(data)
	assert.Equal(t, 50.0, validated.Y, "Y должен быть 50 при значении > 100")
}

func TestValidateCropData_ScaleOutOfRange(t *testing.T) {
	// Scale < 1.0
	data := CropData{X: 50, Y: 50, Scale: 0.5}
	validated := validateCropData(data)
	assert.Equal(t, 1.0, validated.Scale, "Scale должен быть 1.0 при значении < 1.0")

	// Scale > 3.0
	data = CropData{X: 50, Y: 50, Scale: 5.0}
	validated = validateCropData(data)
	assert.Equal(t, 1.0, validated.Scale, "Scale должен быть 1.0 при значении > 3.0")
}

func TestValidateCropData_BoundaryValues(t *testing.T) {
	// Граничные значения должны быть валидны
	data := CropData{X: 0, Y: 100, Scale: 1.0}
	validated := validateCropData(data)
	assert.Equal(t, 0.0, validated.X)
	assert.Equal(t, 100.0, validated.Y)
	assert.Equal(t, 1.0, validated.Scale)

	data = CropData{X: 100, Y: 0, Scale: 3.0}
	validated = validateCropData(data)
	assert.Equal(t, 100.0, validated.X)
	assert.Equal(t, 0.0, validated.Y)
	assert.Equal(t, 3.0, validated.Scale)
}

func TestCreateImageRecord(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "original.jpg",
		Size:     2048,
		Header:   map[string][]string{"Content-Type": {"image/jpeg"}},
	}

	image := createImageRecord(42, "generated.jpg", file, "/path/to/file.jpg", 5)

	assert.Equal(t, uint(42), *image.ProjectID)
	assert.Equal(t, "generated.jpg", image.Filename)
	assert.Equal(t, "original.jpg", image.OriginalName)
	assert.Equal(t, "/path/to/file.jpg", image.FilePath)
	assert.Equal(t, int64(2048), image.FileSize)
	assert.Equal(t, "image/jpeg", image.MimeType)
	assert.Equal(t, 5, image.SortOrder)
}
