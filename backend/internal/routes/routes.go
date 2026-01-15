// Package routes содержит настройку всех HTTP маршрутов приложения.
//
// Организация маршрутов:
//   - Публичные страницы (GET /, /projects, /services, /contact, /privacy)
//   - Публичные API (/api/projects, /api/contact, /api/track/*)
//   - Админ страницы без защиты (GET/POST /admin/login)
//   - Защищённые админ роуты (/admin/* с JWT middleware)
//
// Все админские роуты (кроме login) защищены AuthMiddleware,
// который проверяет JWT токен из HTTP-only cookie.
package routes

import (
	"ledsite/internal/handlers"
	"ledsite/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Setup настраивает все маршруты приложения на переданном Gin router.
//
// Параметры:
//   - router: экземпляр gin.Engine для регистрации маршрутов (создан через gin.Default())
//   - h: handlers с внедрённой зависимостью базы данных
//
// Группы маршрутов:
//   - Публичные страницы (без защиты)
//   - /api - публичные API эндпоинты
//   - /admin/login - страница входа (без защиты)
//   - /admin/* - защищённые админские роуты (JWT required)
//
// Вызывается из main.go:
//
//	router := gin.Default()
//	routes.Setup(router, handlers)
//	router.Run(":8080")
func Setup(router *gin.Engine, h *handlers.Handlers) {
	// Примечание: gin.Default() в main.go уже включает Logger и Recovery middleware

	// Healthcheck endpoint для мониторинга (Kubernetes, Docker, etc.)
	router.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	// SEO endpoints для поисковых систем
	router.GET("/sitemap.xml", h.Sitemap)
	router.GET("/robots.txt", h.RobotsTxt)

	// Публичные страницы
	router.GET("/", h.HomePage)
	router.GET("/projects", h.ProjectsPage)
	router.GET("/services", h.ServicesPage)
	router.GET("/led-screens-guide", h.LEDGuidePage)
	router.GET("/prices", h.PricesPage)
	router.GET("/contact", h.ContactPage)
	router.GET("/privacy", h.PrivacyPage)

	// API (публичное) - JSON эндпоинты без авторизации
	api := router.Group("/api")
	{
		api.GET("/projects", h.GetProjects)                     // Список проектов с пагинацией
		api.POST("/contact", h.SubmitContact)                   // Отправка заявки от клиента
		api.GET("/admin/contacts-7d", h.AdminContacts7Days)     // Статистика заявок за 7 дней (для dashboard)
		api.POST("/track/project-view/:id", h.TrackProjectView) // Трекинг просмотров проекта
		api.POST("/track/price-view/:id", h.TrackPriceView)     // Трекинг просмотров позиции прайса

		// Telegram Bot API - доступ только с localhost (127.0.0.1)
		telegram := api.Group("/telegram")
		{
			telegram.POST("/update-status", h.TelegramUpdateStatus)          // Изменить статус заявки
			telegram.POST("/add-note", h.TelegramAddNote)                    // Добавить заметку
			telegram.POST("/set-reminder", h.TelegramSetReminder)            // Установить напоминание
			telegram.GET("/due-reminders", h.TelegramGetDueReminders)        // Получить напоминания к отправке
			telegram.POST("/mark-reminder-sent", h.TelegramMarkReminderSent) // Пометить напоминание как отправленное
		}
	}

	// Публичные роуты админки (без авторизации)
	router.GET("/admin/login", h.ShowLoginPage)
	router.POST("/admin/login", h.Login)

	// Админка (защищённые роуты) - требуют валидный JWT токен
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware()) // JWT проверка для всех роутов ниже
	{
		admin.GET("/", h.AdminDashboard) // Главная страница админки с аналитикой
		admin.GET("/logout", h.Logout)   // Выход (удаление JWT cookie)

		// Проекты - CRUD операции и управление порядком
		pr := admin.Group("/projects")
		{
			pr.GET("", h.AdminProjects)                      // Страница управления проектами
			pr.POST("", h.CreateProject)                     // Создание нового проекта
			pr.GET("/:id", h.GetProject)                     // Получение проекта для редактирования (JSON)
			pr.POST("/:id/update", h.UpdateProject)          // Обновление проекта
			pr.DELETE("/:id", h.DeleteProject)               // Удаление проекта с изображениями
			pr.POST("/:id/reorder", h.ReorderProject)        // Изменение порядка одного проекта
			pr.POST("/:id/reset-views", h.ResetProjectViews) // Сброс просмотров конкретного проекта
			pr.POST("/bulk-reorder", h.BulkReorderProjects)  // Массовая сортировка (drag & drop)
		}

		// Изображения - загрузка, удаление, кроппинг
		img := admin.Group("/")
		{
			img.POST("upload-images", h.UploadImages)                // Загрузка изображений для проекта
			img.DELETE("images/:id", h.DeleteImage)                  // Удаление изображения
			img.POST("images/:id/crop", h.UpdateImageCrop)           // Обновление настроек кроппинга
			img.POST("images/:id/set-primary", h.SetPrimaryImage)    // Установка главного изображения проекта
			img.POST("analytics/reset", h.ResetAllViews)             // Глобальный сброс статистики просмотров проектов
			img.POST("analytics/reset-prices", h.ResetAllPriceViews) // Глобальный сброс статистики просмотров прайса
		}

		// Заявки (контакты) - CRM система
		ct := admin.Group("/contacts")
		{
			ct.GET("", h.AdminContactsPage)                 // Страница активных заявок (с фильтрами)
			ct.GET("/archive", h.AdminContactsArchivePage)  // Страница архива заявок
			ct.GET("/export.csv", h.AdminContactsExportCSV) // Экспорт в CSV (UTF-8 BOM)
			ct.POST("/bulk", h.BulkUpdateContacts)          // Массовое изменение статуса
			ct.POST("/:id/status", h.UpdateContactStatus)   // Изменение статуса одной заявки
			ct.PATCH("/:id/archive", h.ArchiveContact)      // Архивирование заявки
			ct.PATCH("/:id/restore", h.RestoreContact)      // Восстановление из архива
			ct.DELETE("/:id", h.DeleteContact)              // Удаление (soft или hard)

			// Заметки и напоминания для follow-up
			ct.GET("/:id/notes", h.GetContactNotes)               // Получить все заметки по заявке
			ct.POST("/:id/notes", h.CreateContactNote)            // Добавить заметку
			ct.DELETE("/:id/notes/:note_id", h.DeleteContactNote) // Удалить заметку
			ct.PATCH("/:id/reminder", h.UpdateContactReminder)    // Установить/снять напоминание
		}

		// Цены - CRUD операции для прайс-листа
		prices := admin.Group("/prices")
		{
			prices.GET("", h.AdminPricesPage)                // Страница управления ценами
			prices.POST("", h.CreatePriceItem)               // Создание новой позиции прайса
			prices.GET("/:id", h.GetPriceItem)               // Получение позиции для редактирования (JSON)
			prices.POST("/:id/update", h.UpdatePriceItem)    // Обновление позиции прайса
			prices.DELETE("/:id", h.DeletePriceItem)         // Удаление позиции прайса
			prices.POST("/sort", h.UpdatePriceItemsSorting)  // Обновление порядка позиций (drag & drop)
			prices.POST("/:id/crop", h.UpdatePriceImageCrop) // Обновление настроек кроппинга изображения (старый формат)
			prices.DELETE("/:id/image", h.DeletePriceImage)  // Удаление изображения из позиции (старый формат)

			// Новые endpoints для множественных изображений
			prices.POST("/upload-images", h.UploadPriceImages)             // Загрузка изображений для позиции
			prices.DELETE("/images/:id", h.DeletePriceImageNew)            // Удаление изображения из price_images
			prices.POST("/images/:id/crop", h.UpdatePriceImageCropNew)     // Обновление настроек кроппинга изображения
			prices.POST("/images/:id/set-primary", h.SetPrimaryPriceImage) // Установка главного изображения позиции
			prices.POST("/:id/reset-views", h.ResetPriceItemViews)         // Сброс просмотров конкретной позиции прайса
		}
	}
}
