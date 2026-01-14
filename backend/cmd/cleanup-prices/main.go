package main

import (
	"fmt"
	"log"
	"os"

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
	if err := db.Order("created_at DESC").Find(&priceItems).Error; err != nil {
		log.Fatal("Ошибка получения позиций прайса:", err)
	}

	fmt.Printf("\nВсего позиций прайса: %d\n\n", len(priceItems))

	// Группируем по названию
	priceGroups := make(map[string][]models.PriceItem)
	for _, item := range priceItems {
		priceGroups[item.Title] = append(priceGroups[item.Title], item)
	}

	// Выводим информацию о дубликатах
	fmt.Println("Найденные позиции:")
	for title, items := range priceGroups {
		fmt.Printf("\n'%s': %d позиций\n", title, len(items))
		for _, item := range items {
			fmt.Printf("  ID=%d, Created=%s, Active=%v\n",
				item.ID,
				item.CreatedAt.Format("2006-01-02 15:04:05"),
				item.IsActive)
		}
	}

	// Спрашиваем подтверждение
	fmt.Println("\n-----------------------------------------")
	fmt.Println("Будут удалены все дубликаты, кроме первой созданной позиции.")
	fmt.Print("Продолжить? (yes/no): ")

	var answer string
	fmt.Scanln(&answer)

	if answer != "yes" {
		fmt.Println("Операция отменена")
		os.Exit(0)
	}

	// Удаляем дубликаты
	totalDeleted := 0
	for title, items := range priceGroups {
		if len(items) <= 1 {
			continue // нет дубликатов
		}

		// Оставляем самую раннюю запись
		toKeep := items[len(items)-1] // самая старая (из-за ORDER BY created_at DESC)

		fmt.Printf("\nОбработка '%s':\n", title)
		fmt.Printf("  Оставляем: ID=%d (Created=%s)\n", toKeep.ID, toKeep.CreatedAt.Format("2006-01-02 15:04:05"))

		for i := 0; i < len(items)-1; i++ {
			item := items[i]

			// Удаляем связанные характеристики
			if err := db.Where("price_item_id = ?", item.ID).Delete(&models.PriceSpecification{}).Error; err != nil {
				log.Printf("  ⚠ Ошибка удаления характеристик для ID=%d: %v", item.ID, err)
				continue
			}

			// Удаляем саму позицию
			if err := db.Delete(&item).Error; err != nil {
				log.Printf("  ⚠ Ошибка удаления позиции ID=%d: %v", item.ID, err)
				continue
			}

			fmt.Printf("  ✓ Удалено: ID=%d (Created=%s)\n", item.ID, item.CreatedAt.Format("2006-01-02 15:04:05"))
			totalDeleted++
		}
	}

	fmt.Printf("\n-----------------------------------------\n")
	fmt.Printf("Операция завершена. Удалено дубликатов: %d\n", totalDeleted)

	// Выводим итоговую статистику
	var finalCount int64
	db.Model(&models.PriceItem{}).Count(&finalCount)
	fmt.Printf("Осталось уникальных позиций: %d\n", finalCount)
}
