package routes

import (
	"fmt"
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

	// Админ панель
	admin := router.Group("/admin")
	{
		admin.GET("/", h.AdminDashboard)
		admin.GET("/projects", h.AdminProjects)
		admin.POST("/projects", h.CreateProject)
		admin.GET("/projects/:id", h.GetProject)
		admin.POST("/projects/:id/update", h.UpdateProject)
		admin.DELETE("/projects/:id", h.DeleteProject)
		admin.POST("/upload-images", h.UploadImages)
		admin.DELETE("/images/:id", h.DeleteImage)
		admin.POST("/images/:id/crop", h.UpdateImageCrop)
	}

	// Отладочная информация - ПРИНУДИТЕЛЬНО используем fmt
	fmt.Printf("🔧 Маршрут кроппинга зарегистрирован: POST /admin/images/:id/crop\n")
}
