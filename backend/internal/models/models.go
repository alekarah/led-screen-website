// Package models содержит структуры данных (модели) для работы с базой данных через GORM.
//
// Все модели соответствуют таблицам в PostgreSQL и определяют:
//   - Поля таблиц с типами данных
//   - Связи между таблицами (один-ко-многим, многие-ко-многим)
//   - JSON теги для сериализации в API ответах
//   - GORM теги для маппинга и валидации
//
// Основные сущности:
//   - Category, Project, Image - портфолио проектов
//   - ContactForm, ContactNote - система CRM для заявок
//   - Service - услуги компании
//   - Admin - администраторы системы
//   - ProjectViewDaily - аналитика просмотров
package models

import (
	"time"
)

// Category представляет категорию проектов (интерьерные, уличные, медиафасады и т.д.).
//
// Связи:
//   - many-to-many с Project через таблицу project_categories
//
// Примеры категорий: "Рекламные щиты", "АЗС", "Торговые центры"
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_categories;"`
}

// Project представляет проект из портфолио компании.
//
// Связи:
//   - many-to-many с Category через таблицу project_categories
//   - one-to-many с Image (одному проекту принадлежат несколько изображений)
//   - one-to-many с ProjectViewDaily (статистика просмотров по дням)
//
// Особенности:
//   - Slug генерируется автоматически из Title с транслитерацией
//   - SortOrder определяет порядок отображения (drag & drop в админке)
//   - Featured проекты показываются на главной странице
//   - ViewCount агрегируется из ProjectViewDaily для быстрого доступа
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Size        string    `json:"size"`        // Размер экрана, например "12x5" метров
	PixelPitch  string    `json:"pixel_pitch"` // Шаг пикселя, например "P6.66"
	Completed   bool      `json:"completed" gorm:"default:true"`
	Featured    bool      `json:"featured" gorm:"default:false"`
	ViewCount   int       `json:"view_count" gorm:"default:0"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Categories []Category `json:"categories,omitempty" gorm:"many2many:project_categories;"`
	Images     []Image    `json:"images,omitempty" gorm:"foreignKey:ProjectID"`
}

// Image представляет изображение проекта.
//
// Связи:
//   - many-to-one с Project (много изображений принадлежат одному проекту)
//
// Особенности:
//   - Filename генерируется автоматически: project_{id}_{timestamp}_{index}.{ext}
//   - CropX, CropY, CropScale используются для настройки preview в crop-editor
//   - SortOrder определяет порядок отображения изображений проекта
//   - ProjectID может быть NULL для общих изображений (не привязанных к проекту)
//   - IsPrimary отмечает главное изображение проекта (показывается в карточках и на главной)
//   - Thumbnails: автоматически генерируются в 2-х размерах (small, medium)
type Image struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ProjectID    *uint  `json:"project_id"` // Nullable: может быть null для общих изображений
	Filename     string `json:"filename" gorm:"not null"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path" gorm:"not null"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	SortOrder    int    `json:"sort_order" gorm:"default:0"`
	IsPrimary    bool   `json:"is_primary" gorm:"default:false;index"` // Главное изображение проекта

	// Thumbnails (автоматически генерируются при загрузке и при изменении crop)
	ThumbnailSmallPath  string `json:"thumbnail_small_path"`  // 400x300 для карточек (~50KB)
	ThumbnailMediumPath string `json:"thumbnail_medium_path"` // 1200x900 для галереи (~180KB)

	// Настройки кроппинга для превью (используются в crop-editor.js)
	CropX     float64 `json:"crop_x" gorm:"default:50"`      // Позиция X центра в процентах (0-100)
	CropY     float64 `json:"crop_y" gorm:"default:50"`      // Позиция Y центра в процентах (0-100)
	CropScale float64 `json:"crop_scale" gorm:"default:1.0"` // Масштаб изображения (0.5-3.0)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// Service представляет услугу, предоставляемую компанией.
//
// Примеры услуг:
//   - Продажа интерьерных LED экранов
//   - Продажа уличных LED экранов
//   - Обслуживание LED экранов
//   - Изготовление металлоконструкций
//
// Featured услуги отображаются на главной странице.
type Service struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	ShortDesc   string    `json:"short_desc"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"` // Имя иконки FontAwesome или путь к SVG файлу
	Featured    bool      `json:"featured" gorm:"default:false"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactForm представляет заявку клиента с сайта (CRM система).
//
// Жизненный цикл заявки:
//  1. Создание со статусом "new"
//  2. Обработка менеджером -> статус "processed"
//  3. Архивирование -> статус "archived" + установка ArchivedAt
//
// Дополнительные возможности:
//   - Заметки (ContactNote) для истории взаимодействия с клиентом
//   - Напоминания (RemindAt + RemindFlag) для follow-up звонков
//   - Поиск и фильтрация по множеству полей
//   - Экспорт в CSV для внешних CRM систем
type ContactForm struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Phone       string     `json:"phone" gorm:"not null"`
	Email       string     `json:"email"`
	Company     string     `json:"company"`
	ProjectType string     `json:"project_type"`
	Message     string     `json:"message"`
	Website     string     `json:"website" form:"website" gorm:"-"`                    // Honeypot поле (не сохраняется в БД)
	Source      string     `json:"source"`                                             // Источник заявки: "contact_form", "calculator", "phone_call" и т.д.
	Status      string     `json:"status" gorm:"type:varchar(20);default:'new';index"` // Статус: "new", "processed", "archived"
	CreatedAt   time.Time  `json:"created_at" gorm:"index"`
	ArchivedAt  *time.Time `json:"archived_at" gorm:"index"`         // NULL = активная заявка, NOT NULL = архив
	RemindAt    *time.Time `json:"remind_at" gorm:"index"`           // Дата/время напоминания для перезвона (МСК)
	RemindFlag  bool       `json:"remind_flag" gorm:"default:false"` // Флаг активного напоминания
}

// ContactNote представляет заметку менеджера по заявке клиента.
//
// Связи:
//   - many-to-one с ContactForm (много заметок принадлежат одной заявке)
//
// Используется для:
//   - Истории взаимодействия с клиентом
//   - Важных деталей переговоров
//   - Напоминаний для коллег
//
// Заметки отсортированы по created_at DESC (новые сверху).
type ContactNote struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ContactID uint      `json:"contact_id" gorm:"index;not null"`
	Text      string    `json:"text" gorm:"type:text;not null"`
	Author    string    `json:"author" gorm:"type:varchar(100)"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// ProjectViewDaily представляет агрегированные просмотры проекта по дням.
//
// Связи:
//   - many-to-one с Project (CASCADE DELETE при удалении проекта)
//
// Особенности:
//   - Day хранится как date (обнуленное время до полуночи МСК)
//   - Views инкрементируется через UPSERT (INSERT ... ON CONFLICT)
//   - Уникальный индекс (project_id, day) предотвращает дубликаты
//   - Используется для построения графиков аналитики в dashboard
//
// Трекинг происходит через /api/track/project-view/:id
type ProjectViewDaily struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"not null;uniqueIndex:uniq_project_day;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Day       time.Time `json:"day" gorm:"type:date;not null;uniqueIndex:uniq_project_day"`
	Views     int64     `json:"views" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Admin представляет администратора системы с доступом к админ-панели.
//
// Безопасность:
//   - PasswordHash хранит bcrypt хеш пароля (cost factor 10)
//   - JSON тег "-" исключает PasswordHash из API ответов
//   - IsActive позволяет деактивировать аккаунты без удаления
//   - LastLoginAt отслеживает активность администратора
//
// Создание админа:
//   - Через утилиту: go run cmd/create-admin/main.go
//   - Программно через handlers.HashPassword() + db.Create()
//
// Аутентификация:
//   - JWT токены (HS256) с истечением 7 дней
//   - Хранение в HTTP-only cookie "admin_token"
type Admin struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"unique;not null;size:50"`
	PasswordHash string     `json:"-" gorm:"not null"` // JSON:"-" исключает поле из сериализации (безопасность)
	Email        string     `json:"email" gorm:"size:100"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PriceItem представляет позицию в прайс-листе (например, "Билборд 6x3").
//
// Связи:
//   - one-to-many с PriceSpecification (одной позиции принадлежат несколько характеристик)
//
// Особенности:
//   - HasSpecifications определяет, нужна ли таблица характеристик для этой позиции
//   - PriceFrom - цена "от" в рублях
//   - ImagePath - путь к изображению продукта
//   - SortOrder определяет порядок отображения на странице
//   - IsActive позволяет скрывать позиции без удаления
type PriceItem struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Title             string `json:"title" gorm:"not null"`                   // Название (например, "Билборд 6x3")
	Description       string `json:"description"`                             // Краткое описание
	ImagePath         string `json:"image_path"`                              // Путь к оригинальному изображению
	PriceFrom         int    `json:"price_from" gorm:"not null"`              // Цена "от" в рублях
	HasSpecifications bool   `json:"has_specifications" gorm:"default:false"` // Есть ли таблица характеристик
	SortOrder         int    `json:"sort_order" gorm:"default:0"`             // Порядок отображения
	IsActive          bool   `json:"is_active" gorm:"default:true"`           // Активность

	// Поля для работы с миниатюрами и кроппингом
	ThumbnailSmallPath  string  `json:"thumbnail_small_path"`          // Путь к маленькой миниатюре (400x300)
	ThumbnailMediumPath string  `json:"thumbnail_medium_path"`         // Путь к средней миниатюре (1200x900)
	CropX               float64 `json:"crop_x" gorm:"default:50"`      // Позиция кроппинга по X (0-100%)
	CropY               float64 `json:"crop_y" gorm:"default:50"`      // Позиция кроппинга по Y (0-100%)
	CropScale           float64 `json:"crop_scale" gorm:"default:1.0"` // Масштаб кроппинга (1.0-3.0)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Specifications []PriceSpecification `json:"specifications,omitempty" gorm:"foreignKey:PriceItemID;constraint:OnDelete:CASCADE"`
	Images         []PriceImage         `json:"images,omitempty" gorm:"foreignKey:PriceItemID;constraint:OnDelete:CASCADE"`
}

// PriceSpecification представляет характеристику (строку в таблице) для позиции прайса.
//
// Связи:
//   - many-to-one с PriceItem (много характеристик принадлежат одной позиции)
//
// Особенности:
//   - SpecGroup используется для группировки характеристик (например, "Параметры экрана", "Модуль/кабинет")
//   - Группы отображаются как синие заголовки в таблице
//   - SortOrder определяет порядок отображения внутри группы
//   - SpecKey - название параметра, SpecValue - значение параметра
type PriceSpecification struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PriceItemID uint      `json:"price_item_id" gorm:"not null;index;constraint:OnDelete:CASCADE"`
	SpecGroup   string    `json:"spec_group" gorm:"not null"` // Группа (например, "Параметры экрана")
	SpecKey     string    `json:"spec_key" gorm:"not null"`   // Название параметра (например, "Разрешение по горизонтали")
	SpecValue   string    `json:"spec_value" gorm:"not null"` // Значение (например, "1152")
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	PriceItem *PriceItem `json:"price_item,omitempty" gorm:"foreignKey:PriceItemID"`
}

// PriceImage представляет изображение для позиции прайса.
//
// Связи:
//   - many-to-one с PriceItem (много изображений принадлежат одной позиции прайса)
//
// Особенности:
//   - IsPrimary: только одно изображение может быть главным для каждой позиции
//   - Thumbnails: автоматически генерируются в 2-х размерах (small, medium)
//   - SortOrder: определяет порядок отображения в галерее
type PriceImage struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	PriceItemID  uint   `json:"price_item_id" gorm:"not null;index;constraint:OnDelete:CASCADE"`
	Filename     string `json:"filename" gorm:"not null"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path" gorm:"not null"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	SortOrder    int    `json:"sort_order" gorm:"default:0"`
	IsPrimary    bool   `json:"is_primary" gorm:"default:false;index"` // Главное изображение позиции

	// Thumbnails (автоматически генерируются при загрузке и при изменении crop)
	ThumbnailSmallPath  string `json:"thumbnail_small_path"`  // 400x300 для карточек (~50KB)
	ThumbnailMediumPath string `json:"thumbnail_medium_path"` // 1200x900 для галереи (~180KB)

	// Настройки кроппинга для превью (используются в crop-editor.js)
	CropX     float64 `json:"crop_x" gorm:"default:50"`      // Позиция X центра в процентах (0-100)
	CropY     float64 `json:"crop_y" gorm:"default:50"`      // Позиция Y центра в процентах (0-100)
	CropScale float64 `json:"crop_scale" gorm:"default:1.0"` // Масштаб изображения (0.5-3.0)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PriceItem *PriceItem `json:"price_item,omitempty" gorm:"foreignKey:PriceItemID"`
}

// PriceViewDaily представляет агрегированные просмотры позиции прайса по дням.
//
// Связи:
//   - many-to-one с PriceItem (CASCADE DELETE при удалении позиции)
//
// Особенности:
//   - Day хранится как date (обнуленное время до полуночи МСК)
//   - Views инкрементируется через UPSERT (INSERT ... ON CONFLICT)
//   - Уникальный индекс (price_item_id, day) предотвращает дубликаты
//   - Используется для построения аналитики топ-5 позиций в dashboard
//
// Трекинг происходит через /api/track/price-view/:id
type PriceViewDaily struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PriceItemID uint      `json:"price_item_id" gorm:"not null;uniqueIndex:uniq_price_day;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Day         time.Time `json:"day" gorm:"type:date;not null;uniqueIndex:uniq_price_day"`
	Views       int64     `json:"views" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MapPoint представляет точку на карте — адрес выполненной работы.
//
// Точки отображаются на Яндекс.Карте на странице контактов.
// Управляются через админку: /admin/map-points
type MapPoint struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude" gorm:"not null"`
	Longitude   float64   `json:"longitude" gorm:"not null"`
	PanoramaURL string    `json:"panorama_url"` // Ссылка на панораму Яндекс.Карт
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PromoPopup представляет настройки всплывающего окна для акций и промо.
//
// Особенности:
//   - Только одна запись в таблице (singleton pattern)
//   - IsActive определяет показывать ли popup на сайте
//   - Pages - JSON массив страниц для показа: ["home", "prices", "projects", "services", "contact", "led-guide"]
//   - TTLHours - через сколько часов показать снова (0 = показывать всегда)
//   - ShowDelaySeconds - задержка показа после загрузки страницы
//
// Управляется через /admin/promo
type PromoPopup struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Title            string    `json:"title" gorm:"not null;default:''"`
	Content          string    `json:"content" gorm:"type:text;not null;default:''"`
	IsActive         bool      `json:"is_active" gorm:"default:false"`
	Pages            string    `json:"pages" gorm:"type:text;not null;default:'[\"home\"]'"` // JSON массив страниц
	TTLHours         int       `json:"ttl_hours" gorm:"default:24"`                          // 0 = показывать всегда
	ShowDelaySeconds int       `json:"show_delay_seconds" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
