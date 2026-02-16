package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"ledsite/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// setupAuthRouter создаёт роутер с минимальным шаблоном для тестов авторизации
func setupAuthRouter(t *testing.T) (*gin.Engine, *Handlers) {
	router, h := setupTestRouter(t)

	// Минимальный шаблон для c.HTML() — без него Gin паникует
	tmpl := template.Must(template.New("admin_login.html").Parse(`{{.Error}}`))
	router.SetHTMLTemplate(tmpl)

	return router, h
}

// createTestAdmin создаёт тестового администратора в БД
func createTestAdmin(t *testing.T, h *Handlers, username, password string, isActive bool) models.Admin {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	assert.NoError(t, err)

	admin := models.Admin{
		Username:     username,
		PasswordHash: string(hash),
		IsActive:     isActive,
	}
	err = h.db.Create(&admin).Error
	assert.NoError(t, err)

	// GORM игнорирует false для bool с default:true, поэтому обновляем явно
	if !isActive {
		h.db.Model(&admin).Update("is_active", false)
	}

	return admin
}

// postLoginForm отправляет POST-запрос на /admin/login с form-data
func postLoginForm(router *gin.Engine, username, password string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	req, _ := http.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------- ShowLoginPage ----------

func TestShowLoginPage(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.GET("/admin/login", h.ShowLoginPage)

	req, _ := http.NewRequest("GET", "/admin/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------- Login ----------

func TestLogin_Success(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "admin", "secret123", true)

	w := postLoginForm(router, "admin", "secret123")

	// Успешный логин → редирект на /admin
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin", w.Header().Get("Location"))

	// Проверяем что cookie установлен
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "admin_token" {
			tokenCookie = c
			break
		}
	}
	assert.NotNil(t, tokenCookie, "cookie admin_token должен быть установлен")
	assert.True(t, tokenCookie.HttpOnly, "cookie должен быть HttpOnly")
	assert.NotEmpty(t, tokenCookie.Value)
}

func TestLogin_UpdatesLastLoginAt(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	admin := createTestAdmin(t, h, "admin", "secret123", true)

	// До логина LastLoginAt == nil
	assert.Nil(t, admin.LastLoginAt)

	postLoginForm(router, "admin", "secret123")

	// После логина LastLoginAt обновился
	var updated models.Admin
	h.db.First(&updated, admin.ID)
	assert.NotNil(t, updated.LastLoginAt)
}

func TestLogin_EmptyFields(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)

	w := postLoginForm(router, "", "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_EmptyUsername(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)

	w := postLoginForm(router, "", "password")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_EmptyPassword(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)

	w := postLoginForm(router, "admin", "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_WrongUsername(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "admin", "secret123", true)

	w := postLoginForm(router, "nonexistent", "secret123")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "admin", "secret123", true)

	w := postLoginForm(router, "admin", "wrongpassword")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_InactiveAdmin(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "inactive", "secret123", false)

	w := postLoginForm(router, "inactive", "secret123")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Cookie не должен быть установлен
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "admin_token" {
			t.Fatal("cookie admin_token не должен быть установлен для неактивного админа")
		}
	}
}

func TestLogin_CookieNotSecureInDev(t *testing.T) {
	os.Unsetenv("ENVIRONMENT")
	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "admin", "secret123", true)

	w := postLoginForm(router, "admin", "secret123")

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "admin_token" {
			assert.False(t, c.Secure, "cookie не должен быть Secure в dev-режиме")
			return
		}
	}
	t.Fatal("cookie admin_token не найден")
}

func TestLogin_CookieSecureInProduction(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	router, h := setupAuthRouter(t)
	router.POST("/admin/login", h.Login)
	createTestAdmin(t, h, "admin", "secret123", true)

	w := postLoginForm(router, "admin", "secret123")

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "admin_token" {
			assert.True(t, c.Secure, "cookie должен быть Secure в production")
			return
		}
	}
	t.Fatal("cookie admin_token не найден")
}

// ---------- Logout ----------

func TestLogout_ClearsCookie(t *testing.T) {
	router, h := setupAuthRouter(t)
	router.GET("/admin/logout", h.Logout)

	req, _ := http.NewRequest("GET", "/admin/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/admin/login", w.Header().Get("Location"))

	// Cookie должен быть удалён (MaxAge < 0)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "admin_token" {
			assert.True(t, c.MaxAge < 0, "cookie должен быть удалён (MaxAge < 0)")
			return
		}
	}
}

// ---------- generateJWT ----------

func TestGenerateJWT_Valid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	tokenStr, err := generateJWT(1, "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// Парсим и проверяем claims
	claims, err := ValidateJWT(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, "ledsite-admin", claims.Issuer)
}

func TestGenerateJWT_FallbackSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	tokenStr, err := generateJWT(1, "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
}

func TestGenerateJWT_ExpiresIn7Days(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	tokenStr, err := generateJWT(1, "admin")
	assert.NoError(t, err)

	claims, err := ValidateJWT(tokenStr)
	assert.NoError(t, err)

	// Проверяем что токен истекает примерно через 7 дней (±1 минута)
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	assert.InDelta(t, expectedExpiry.Unix(), claims.ExpiresAt.Unix(), 60)
}

// ---------- ValidateJWT ----------

func TestValidateJWT_Valid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	tokenStr, _ := generateJWT(42, "testuser")
	claims, err := ValidateJWT(tokenStr)

	assert.NoError(t, err)
	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-one")
	tokenStr, _ := generateJWT(1, "admin")

	// Валидируем с другим секретом
	os.Setenv("JWT_SECRET", "secret-two")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateJWT(tokenStr)
	assert.Error(t, err)
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Создаём токен с истекшим сроком вручную
	claims := JWTClaims{
		UserID:   1,
		Username: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "ledsite-admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret"))

	_, err := ValidateJWT(tokenStr)
	assert.Error(t, err)
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateJWT("not.a.valid.jwt")
	assert.Error(t, err)
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	_, err := ValidateJWT("")
	assert.Error(t, err)
}

// ---------- HashPassword ----------

func TestHashPassword_Success(t *testing.T) {
	hash, err := HashPassword("mypassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword", hash)

	// Проверяем что хеш совпадает с паролем
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("mypassword"))
	assert.NoError(t, err)
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	// Два хеша одного пароля должны отличаться (разные соли)
	hash1, _ := HashPassword("same-password")
	hash2, _ := HashPassword("same-password")
	assert.NotEqual(t, hash1, hash2)
}

func TestHashPassword_WrongPasswordFails(t *testing.T) {
	hash, _ := HashPassword("correct")
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong"))
	assert.Error(t, err)
}
