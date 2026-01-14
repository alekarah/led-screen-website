package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ledsite/internal/config"
	"ledsite/internal/models"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Инициализируем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Подключено к базе данных")

	// Получаем все позиции прайса
	var priceItems []models.PriceItem
	if err := db.Find(&priceItems).Error; err != nil {
		log.Fatal("Ошибка получения позиций прайса:", err)
	}

	fmt.Printf("Найдено позиций прайса: %d\n\n", len(priceItems))

	updatedCount := 0

	for _, item := range priceItems {
		needsUpdate := false
		originalImagePath := item.ImagePath
		originalSmallPath := item.ThumbnailSmallPath
		originalMediumPath := item.ThumbnailMediumPath

		// Исправляем ImagePath
		if item.ImagePath != "" {
			fixed := fixPath(item.ImagePath)
			if fixed != item.ImagePath {
				item.ImagePath = fixed
				needsUpdate = true
			}
		}

		// Исправляем ThumbnailSmallPath
		if item.ThumbnailSmallPath != "" {
			fixed := fixPath(item.ThumbnailSmallPath)
			if fixed != item.ThumbnailSmallPath {
				item.ThumbnailSmallPath = fixed
				needsUpdate = true
			}
		}

		// Исправляем ThumbnailMediumPath
		if item.ThumbnailMediumPath != "" {
			fixed := fixPath(item.ThumbnailMediumPath)
			if fixed != item.ThumbnailMediumPath {
				item.ThumbnailMediumPath = fixed
				needsUpdate = true
			}
		}

		if needsUpdate {
			fmt.Printf("Обновление позиции ID=%d '%s':\n", item.ID, item.Title)
			if originalImagePath != item.ImagePath {
				fmt.Printf("  ImagePath: %s -> %s\n", originalImagePath, item.ImagePath)
			}
			if originalSmallPath != item.ThumbnailSmallPath {
				fmt.Printf("  ThumbnailSmallPath: %s -> %s\n", originalSmallPath, item.ThumbnailSmallPath)
			}
			if originalMediumPath != item.ThumbnailMediumPath {
				fmt.Printf("  ThumbnailMediumPath: %s -> %s\n", originalMediumPath, item.ThumbnailMediumPath)
			}

			if err := db.Save(&item).Error; err != nil {
				log.Printf("  ⚠ Ошибка обновления: %v\n", err)
			} else {
				fmt.Println("  ✓ Обновлено")
				updatedCount++
			}
		}
	}

	fmt.Printf("\n-----------------------------------------\n")
	fmt.Printf("Обновлено позиций: %d\n", updatedCount)
}

// fixPath - конвертирует любой путь в правильный веб-путь
func fixPath(path string) string {
	if path == "" {
		return path
	}

	// Нормализуем разделители
	path = strings.ReplaceAll(path, "\\", "/")

	// Если путь уже начинается с /static/uploads/ и не содержит frontend, всё ОК
	if strings.HasPrefix(path, "/static/uploads/") && !strings.Contains(path, "frontend") {
		return path
	}

	// Ищем последнее вхождение /static/uploads/ или static/uploads/
	if idx := strings.LastIndex(path, "/static/uploads/"); idx != -1 {
		// Берем всё начиная с /static/uploads/
		return path[idx:]
	}

	if idx := strings.LastIndex(path, "static/uploads/"); idx != -1 {
		// Добавляем ведущий слеш
		return "/" + path[idx:]
	}

	// Если путь содержит frontend, убираем её
	if strings.HasPrefix(path, "frontend/") {
		return "/" + strings.TrimPrefix(path, "frontend/")
	}

	// Если путь не содержит static/uploads, возвращаем как есть
	return path
}
