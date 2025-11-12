package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter создает тестовый Gin router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// JWTClaims - дублируем структуру из handlers для тестов
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// createTestToken создает валидный JWT токен для тестов
func createTestToken(adminID uint, expiresIn time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key-for-testing-only"
		os.Setenv("JWT_SECRET", secret)
	}

	claims := JWTClaims{
		UserID:   adminID,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ledsite-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// TestAuthMiddleware_ValidToken проверяет доступ с валидным токеном
func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupTestRouter()

	// Защищенный роут
	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		adminID, _ := c.Get("admin_id")
		c.JSON(http.StatusOK, gin.H{"admin_id": adminID})
	})

	// Создаем валидный токен
	tokenString, err := createTestToken(1, time.Hour)
	assert.NoError(t, err)

	// Создаем запрос с токеном в cookie
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "admin_token",
		Value: tokenString,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_MissingToken проверяет отказ без токена
func TestAuthMiddleware_MissingToken(t *testing.T) {
	router := setupTestRouter()

	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Запрос без токена
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Должен быть редирект на /admin/login
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/login", w.Header().Get("Location"))
}

// TestAuthMiddleware_InvalidToken проверяет обработку невалидного токена
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupTestRouter()

	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Запрос с невалидным токеном
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "admin_token",
		Value: "invalid.token.here",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/login", w.Header().Get("Location"))
}

// TestAuthMiddleware_ExpiredToken проверяет обработку истекшего токена
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	router := setupTestRouter()

	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Создаем токен, который истек 1 час назад
	tokenString, err := createTestToken(1, -time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/admin/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "admin_token",
		Value: tokenString,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/login", w.Header().Get("Location"))
}

// TestAuthMiddleware_AdminIDInContext проверяет, что admin_id добавляется в контекст
func TestAuthMiddleware_AdminIDInContext(t *testing.T) {
	router := setupTestRouter()

	var capturedAdminID uint

	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		adminID, exists := c.Get("admin_id")
		assert.True(t, exists, "admin_id должен быть в контексте")

		capturedAdminID = adminID.(uint)
		c.JSON(http.StatusOK, gin.H{"admin_id": adminID})
	})

	// Создаем токен для админа с ID = 42
	tokenString, err := createTestToken(42, time.Hour)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/admin/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "admin_token",
		Value: tokenString,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(42), capturedAdminID)
}

// TestAuthMiddleware_MultipleRequests проверяет работу с несколькими запросами
func TestAuthMiddleware_MultipleRequests(t *testing.T) {
	router := setupTestRouter()

	router.GET("/admin/test", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tokenString, err := createTestToken(1, time.Hour)
	assert.NoError(t, err)

	// Делаем 5 запросов с одним токеном
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/admin/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "admin_token",
			Value: tokenString,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Запрос %d должен быть успешным", i+1)
	}
}
