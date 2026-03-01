// Package database управляет подключением к PostgreSQL и миграциями.
//
// Основные функции:
//   - Connect() - установка соединения с БД и настройка connection pool
//   - Migrate() - автоматическое создание/обновление таблиц через GORM AutoMigrate
//   - seedInitialData() - создание начальных данных (категории, услуги, настройки)
//
// Особенности:
//   - Connection pooling настраивается через Config (MaxOpenConns, MaxIdleConns)
//   - Уровень логирования SQL запросов управляется через DBLogLevel
//   - Индексы создаются автоматически при миграции
//   - Seed данные создаются только если их еще нет (идемпотентность)
//
// Пример использования:
//
//	cfg := config.Load()
//	db, err := database.Connect(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	database.Migrate(db)
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

// gormLogLevel конвертирует строковое представление уровня логирования в GORM LogLevel.
//
// Поддерживаемые значения:
//   - "silent" - не логировать SQL запросы
//   - "error" - только ошибки
//   - "warn" - ошибки и предупреждения (по умолчанию)
//   - "info" - все запросы включая SELECT
//
// Используется для настройки GORM logger при подключении к БД.
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

// Connect устанавливает подключение к PostgreSQL с настройками connection pool.
//
// Параметры:
//   - cfg: конфигурация с DatabaseURL и параметрами connection pool
//
// Настраивает:
//   - GORM logger (уровень логирования SQL запросов)
//   - MaxOpenConns - максимальное количество открытых соединений
//   - MaxIdleConns - максимальное количество idle соединений в pool
//   - ConnMaxLifetime - максимальное время жизни соединения
//
// Connection pool важен для production:
//   - Предотвращает исчерпание соединений PostgreSQL
//   - Переиспользует соединения вместо создания новых
//   - Автоматически закрывает устаревшие соединения
//
// Возвращает ошибку если не удалось подключиться к БД.
//
// Пример:
//
//	db, err := database.Connect(cfg)
//	if err != nil {
//	    log.Fatal("Failed to connect:", err)
//	}
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

// Migrate выполняет миграции базы данных и создает начальные данные.
//
// Выполняемые операции:
//  1. GORM AutoMigrate для всех моделей (создает/обновляет таблицы)
//  2. Создание дополнительных индексов для project_view_dailies
//  3. Вызов seedInitialData() для создания начальных данных
//
// AutoMigrate безопасен:
//   - Создает таблицы если их нет
//   - Добавляет новые колонки
//   - НЕ удаляет существующие данные
//   - НЕ изменяет типы существующих колонок
//
// Для сложных миграций (изменение типов, удаление колонок) нужны отдельные SQL скрипты.
//
// Индексы для аналитики просмотров:
//   - uniq_project_day: уникальный индекс (project_id, day) для UPSERT
//   - idx_pvd_project: быстрый поиск по проекту
//   - idx_pvd_day: быстрый поиск по дате
//
// Возвращает ошибку если миграция не удалась.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Category{},
		&models.Project{},
		&models.Image{},
		&models.Service{},
		&models.ContactForm{},
		&models.ContactNote{},
		&models.ProjectViewDaily{},
		&models.Admin{},
		&models.PriceItem{},
		&models.PriceSpecification{},
		&models.PriceImage{},
		&models.PriceViewDaily{},
		&models.PromoPopup{},
		&models.MapPoint{},
		&models.SiteSettings{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS uniq_project_day
        ON project_view_dailies (project_id, day)
    `).Error; err != nil {
		return fmt.Errorf("ensure uniq_project_day index: %w", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_pvd_project     ON project_view_dailies (project_id);
		CREATE INDEX IF NOT EXISTS idx_pvd_day         ON project_view_dailies (day);
		CREATE INDEX IF NOT EXISTS idx_pvd_project_day ON project_view_dailies (project_id, day);
	`).Error; err != nil {
		return fmt.Errorf("ensure pvd indexes: %w", err)
	}

	return seedInitialData(db)
}

// seedInitialData создает начальные данные в базе (идемпотентная операция).
//
// Создает:
//   - 6 базовых категорий проектов (Рекламные щиты, АЗС, ТЦ, и т.д.)
//   - 4 базовые услуги компании (Продажа, обслуживание, металлоконструкции)
//   - Базовые настройки сайта (название, телефон, email, SEO)
//
// Особенности:
//   - Проверяет существование по slug перед созданием (идемпотентность)
//   - Можно запускать многократно - не создаст дубликаты
//   - При первом запуске (>6 категорий) очищает старые тестовые данные
//
// Очистка старых категорий нужна для обратной совместимости
// с предыдущей версией, где было больше категорий.
//
// Возвращает ошибку если не удалось создать данные.
func seedInitialData(db *gorm.DB) error {
	// Проверяем количество категорий для обратной совместимости
	var categoryCount int64
	db.Model(&models.Category{}).Count(&categoryCount)

	// Если категорий больше 5, значит это старая версия с тестовыми данными - очищаем
	if categoryCount > 5 {
		// Удаляем связи проектов с категориями
		db.Exec("DELETE FROM project_categories")

		// Удаляем все категории
		db.Exec("DELETE FROM categories")
	}

	// Создаем базовые категории
	categories := []models.Category{
		{Name: "Изготовление металлоконструкций", Slug: "metalwork", Description: "Проектирование и изготовление металлоконструкций для LED экранов"},
		{Name: "Уличные решения digital реклама", Slug: "outdoor-solutions", Description: "LED экраны для наружной рекламы и уличной установки"},
		{Name: "Интерьерные решения", Slug: "interior-solutions", Description: "LED экраны для помещений и интерьеров"},
		{Name: "Медиафасады", Slug: "media-facades", Description: "Медиафасады и крупноформатные LED экраны для зданий"},
		{Name: "Сервис LED экранов", Slug: "led-service", Description: "Техническое обслуживание, ремонт и настройка LED экранов"},
	}

	// Создаем категории только если их еще нет (проверка по slug)
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

	// Создаем услуги только если их еще нет (проверка по slug)
	for _, service := range services {
		var existingService models.Service
		if db.Where("slug = ?", service.Slug).First(&existingService).Error == gorm.ErrRecordNotFound {
			if err := db.Create(&service).Error; err != nil {
				return fmt.Errorf("failed to create service %s: %w", service.Name, err)
			}
		}
	}

	// Создаем запись PromoPopup если её ещё нет (singleton)
	var promoCount int64
	db.Model(&models.PromoPopup{}).Count(&promoCount)
	if promoCount == 0 {
		promo := models.PromoPopup{
			Title:            "",
			Content:          "",
			IsActive:         false,
			Pages:            `["home"]`,
			TTLHours:         24,
			ShowDelaySeconds: 0,
		}
		if err := db.Create(&promo).Error; err != nil {
			return fmt.Errorf("failed to create promo popup: %w", err)
		}
	}

	// Заполняем пустые поля SiteSettings дефолтными значениями (для существующих записей)
	var settings models.SiteSettings
	if db.First(&settings).Error == nil {
		updates := map[string]interface{}{}
		if settings.PhoneNote == "" {
			updates["phone_note"] = "Звонки принимаем с 9:00 до 21:00"
		}
		if settings.EmailNote == "" {
			updates["email_note"] = "Ответим в течение 2 часов"
		}
		if settings.Address == "" {
			updates["address"] = "Санкт-Петербург\nЛенинградская область"
		}
		if settings.AddressNote == "" {
			updates["address_note"] = "Выезд на объект бесплатно"
		}
		if settings.WorkHours == "" {
			updates["work_hours"] = "Пн-Пт: 9:00 - 18:00\nСб-Вс: 10:00 - 16:00"
		}
		if settings.WorkHoursNote == "" {
			updates["work_hours_note"] = "Аварийные вызовы 24/7"
		}
		if settings.StatsProjects == 0 {
			updates["stats_projects"] = 200
		}
		if settings.StatsYears == 0 {
			updates["stats_years"] = 5
		}
		if len(updates) > 0 {
			db.Model(&settings).Updates(updates)
		}
	}

	return nil
}
