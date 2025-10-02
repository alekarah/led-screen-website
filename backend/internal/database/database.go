package database

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ledsite/internal/config"
	"ledsite/internal/models"
)

func gormLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

// Connect подключается к БД с учётом логгера и пула соединений
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel(cfg.DBLogLevel)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Пул соединений
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err == nil {
		if cfg.DBMaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		}
		if cfg.DBMaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
		}
		if cfg.DBConnMaxLifetimeMin > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute)
		}
	}

	return db, nil
}

// Migrate выполняет миграции
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Project{},
		&models.Image{},
		&models.Service{},
		&models.ContactForm{},
		&models.Settings{},
		&models.ContactNote{},
		&models.ProjectViewDaily{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS uniq_project_day
        ON project_view_dailies (project_id, day)
    `).Error; err != nil {
		return fmt.Errorf("ensure uniq_project_day index: %w", err)
	}

	return seedInitialData(db)
}

// seedInitialData создает начальные данные в базе
func seedInitialData(db *gorm.DB) error {
	// Очищаем старые категории при первом запуске
	var categoryCount int64
	db.Model(&models.Category{}).Count(&categoryCount)

	// Если категорий больше 6, значит есть старые - очищаем все
	if categoryCount > 6 {
		// Удаляем связи проектов с категориями
		db.Exec("DELETE FROM project_categories")

		// Удаляем все категории
		db.Exec("DELETE FROM categories")
	}

	// Создаем базовые категории
	categories := []models.Category{
		{Name: "Рекламные щиты", Slug: "billboards", Description: "Рекламные щиты различных размеров"},
		{Name: "АЗС (автозаправки)", Slug: "gas-stations", Description: "Тотемы и фасады для автозаправочных станций"},
		{Name: "Торговые центры", Slug: "shopping-centers", Description: "Экраны для торговых центров"},
		{Name: "Фундаментные работы", Slug: "foundation-work", Description: "Изготовление и монтаж фундаментных блоков"},
		{Name: "Обслуживание", Slug: "maintenance", Description: "Техническое обслуживание LED экранов"},
		{Name: "Ремонт модулей", Slug: "module-repair", Description: "Ремонт и замена модулей"},
	}

	for _, category := range categories {
		var existingCategory models.Category
		if db.Where("slug = ?", category.Slug).First(&existingCategory).Error == gorm.ErrRecordNotFound {
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category %s: %w", category.Name, err)
			}
		}
	}

	// Создаем базовые услуги
	services := []models.Service{
		{
			Name:        "Продажа интерьерных LED экранов",
			Slug:        "interior-sales",
			ShortDesc:   "Поставка LED экранов для помещений",
			Description: "Полный цикл поставки LED экранов для интерьера: от консультации до пуско-наладочных работ",
			Icon:        "monitor",
			Featured:    true,
			SortOrder:   1,
		},
		{
			Name:        "Продажа уличных LED экранов",
			Slug:        "outdoor-sales",
			ShortDesc:   "Рекламные щиты и уличные дисплеи",
			Description: "Изготовление и монтаж уличных LED экранов любой сложности с полным комплексом работ",
			Icon:        "billboard",
			Featured:    true,
			SortOrder:   2,
		},
		{
			Name:        "Обслуживание LED экранов",
			Slug:        "maintenance",
			ShortDesc:   "Техническое обслуживание и ремонт",
			Description: "Профессиональное обслуживание LED дисплеев, замена модулей, диагностика",
			Icon:        "tools",
			Featured:    true,
			SortOrder:   3,
		},
		{
			Name:        "Изготовление металлоконструкций",
			Slug:        "metalwork",
			ShortDesc:   "Каркасы и основания для экранов",
			Description: "Проектирование и изготовление металлоконструкций, фундаментных блоков",
			Icon:        "construction",
			Featured:    false,
			SortOrder:   4,
		},
	}

	for _, service := range services {
		var existingService models.Service
		if db.Where("slug = ?", service.Slug).First(&existingService).Error == gorm.ErrRecordNotFound {
			if err := db.Create(&service).Error; err != nil {
				return fmt.Errorf("failed to create service %s: %w", service.Name, err)
			}
		}
	}

	// Создаем базовые настройки сайта
	settings := []models.Settings{
		{Key: "company_name", Value: "Service 'n' Repair LED Display", Type: "text"},
		{Key: "company_phone", Value: "+7 (921) 429-17-02", Type: "text"},
		{Key: "company_email", Value: "info@ledsite.ru", Type: "text"},
		{Key: "company_address", Value: "Санкт-Петербург и Ленинградская область", Type: "text"},
		{Key: "meta_title", Value: "LED экраны в СПб | Продажа, монтаж, обслуживание", Type: "text"},
		{Key: "meta_description", Value: "Продажа и обслуживание LED экранов в Санкт-Петербурге. Интерьерные и уличные дисплеи, ремонт, металлоконструкции.", Type: "text"},
	}

	for _, setting := range settings {
		var existingSetting models.Settings
		if db.Where("key = ?", setting.Key).First(&existingSetting).Error == gorm.ErrRecordNotFound {
			if err := db.Create(&setting).Error; err != nil {
				return fmt.Errorf("failed to create setting %s: %w", setting.Key, err)
			}
		}
	}

	return nil
}
