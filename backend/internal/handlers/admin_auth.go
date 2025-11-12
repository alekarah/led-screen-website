package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ledsite/internal/models"
)

// LoginPageData - данные для страницы входа
type LoginPageData struct {
	Error string
}

// LoginRequest - данные из формы входа
type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// JWTClaims - структура для JWT токена
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ShowLoginPage отображает страницу входа
func (h *Handlers) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_login.html", LoginPageData{})
}

// Login обрабатывает вход администратора
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "admin_login.html", LoginPageData{
			Error: "Пожалуйста, заполните все поля",
		})
		return
	}

	// Ищем пользователя в базе данных
	var admin models.Admin
	if err := h.db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "admin_login.html", LoginPageData{
			Error: "Неверное имя пользователя или пароль",
		})
		return
	}

	// Проверяем активность пользователя
	if !admin.IsActive {
		c.HTML(http.StatusUnauthorized, "admin_login.html", LoginPageData{
			Error: "Аккаунт деактивирован",
		})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.HTML(http.StatusUnauthorized, "admin_login.html", LoginPageData{
			Error: "Неверное имя пользователя или пароль",
		})
		return
	}

	// Обновляем время последнего входа
	now := time.Now()
	admin.LastLoginAt = &now
	h.db.Save(&admin)

	// Создаём JWT токен
	token, err := generateJWT(admin.ID, admin.Username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin_login.html", LoginPageData{
			Error: "Ошибка создания сессии",
		})
		return
	}

	// Устанавливаем cookie с токеном
	// Secure флаг включается автоматически в production для защиты токена через HTTPS
	secure := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie(
		"admin_token", // название cookie
		token,         // значение токена
		3600*24*7,     // время жизни 7 дней (в секундах)
		"/",           // путь
		"",            // домен (пустая строка = текущий домен)
		secure,        // secure (автоматически true в production)
		true,          // httpOnly (защита от XSS)
	)

	// Перенаправляем на главную страницу админки
	c.Redirect(http.StatusFound, "/admin")
}

// Logout выход из системы
func (h *Handlers) Logout(c *gin.Context) {
	// Удаляем cookie
	c.SetCookie(
		"admin_token",
		"",
		-1, // время жизни -1 = удалить
		"/",
		"",
		false,
		true,
	)

	// Перенаправляем на страницу входа
	c.Redirect(http.StatusFound, "/admin/login")
}

// generateJWT создаёт JWT токен для пользователя
func generateJWT(userID uint, username string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production" // fallback для разработки
	}

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 дней
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ledsite-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT проверяет JWT токен и возвращает claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
