package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/models"
)

func main() {
	fmt.Println("=== Утилита очистки путей thumbnails ===")
	fmt.Println()

	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Инициализируем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Очищаем пути к thumbnails для ВСЕХ изображений
	// Используем Session для разрешения глобального обновления
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&models.Image{}).
		Updates(map[string]interface{}{
			"thumbnail_small_path":  "",
			"thumbnail_medium_path": "",
			"thumbnail_large_path":  "",
		})

	if result.Error != nil {
		log.Fatal("Failed to clear thumbnail paths:", result.Error)
	}

	fmt.Printf("✅ Очищено путей thumbnails для %d изображений\n", result.RowsAffected)
	fmt.Println()
	fmt.Println("Теперь старые изображения будут использовать оригиналы с CSS transform.")
	fmt.Println("Новые изображения автоматически получат thumbnails при загрузке.")
}
