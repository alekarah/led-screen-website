package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/handlers"
	"ledsite/internal/models"
	"ledsite/internal/routes"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Инициализируем конфигурацию
	cfg := config.Load()

	// КРИТИЧНАЯ ПРОВЕРКА: в production обязательно нужен уникальный JWT_SECRET
	if cfg.Environment == "production" {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" || jwtSecret == "your-secret-key-change-in-production" {
			log.Fatal("CRITICAL SECURITY ERROR: Change JWT_SECRET in production!\n" +
				"Generate a strong secret with: openssl rand -base64 32")
		}
		log.Println("✓ JWT_SECRET validation passed")
	}

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
	h := handlers.New(db, cfg.MaxUploadSize, cfg.UploadPath)

	// Настраиваем Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
		"formatPrice": func(price int) string {
			// Форматируем число с разделителями тысяч (пробелами)
			if price < 1000 {
				return fmt.Sprintf("%d", price)
			}
			// Преобразуем в строку и добавляем пробелы
			s := fmt.Sprintf("%d", price)
			n := len(s)
			if n <= 3 {
				return s
			}
			// Вставляем пробелы каждые 3 цифры справа
			var result []byte
			for i, c := range s {
				if i > 0 && (n-i)%3 == 0 {
					result = append(result, ' ')
				}
				result = append(result, byte(c))
			}
			return string(result)
		},
		"fmtTime": func(t time.Time) string {
			// Мск, если нужна локаль с часовым поясом
			loc, err := time.LoadLocation("Europe/Moscow")
			if err != nil {
				loc = time.UTC // fallback to UTC
			}
			return t.In(loc).Format("02.01.2006 15:04")
		},
		"translateProjectType": func(projectType string) string {
			translations := map[string]string{
				"indoor":       "Внутренние LED экраны",
				"outdoor":      "Наружные LED экраны",
				"rental":       "Аренда оборудования",
				"service":      "Техническое обслуживание",
				"repair":       "Ремонт LED экранов",
				"consultation": "Консультация",
			}
			if translated, ok := translations[projectType]; ok {
				return translated
			}
			// Если тип не найден, возвращаем как есть
			return projectType
		},
		"add": func(a, b int) int { return a + b },
		"primaryImage": func(images []models.Image) models.Image {
			// Ищем главное изображение (is_primary = true)
			for _, img := range images {
				if img.IsPrimary {
					return img
				}
			}
			// Если главного нет, возвращаем первое (или пустое если массив пустой)
			if len(images) > 0 {
				return images[0]
			}
			return models.Image{}
		},
		// getThumbnail возвращает путь к thumbnail нужного размера
		// size: "small" (карточки), "medium" (галерея)
		// fallback к оригиналу если thumbnail не существует
		"getThumbnail": func(img models.Image, size string) string {
			var thumbPath string
			switch size {
			case "small":
				thumbPath = img.ThumbnailSmallPath
			case "medium":
				thumbPath = img.ThumbnailMediumPath
			default:
				thumbPath = ""
			}
			// Fallback к filename если thumbnail не существует
			if thumbPath == "" {
				return img.Filename
			}
			// Извлекаем только имя файла из полного пути
			parts := []rune(thumbPath)
			for i := len(parts) - 1; i >= 0; i-- {
				if parts[i] == '/' || parts[i] == '\\' {
					return string(parts[i+1:])
				}
			}
			return thumbPath
		},
		"toJSON": func(v interface{}) template.JS {
			// Конвертируем любое значение в JSON для использования в JavaScript
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		// imgVersion возвращает unix timestamp обновления изображения для cache-busting
		"imgVersion": func(img models.Image) int64 {
			if !img.UpdatedAt.IsZero() {
				return img.UpdatedAt.Unix()
			}
			if !img.CreatedAt.IsZero() {
				return img.CreatedAt.Unix()
			}
			return 0
		},
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

	// Bind только на localhost для безопасности (Nginx будет проксировать)
	bindAddr := "127.0.0.1:" + port
	log.Printf("Server starting on %s", bindAddr)
	log.Fatal(router.Run(bindAddr))
}
