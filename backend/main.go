package main

import (
	"html/template"
	"log"
	"os"
	"time"

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
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Выполняем миграции
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализируем handlers
	h := handlers.New(db, cfg.MaxUploadSize)

	// Настраиваем Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
		"fmtTime": func(t time.Time) string {
			// Мск, если нужна локаль с часовым поясом
			loc, _ := time.LoadLocation("Europe/Moscow")
			return t.In(loc).Format("02.01.2006 15:04")
		},
		"add": func(a, b int) int { return a + b },
	}
	router.SetFuncMap(funcMap)

	// Статические файлы из frontend
	router.Static("/static", "../frontend/static")
	router.LoadHTMLGlob("../frontend/templates/*.html")

	// Настраиваем маршруты
	routes.Setup(router, h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
