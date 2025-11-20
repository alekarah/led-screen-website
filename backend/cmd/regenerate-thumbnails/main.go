package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/handlers"
	"ledsite/internal/models"
)

func main() {
	fmt.Println("=== Утилита регенерации thumbnails ===")
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

	// Получаем все изображения из базы
	var images []models.Image
	if err := db.Find(&images).Error; err != nil {
		log.Fatal("Failed to fetch images:", err)
	}

	if len(images) == 0 {
		fmt.Println("Изображений в базе данных не найдено.")
		return
	}

	fmt.Printf("Найдено %d изображений для обработки\n\n", len(images))

	successCount := 0
	errorCount := 0
	skippedCount := 0

	for i, image := range images {
		fmt.Printf("[%d/%d] Обработка изображения ID=%d: %s\n", i+1, len(images), image.ID, image.Filename)

		// Путь из БД относительно backend/, но мы в backend/cmd/regenerate-thumbnails/
		// Поднимаемся на 2 уровня вверх
		adjustedPath := "../../" + image.FilePath

		// Проверяем существование оригинального файла
		if _, err := os.Stat(adjustedPath); os.IsNotExist(err) {
			fmt.Printf("  ⚠️  ПРОПУЩЕНО: оригинальный файл не найден: %s\n", adjustedPath)
			skippedCount++
			continue
		}

		// Генерируем thumbnails с текущими настройками кроппинга
		cropParams := handlers.CropParams{
			X:     image.CropX,
			Y:     image.CropY,
			Scale: image.CropScale,
		}

		thumbnails, err := handlers.GenerateThumbnails(adjustedPath, cropParams)
		if err != nil {
			fmt.Printf("  ❌ ОШИБКА: %v\n", err)
			errorCount++
			continue
		}

		// Обновляем пути к thumbnails в модели
		updated := false
		if path, ok := thumbnails[handlers.ThumbnailSmall.Suffix]; ok {
			image.ThumbnailSmallPath = path
			updated = true
		}
		if path, ok := thumbnails[handlers.ThumbnailMedium.Suffix]; ok {
			image.ThumbnailMediumPath = path
			updated = true
		}

		if updated {
			if err := db.Save(&image).Error; err != nil {
				fmt.Printf("  ❌ ОШИБКА сохранения в БД: %v\n", err)
				errorCount++
				continue
			}
			fmt.Printf("  ✅ Успешно: созданы thumbnails (small, medium, large)\n")
			successCount++
		} else {
			fmt.Printf("  ⚠️  ПРОПУЩЕНО: не удалось создать thumbnails\n")
			skippedCount++
		}
	}

	fmt.Println()
	fmt.Println("=== Итоги ===")
	fmt.Printf("Всего изображений: %d\n", len(images))
	fmt.Printf("✅ Успешно обработано: %d\n", successCount)
	fmt.Printf("❌ Ошибок: %d\n", errorCount)
	fmt.Printf("⚠️  Пропущено: %d\n", skippedCount)
	fmt.Println()

	if successCount > 0 {
		fmt.Println("Thumbnails успешно регенерированы!")
		fmt.Println("Теперь можно запустить приложение и проверить результат.")
	}
}
