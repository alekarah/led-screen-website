package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/handlers"
	"ledsite/internal/routes"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Инициализируем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Выполняем миграции
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализируем handlers
	h := handlers.New(db)

	// Настраиваем Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Настраиваем маршруты
	routes.Setup(router, h)

	// Статические файлы из frontend
	router.Static("/static", "../frontend/static")
	router.LoadHTMLGlob("../frontend/templates/*.html")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
