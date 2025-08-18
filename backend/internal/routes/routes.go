package routes

import (
	"ledsite/internal/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, h *handlers.Handlers) {
	// Middleware для логирования
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Healthcheck endpoint (для prod/devops)
	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Главная страница
	router.GET("/", h.HomePage)

	// Страницы сайта
	router.GET("/projects", h.ProjectsPage)
	router.GET("/services", h.ServicesPage)
	router.GET("/contact", h.ContactPage)
	router.GET("/privacy", h.PrivacyPage)

	// API маршруты
	api := router.Group("/api")
	{
		api.GET("/projects", h.GetProjects)
		api.POST("/contact", h.SubmitContact)
		api.GET("/admin/contacts-7d", h.AdminContacts7Days)
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
		admin.POST("/projects/:id/reorder", h.ReorderProject)
		admin.POST("/projects/bulk-reorder", h.BulkReorderProjects)
		admin.POST("/projects/reset-order", h.ResetProjectOrder)
		admin.POST("/upload-images", h.UploadImages)
		admin.DELETE("/images/:id", h.DeleteImage)
		admin.POST("/images/:id/crop", h.UpdateImageCrop)
		admin.GET("/contacts", h.AdminContactsPage)
		admin.POST("/contacts/:id/done", h.MarkContactDone)
	}
}
