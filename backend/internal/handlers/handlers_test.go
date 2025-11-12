package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB создает тестовую БД в памяти для изоляции тестов
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Автоматическая миграция всех моделей
	err = db.AutoMigrate(
		&models.Category{},
		&models.Project{},
		&models.Image{},
		&models.Service{},
		&models.ContactForm{},
		&models.ContactNote{},
		&models.ProjectViewDaily{},
		&models.Admin{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupTestRouter создает тестовый Gin router с тестовой БД
func setupTestRouter(t *testing.T) (*gin.Engine, *Handlers) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	db := setupTestDB(t)
	handlers := New(db, 10*1024*1024) // 10MB max upload

	return router, handlers
}

// TestGetProjects_EmptyDatabase проверяет возврат пустого списка
func TestGetProjects_EmptyDatabase(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/projects", h.GetProjects)

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["total"])

	projects := response["projects"].([]interface{})
	assert.Equal(t, 0, len(projects))
}

// TestGetProjects_WithData проверяет возврат проектов из БД
func TestGetProjects_WithData(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/projects", h.GetProjects)

	// Создаем тестовый проект
	project := models.Project{
		Title:       "Тестовый проект",
		Description: "Описание тестового проекта",
		Slug:        "test-project",
	}
	h.db.Create(&project)

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["total"])

	projects := response["projects"].([]interface{})
	assert.Equal(t, 1, len(projects))

	firstProject := projects[0].(map[string]interface{})
	assert.Equal(t, "Тестовый проект", firstProject["title"])
}

// TestGetProjects_Pagination проверяет пагинацию
func TestGetProjects_Pagination(t *testing.T) {
	router, h := setupTestRouter(t)
	router.GET("/api/projects", h.GetProjects)

	// Создаем 15 проектов
	for i := 1; i <= 15; i++ {
		project := models.Project{
			Title: "Проект " + string(rune(i)),
			Slug:  "project-" + string(rune(i)),
		}
		h.db.Create(&project)
	}

	// Запрашиваем первую страницу (лимит 10)
	req, _ := http.NewRequest("GET", "/api/projects?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(15), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(10), response["limit"])

	projects := response["projects"].([]interface{})
	assert.Equal(t, 10, len(projects))
}

// TestSubmitContact_Success проверяет успешную отправку заявки
func TestSubmitContact_Success(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/contact", h.SubmitContact)

	contactData := map[string]string{
		"name":    "Иван Иванов",
		"phone":   "+79211234567",
		"email":   "ivan@example.com",
		"company": "ООО Тест",
		"message": "Хочу заказать LED экран",
	}

	jsonData, _ := json.Marshal(contactData)
	req, _ := http.NewRequest("POST", "/api/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "успешно отправлена")

	// Проверяем что заявка сохранилась в БД
	var contact models.ContactForm
	h.db.First(&contact)
	assert.Equal(t, "Иван Иванов", contact.Name)
	assert.Equal(t, "+79211234567", contact.Phone)
	assert.Equal(t, "new", contact.Status)
}

// TestSubmitContact_MissingRequiredFields проверяет валидацию
func TestSubmitContact_MissingRequiredFields(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/contact", h.SubmitContact)

	// Отправляем без обязательных полей
	contactData := map[string]string{
		"email": "test@example.com",
	}

	jsonData, _ := json.Marshal(contactData)
	req, _ := http.NewRequest("POST", "/api/contact", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "обязательны")
}

// TestTrackProjectView проверяет трекинг просмотров
func TestTrackProjectView(t *testing.T) {
	t.Skip("Skipping TestTrackProjectView: SQLite doesn't support PostgreSQL ON CONFLICT syntax")

	router, h := setupTestRouter(t)
	router.POST("/api/track/project-view/:id", h.TrackProjectView)

	// Создаем проект
	project := models.Project{
		Title: "Проект для трекинга",
		Slug:  "track-project",
	}
	h.db.Create(&project)

	// Трекаем просмотр
	req, _ := http.NewRequest("POST", "/api/track/project-view/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем что запись создалась
	var viewRecord models.ProjectViewDaily
	h.db.Where("project_id = ?", project.ID).First(&viewRecord)
	assert.Equal(t, uint(1), viewRecord.ProjectID)
	assert.Equal(t, int64(1), viewRecord.Views)
}

// TestTrackProjectView_InvalidID проверяет обработку невалидного ID
func TestTrackProjectView_InvalidID(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/track/project-view/:id", h.TrackProjectView)

	req, _ := http.NewRequest("POST", "/api/track/project-view/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid")
}

// TestTrackProjectView_NotFound проверяет несуществующий проект
func TestTrackProjectView_NotFound(t *testing.T) {
	router, h := setupTestRouter(t)
	router.POST("/api/track/project-view/:id", h.TrackProjectView)

	req, _ := http.NewRequest("POST", "/api/track/project-view/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not found")
}
