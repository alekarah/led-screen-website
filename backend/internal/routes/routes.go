package routes

import (
	"ledsite/internal/handlers"
	"ledsite/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, h *handlers.Handlers) {
	// Глобальные middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Healthcheck
	router.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	// Публичные страницы
	router.GET("/", h.HomePage)
	router.GET("/projects", h.ProjectsPage)
	router.GET("/services", h.ServicesPage)
	router.GET("/contact", h.ContactPage)
	router.GET("/privacy", h.PrivacyPage)

	// API (публичное)
	api := router.Group("/api")
	{
		api.GET("/projects", h.GetProjects)
		api.POST("/contact", h.SubmitContact)
		api.GET("/admin/contacts-7d", h.AdminContacts7Days)
		api.POST("/track/project-view/:id", h.TrackProjectView)
	}

	// Публичные роуты админки (без авторизации)
	router.GET("/admin/login", h.ShowLoginPage)
	router.POST("/admin/login", h.Login)

	// Админка (защищённые роуты)
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware()) // Применяем middleware для всех роутов ниже
	{
		admin.GET("/", h.AdminDashboard)
		admin.GET("/logout", h.Logout)

		// Проекты
		pr := admin.Group("/projects")
		{
			pr.GET("", h.AdminProjects)
			pr.POST("", h.CreateProject)
			pr.GET("/:id", h.GetProject)
			pr.POST("/:id/update", h.UpdateProject)
			pr.DELETE("/:id", h.DeleteProject)
			pr.POST("/:id/reorder", h.ReorderProject)
			pr.POST("/:id/reset-views", h.ResetProjectViews)
			pr.POST("/bulk-reorder", h.BulkReorderProjects)
			pr.POST("/reset-order", h.ResetProjectOrder)
		}

		// Изображения
		img := admin.Group("/")
		{
			img.POST("upload-images", h.UploadImages)
			img.DELETE("images/:id", h.DeleteImage)
			img.POST("images/:id/crop", h.UpdateImageCrop)
			img.POST("analytics/reset", h.ResetAllViews)
		}

		// Заявки (контакты)
		ct := admin.Group("/contacts")
		{
			ct.GET("", h.AdminContactsPage)
			ct.GET("/archive", h.AdminContactsArchivePage)
			ct.GET("/export.csv", h.AdminContactsExportCSV)
			ct.POST("/bulk", h.BulkUpdateContacts)
			ct.POST("/:id/status", h.UpdateContactStatus)
			ct.PATCH("/:id/archive", h.ArchiveContact)
			ct.PATCH("/:id/restore", h.RestoreContact)
			ct.DELETE("/:id", h.DeleteContact)

			// Заметки и напоминания
			ct.GET("/:id/notes", h.GetContactNotes)
			ct.POST("/:id/notes", h.CreateContactNote)
			ct.DELETE("/:id/notes/:note_id", h.DeleteContactNote)
			ct.PATCH("/:id/reminder", h.UpdateContactReminder)
		}
	}
}
