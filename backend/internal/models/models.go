package models

import (
	"time"
)

// Category - категории проектов (интерьерные, уличные, медиафасады и т.д.)
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_categories;"`
}

// Project - проекты из портфолио
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Size        string    `json:"size"`        // например "12x5"
	PixelPitch  string    `json:"pixel_pitch"` // например "P6.66"
	Completed   bool      `json:"completed" gorm:"default:true"`
	Featured    bool      `json:"featured" gorm:"default:false"`
	ViewCount   int       `json:"view_count" gorm:"default:0"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Categories []Category `json:"categories,omitempty" gorm:"many2many:project_categories;"`
	Images     []Image    `json:"images,omitempty" gorm:"foreignKey:ProjectID"`
}

// Image - изображения проектов
type Image struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ProjectID     *uint  `json:"project_id"` // может быть null для общих изображений
	Filename      string `json:"filename" gorm:"not null"`
	OriginalName  string `json:"original_name"`
	FilePath      string `json:"file_path" gorm:"not null"`
	ThumbnailPath string `json:"thumbnail_path"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	Alt           string `json:"alt"`
	Caption       string `json:"caption"`
	SortOrder     int    `json:"sort_order" gorm:"default:0"`

	// Настройки кроппинга для превью
	CropX     float64 `json:"crop_x" gorm:"default:50"`      // позиция X в процентах (0-100)
	CropY     float64 `json:"crop_y" gorm:"default:50"`      // позиция Y в процентах (0-100)
	CropScale float64 `json:"crop_scale" gorm:"default:1.0"` // масштаб (0.5-3.0)

	CreatedAt time.Time `json:"created_at"`

	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// Service - услуги компании
type Service struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	ShortDesc   string    `json:"short_desc"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"` // имя иконки или путь к файлу
	Featured    bool      `json:"featured" gorm:"default:false"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContactForm - заявки с сайта
type ContactForm struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Phone       string    `json:"phone" gorm:"not null"`
	Email       string    `json:"email"`
	Company     string    `json:"company"`
	ProjectType string    `json:"project_type"`
	Message     string    `json:"message"`
	Source      string    `json:"source"` // откуда пришла заявка (контактная форма, калькулятор и т.д.)
	Status      string    `json:"status" gorm:"type:varchar(20);default:'new'"`
	CreatedAt   time.Time `json:"created_at"`
}

// Settings - настройки сайта
type Settings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"unique;not null"`
	Value     string    `json:"value"`
	Type      string    `json:"type"` // text, number, boolean, json
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
