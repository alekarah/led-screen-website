// Package handlers содержит HTTP обработчики для всех маршрутов приложения.
//
// Включает в себя:
//   - Публичные страницы (главная, портфолио, услуги, контакты)
//   - Публичные API (получение проектов, отправка заявок, трекинг просмотров)
//   - Административные страницы и API (защищены JWT)
//
// Все handlers работают с GORM для доступа к базе данных и используют
// Gin Context для обработки HTTP запросов и ответов.
package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ledsite/internal/models"
)

// Handlers содержит зависимости для HTTP обработчиков.
// Все методы handlers используют внедренное подключение к базе данных.
type Handlers struct {
	db            *gorm.DB // Подключение к PostgreSQL через GORM
	maxUploadSize int64    // Максимальный размер загружаемого файла в байтах
	uploadPath    string   // Путь для сохранения загружаемых файлов
}

// New создает новый экземпляр Handlers с внедренной зависимостью базы данных.
// Это центральная точка инициализации всех обработчиков приложения.
//
// Пример использования:
//
//	db, _ := database.Connect(cfg)
//	handlers := handlers.New(db, cfg.MaxUploadSize, cfg.UploadPath)
//	routes.Setup(router, handlers)
func New(db *gorm.DB, maxUploadSize int64, uploadPath string) *Handlers {
	return &Handlers{
		db:            db,
		maxUploadSize: maxUploadSize,
		uploadPath:    uploadPath,
	}
}

// HomePage рендерит главную страницу сайта с избранными проектами и услугами.
//
// Отображает:
//   - До 6 проектов с флагом featured=true (отсортированных по sort_order)
//   - Если нет featured проектов - показывает последние 6 проектов
//   - Основные услуги (featured services)
//
// GET /
func (h *Handlers) HomePage(c *gin.Context) {
	// Получаем проекты для главной страницы (только те, что помечены для показа)
	var featuredProjects []models.Project
	h.db.Where("featured = ?", true).
		Preload("Categories").
		Preload("Images").
		Order("sort_order ASC, created_at DESC").
		Limit(6).
		Find(&featuredProjects)

	// Если нет рекомендуемых проектов, показываем последние
	if len(featuredProjects) == 0 {
		h.db.Preload("Categories").
			Preload("Images").
			Order("sort_order ASC, created_at DESC").
			Limit(6).
			Find(&featuredProjects)
	}

	// Получаем основные услуги
	var services []models.Service
	h.db.Where("featured = ?", true).
		Order("sort_order").
		Find(&services)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":         "LED экраны в СПб | Service 'n' Repair",
		"description":   "Поставка, монтаж и обслуживание LED дисплеев для бизнеса в Санкт-Петербурге. Портфолио проектов. Гарантия качества.",
		"ogTitle":       "S'n'R - Продажа и обслуживание LED дисплеев",
		"ogDescription": "Поставка, монтаж и обслуживание LED дисплеев для бизнеса в Санкт-Петербурге. Портфолио проектов. Гарантия качества.",
		"ogUrl":         "/",
		"ogImage":       "https://s-n-r.ru/static/images/og-preview.png",
		"projects":      featuredProjects,
		"services":      services,
		"PageID":        "home",
	})
}

// ProjectsPage отображает страницу портфолио с возможностью фильтрации по категориям.
//
// Query параметры:
//   - category (string): slug категории для фильтрации проектов
//
// Проекты отсортированы по sort_order (ASC), затем по created_at (DESC).
//
// GET /projects?category=shopping-centers
func (h *Handlers) ProjectsPage(c *gin.Context) {
	// Получаем параметр фильтрации
	categorySlug := c.Query("category")

	query := h.db.Preload("Categories").Preload("Images")

	if categorySlug != "" {
		query = query.Joins("JOIN project_categories pc ON projects.id = pc.project_id").
			Joins("JOIN categories cat ON pc.category_id = cat.id").
			Where("cat.slug = ?", categorySlug)
	}

	var projects []models.Project
	query.Order("sort_order ASC, created_at DESC").Find(&projects)

	var categories []models.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":          "Портфолио - LED Display",
		"description":    "Портфолио реализованных проектов LED дисплеев и экранов в Санкт-Петербурге. Торговые центры, АЗС, рекламные щиты.",
		"ogTitle":        "Портфолио проектов LED дисплеев | S'n'R",
		"ogDescription":  "Портфолио реализованных проектов LED дисплеев и экранов в Санкт-Петербурге. Торговые центры, АЗС, рекламные щиты.",
		"ogUrl":          "/projects",
		"projects":       projects,
		"categories":     categories,
		"activeCategory": categorySlug,
		"PageID":         "projects",
	})
}

// ServicesPage рендерит страницу со списком всех услуг компании.
// Услуги отсортированы по полю sort_order.
//
// GET /services
func (h *Handlers) ServicesPage(c *gin.Context) {
	var services []models.Service
	h.db.Order("sort_order").Find(&services)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":         "Услуги | LED экраны",
		"description":   "Услуги по продаже, монтажу и обслуживанию LED дисплеев в СПб. Интерьерные и уличные экраны, металлоконструкции.",
		"ogTitle":       "Услуги LED дисплеев | S'n'R",
		"ogDescription": "Услуги по продаже, монтажу и обслуживанию LED дисплеев в СПб. Интерьерные и уличные экраны, металлоконструкции.",
		"ogUrl":         "/services",
		"services":      services,
		"PageID":        "services",
	})
}

// LEDGuidePage рендерит информационную страницу о LED экранах.
//
// GET /led-screens-guide
func (h *Handlers) LEDGuidePage(c *gin.Context) {
	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":         "Всё о LED-экранах: виды, выбор и частые вопросы | S'n'R",
		"description":   "Узнайте всё о LED-экранах: различия уличных и интерьерных моделей, шаг пикселя, срок службы, советы по установке и ответы на частые вопросы.",
		"ogTitle":       "LED-экраны — виды, выбор, частые вопросы | S'n'R Санкт-Петербург",
		"ogDescription": "Полное руководство по выбору LED-экранов: типы, характеристики, сравнение, преимущества и ответы на частые вопросы.",
		"ogUrl":         "/led-screens-guide",
		"PageID":        "led-guide",
	})
}

// ContactPage рендерит страницу контактов с формой обратной связи и картой.
//
// GET /contact
func (h *Handlers) ContactPage(c *gin.Context) {
	// Загружаем активные точки для карты
	var mapPoints []models.MapPoint
	h.db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&mapPoints)

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":         "Контакты | LED экраны",
		"description":   "Свяжитесь с нами для консультации по LED дисплеям. Телефон, email, форма обратной связи.",
		"ogTitle":       "Контакты | S'n'R",
		"ogDescription": "Свяжитесь с нами для консультации по LED дисплеям. Телефон, email, форма обратной связи.",
		"ogUrl":         "/contact",
		"PageID":        "contact",
		"mapPoints":     mapPoints,
	})
}

// PrivacyPage рендерит страницу политики обработки персональных данных.
//
// GET /privacy
func (h *Handlers) PrivacyPage(c *gin.Context) {
	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":  "Обработка персональных данных",
		"PageID": "privacy",
		"ogUrl":  "/privacy",
	})
}

// ==================== Публичные API ====================

// GetProjects возвращает список проектов в формате JSON с поддержкой пагинации и фильтрации.
//
// Query параметры:
//   - page (int): номер страницы (по умолчанию: 1)
//   - limit (int): количество проектов на странице (по умолчанию: 12)
//   - category (string): slug категории для фильтрации
//
// Ответ включает:
//   - projects: массив проектов с категориями и изображениями
//   - total: общее количество проектов
//   - page, limit: параметры пагинации
//
// GET /api/projects?page=1&limit=12&category=shopping-centers
func (h *Handlers) GetProjects(c *gin.Context) {
	var projects []models.Project

	// Параметры пагинации
	page := 1
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	limit := 12
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "12")); err == nil && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit

	// Фильтрация по категории
	category := c.Query("category")

	query := h.db.Preload("Categories").Preload("Images")

	if category != "" {
		query = query.Joins("JOIN project_categories pc ON projects.id = pc.project_id").
			Joins("JOIN categories cat ON pc.category_id = cat.id").
			Where("cat.slug = ?", category)
	}

	var total int64
	query.Model(&models.Project{}).Count(&total)

	query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects)

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// SubmitContact обрабатывает отправку формы обратной связи от клиентов.
//
// Принимает данные в формате JSON или form-data:
//   - name (required): имя клиента
//   - phone (required): телефон
//   - email (optional): email
//   - company (optional): название компании
//   - project_type (optional): тип проекта
//   - message (optional): сообщение
//
// Статус заявки по умолчанию устанавливается в "new".
//
// POST /api/contact
func (h *Handlers) SubmitContact(c *gin.Context) {
	var form models.ContactForm

	// Проверяем Content-Type и парсим данные соответственно
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверные данные формы",
			})
			return
		}
	} else {
		// Парсим данные формы
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверные данные формы",
			})
			return
		}
	}

	// Honeypot защита от спама: если скрытое поле заполнено - это бот
	if form.Website != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные формы",
		})
		return
	}

	// Простая валидация
	if form.Name == "" || form.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Имя и телефон обязательны для заполнения",
		})
		return
	}

	// Сохраняем в базу
	if err := h.db.Create(&form).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка сохранения заявки",
		})
		return
	}

	// Отправляем уведомление в Telegram (асинхронно, не блокируем ответ пользователю)
	SendTelegramNotificationAsync(&form)

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно отправлена! Мы свяжемся с вами в ближайшее время.",
	})
}

// TrackProjectView увеличивает счетчик просмотров проекта за текущий день (по МСК).
//
// Использует UPSERT (INSERT ... ON CONFLICT) для атомарного инкремента счетчика.
// Если запись за сегодняшний день уже существует - увеличивает views на 1,
// иначе создает новую запись с views=1.
//
// Параметры:
//   - id (path): ID проекта
//
// Просмотры агрегируются по дням в таблице project_view_dailies.
//
// POST /api/track/project-view/:id
func (h *Handlers) TrackProjectView(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("id"))
	if err != nil || pid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var exists int64
	if err := h.db.Model(&models.Project{}).Where("id = ?", pid).Count(&exists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Получаем текущую дату по Московскому времени, обнуляем время до полуночи
	now := NowMSK()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)

	rec := models.ProjectViewDaily{
		ProjectID: uint(pid),
		Day:       day,
		Views:     1,
	}

	// UPSERT: если запись (project_id, day) существует - инкрементируем views,
	// иначе создаем новую запись с views=1
	if err := h.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "project_id"},
			{Name: "day"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"views": gorm.Expr(`"project_view_dailies"."views" + EXCLUDED."views"`),
		}),
	}).Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db upsert error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// TrackPriceView увеличивает счетчик просмотров позиции прайса за текущий день (по МСК).
//
// Использует UPSERT (INSERT ... ON CONFLICT) для атомарного инкремента счетчика.
// Если запись за сегодняшний день уже существует - увеличивает views на 1,
// иначе создает новую запись с views=1.
//
// Параметры:
//   - id (path): ID позиции прайса
//
// Просмотры агрегируются по дням в таблице price_view_dailies.
//
// POST /api/track/price-view/:id
func (h *Handlers) TrackPriceView(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("id"))
	if err != nil || pid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price item id"})
		return
	}

	var exists int64
	if err := h.db.Model(&models.PriceItem{}).Where("id = ?", pid).Count(&exists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "price item not found"})
		return
	}

	// Получаем текущую дату по Московскому времени, обнуляем время до полуночи
	now := NowMSK()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLoc)

	rec := models.PriceViewDaily{
		PriceItemID: uint(pid),
		Day:         day,
		Views:       1,
	}

	// UPSERT: если запись (price_item_id, day) существует - инкрементируем views,
	// иначе создаем новую запись с views=1
	if err := h.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "price_item_id"},
			{Name: "day"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"views": gorm.Expr(`"price_view_dailies"."views" + EXCLUDED."views"`),
		}),
	}).Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db upsert error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ==================== Утилитные функции ====================

// generateSlug создает URL-friendly slug из русского или английского названия.
//
// Выполняет:
//   - Приведение к нижнему регистру
//   - Замену пробелов на дефисы
//   - Транслитерацию русских букв в латиницу
//   - Добавление timestamp для гарантии уникальности
//
// Пример: "LED экран на ТЦ Мега" -> "led-ekran-na-tc-mega-1234567890"
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "ё", "e")
	slug = strings.ReplaceAll(slug, "а", "a")
	slug = strings.ReplaceAll(slug, "б", "b")
	slug = strings.ReplaceAll(slug, "в", "v")
	slug = strings.ReplaceAll(slug, "г", "g")
	slug = strings.ReplaceAll(slug, "д", "d")
	slug = strings.ReplaceAll(slug, "е", "e")
	slug = strings.ReplaceAll(slug, "ж", "zh")
	slug = strings.ReplaceAll(slug, "з", "z")
	slug = strings.ReplaceAll(slug, "и", "i")
	slug = strings.ReplaceAll(slug, "й", "y")
	slug = strings.ReplaceAll(slug, "к", "k")
	slug = strings.ReplaceAll(slug, "л", "l")
	slug = strings.ReplaceAll(slug, "м", "m")
	slug = strings.ReplaceAll(slug, "н", "n")
	slug = strings.ReplaceAll(slug, "о", "o")
	slug = strings.ReplaceAll(slug, "п", "p")
	slug = strings.ReplaceAll(slug, "р", "r")
	slug = strings.ReplaceAll(slug, "с", "s")
	slug = strings.ReplaceAll(slug, "т", "t")
	slug = strings.ReplaceAll(slug, "у", "u")
	slug = strings.ReplaceAll(slug, "ф", "f")
	slug = strings.ReplaceAll(slug, "х", "h")
	slug = strings.ReplaceAll(slug, "ц", "ts")
	slug = strings.ReplaceAll(slug, "ч", "ch")
	slug = strings.ReplaceAll(slug, "ш", "sh")
	slug = strings.ReplaceAll(slug, "щ", "sch")
	slug = strings.ReplaceAll(slug, "ъ", "")
	slug = strings.ReplaceAll(slug, "ы", "y")
	slug = strings.ReplaceAll(slug, "ь", "")
	slug = strings.ReplaceAll(slug, "э", "e")
	slug = strings.ReplaceAll(slug, "ю", "yu")
	slug = strings.ReplaceAll(slug, "я", "ya")

	// Добавляем Unix timestamp для гарантии уникальности slug
	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}

// isImageFile проверяет, является ли файл допустимым форматом изображения.
//
// Поддерживаемые форматы: .jpg, .jpeg, .jfif, .png, .gif, .webp
//
// Используется для валидации загружаемых файлов проектов.
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".jfif", ".png", ".gif", ".webp"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

// PricesPage рендерит публичную страницу с ценами на LED экраны и услуги.
//
// Отображает:
//   - Активные позиции прайса (is_active = true)
//   - Карточки с изображениями, названиями, ценами
//   - Раскрывающиеся таблицы характеристик (если has_specifications = true)
//
// Позиции отсортированы по sort_order (ASC), затем по created_at (DESC).
//
// GET /prices
func (h *Handlers) PricesPage(c *gin.Context) {
	var priceItems []models.PriceItem

	// Получаем только активные позиции с характеристиками и изображениями
	h.db.Where("is_active = ?", true).
		Preload("Specifications", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, created_at DESC").
		Find(&priceItems)

	// Группируем характеристики по группам для каждой позиции
	type GroupedSpec struct {
		Group string
		Specs []models.PriceSpecification
	}

	priceItemsWithGroupedSpecs := make([]struct {
		PriceItem    models.PriceItem
		GroupedSpecs []GroupedSpec
	}, len(priceItems))

	for i, item := range priceItems {
		priceItemsWithGroupedSpecs[i].PriceItem = item

		if item.HasSpecifications && len(item.Specifications) > 0 {
			groupsMap := make(map[string][]models.PriceSpecification)
			groupsOrder := []string{} // Сохраняем порядок групп

			// Группируем спецификации, сохраняя порядок первого появления группы
			for _, spec := range item.Specifications {
				// Если группа встречается первый раз, добавляем в порядок
				if _, exists := groupsMap[spec.SpecGroup]; !exists {
					groupsOrder = append(groupsOrder, spec.SpecGroup)
				}
				groupsMap[spec.SpecGroup] = append(groupsMap[spec.SpecGroup], spec)
			}

			// Преобразуем map в slice в порядке первого появления групп
			for _, group := range groupsOrder {
				specs := groupsMap[group]
				priceItemsWithGroupedSpecs[i].GroupedSpecs = append(
					priceItemsWithGroupedSpecs[i].GroupedSpecs,
					GroupedSpec{Group: group, Specs: specs},
				)
			}
		}
	}

	c.HTML(http.StatusOK, "public_base.html", gin.H{
		"title":         "Цены на LED экраны | S'n'R",
		"description":   "Цены на LED экраны и дисплеи в Санкт-Петербурге. Прайс-лист на уличные, интерьерные экраны, медиафасады.",
		"ogTitle":       "Цены на LED экраны и дисплеи | S'n'R",
		"ogDescription": "Цены на LED экраны и дисплеи в Санкт-Петербурге. Прайс-лист на уличные, интерьерные экраны, медиафасады.",
		"ogUrl":         "/prices",
		"priceItems":    priceItemsWithGroupedSpecs,
		"PageID":        "prices",
	})
}
