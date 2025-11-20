package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/models"
)

func main() {
	fmt.Println("=== Проверка базы данных ===")
	fmt.Println()

	// Загружаем переменные окружения (из backend/.env)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Инициализируем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Получаем ВСЕ изображения
	var images []models.Image
	if err := db.Find(&images).Error; err != nil {
		log.Fatal("Failed to fetch images:", err)
	}

	fmt.Printf("Найдено %d изображений:\n\n", len(images))

	for _, img := range images {
		fmt.Printf("ID: %d\n", img.ID)
		fmt.Printf("  Filename: %s\n", img.Filename)
		fmt.Printf("  Crop: X=%v, Y=%v, Scale=%v\n", img.CropX, img.CropY, img.CropScale)
		fmt.Printf("  Small:  %s\n", img.ThumbnailSmallPath)
		fmt.Printf("  Medium: %s\n", img.ThumbnailMediumPath)
		fmt.Println()
	}
}
