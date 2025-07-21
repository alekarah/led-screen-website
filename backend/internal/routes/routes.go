package routes

import (
	"ledsite/internal/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, h *handlers.Handlers) {
	// Middleware для логирования
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Главная страница
	router.GET("/", h.HomePage)

	// Страницы сайта
	router.GET("/projects", h.ProjectsPage)
	router.GET("/projects/:slug", h.ProjectDetail)
	router.GET("/services", h.ServicesPage)
	router.GET("/contact", h.ContactPage)

	// API маршруты
	api := router.Group("/api")
	{
		api.GET("/projects", h.GetProjects)
		api.POST("/contact", h.SubmitContact)
	}

	// Админ панель (пока простая)
	admin := router.Group("/admin")
	{
		admin.GET("/", func(c *gin.Context) {
			c.HTML(200, "admin.html", gin.H{
				"title": "Админ панель",
			})
		})
	}
}
