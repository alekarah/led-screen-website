// Package middleware содержит HTTP middleware функции для обработки запросов.
//
// Middleware выполняются до основных handlers и используются для:
//   - Аутентификации и авторизации (AuthMiddleware)
//   - Логирования запросов (встроенный gin.Logger)
//   - Обработки паник (встроенный gin.Recovery)
//   - CORS (опционально)
//   - Rate limiting (опционально)
package middleware

import (
	"net/http"

	"ledsite/internal/handlers"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет наличие и валидность JWT токена для защищённых роутов.
//
// Выполняет следующие проверки:
//  1. Извлекает JWT токен из HTTP-only cookie "admin_token"
//  2. Валидирует токен (подпись, срок действия)
//  3. Извлекает claims (admin_id, admin_username)
//  4. Сохраняет данные админа в gin.Context для использования в handlers
//
// При ошибке:
//   - Удаляет невалидный cookie
//   - Перенаправляет на /admin/login
//   - Вызывает c.Abort() для прерывания цепочки обработчиков
//
// Используется для всех роутов под /admin/* (кроме /admin/login).
//
// Пример использования:
//
//	admin := router.Group("/admin")
//	admin.Use(middleware.AuthMiddleware())
//	{
//	    admin.GET("/", h.AdminDashboard)
//	}
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
