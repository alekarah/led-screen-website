package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ledsite/internal/models"
)

// Connect подключается к базе данных PostgreSQL
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Migrate выполняет миграции базы данных
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Category{},
		&models.Project{},
		&models.Image{},
		&models.Service{},
		&models.ContactForm{},
		&models.Settings{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Создаем начальные данные
	return seedInitialData(db)
}

// seedInitialData создает начальные данные в базе
func seedInitialData(db *gorm.DB) error {
	// Создаем базовые категории
	categories := []models.Category{
		{Name: "Интерьерные экраны", Slug: "interior", Description: "LED экраны для помещений"},
		{Name: "Уличные экраны", Slug: "outdoor", Description: "Рекламные щиты и уличные дисплеи"},
		{Name: "Медиафасады", Slug: "mediafacade", Description: "Фасадные LED экраны"},
		{Name: "АЗС", Slug: "gas-station", Description: "Экраны для автозаправочных станций"},
		{Name: "Торговые центры", Slug: "shopping", Description: "Экраны для торговых центров"},
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
