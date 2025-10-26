package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ledsite/internal/handlers"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из cookie
		token, err := c.Cookie("admin_token")
		if err != nil {
			// Нет cookie - перенаправляем на страницу входа
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// Проверяем валидность токена
		claims, err := handlers.ValidateJWT(token)
		if err != nil {
			// Токен невалидный - удаляем cookie и перенаправляем
			c.SetCookie("admin_token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// Сохраняем информацию о пользователе в контекст
		c.Set("admin_id", claims.UserID)
		c.Set("admin_username", claims.Username)

		c.Next()
	}
}
