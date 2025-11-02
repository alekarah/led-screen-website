// Package config управляет конфигурацией приложения из переменных окружения.
//
// Все настройки загружаются из .env файла через godotenv или из системных переменных.
// Если переменная не установлена - используется значение по умолчанию.
//
// Основные группы настроек:
//   - Приложение (DatabaseURL, Port, Environment)
//   - SMTP для отправки email (в разработке)
//   - Загрузка файлов (UploadPath, MaxUploadSize)
//   - Connection pooling для PostgreSQL
//
// Пример использования:
//
//	cfg := config.Load()
//	db := database.Connect(cfg)
//	router.Run(":" + cfg.Port)
package config

import (
	"os"
	"strconv"
)

// Config содержит всю конфигурацию приложения.
//
// Значения загружаются из переменных окружения при вызове Load().
// Все поля имеют разумные значения по умолчанию для разработки.
type Config struct {
	// Основные настройки приложения
	DatabaseURL string // PostgreSQL DSN (Data Source Name)
	Port        string // Порт для HTTP сервера (например "8080")
	Environment string // "development" или "production"

	// SMTP для email уведомлений (в разработке, не используется)
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string

	// Файловые загрузки
	UploadPath    string // Путь для сохранения изображений
	MaxUploadSize int64  // Максимальный размер файла в байтах (10MB)

	// Настройки connection pool для PostgreSQL
	DBLogLevel           string // Уровень логирования GORM: "silent", "error", "warn", "info"
	DBMaxOpenConns       int    // Максимальное количество открытых соединений (рекомендуется 20-25)
	DBMaxIdleConns       int    // Максимальное количество idle соединений (рекомендуется 10)
	DBConnMaxLifetimeMin int    // Максимальное время жизни соединения в минутах (рекомендуется 30)
}

// Load загружает конфигурацию из переменных окружения с fallback на значения по умолчанию.
//
// Переменные окружения должны быть установлены до вызова этой функции.
// Обычно загружаются через godotenv.Load() из .env файла.
//
// Рекомендуется вызывать в начале main():
//
//	godotenv.Load()
//	cfg := config.Load()
//
// Для production обязательно установить:
//   - DATABASE_URL (с сильным паролем)
//   - JWT_SECRET (сгенерировать через openssl rand -base64 32)
//   - ENVIRONMENT=production
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/led_display_db?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		UploadPath:    getEnv("UPLOAD_PATH", "../frontend/static/uploads"),
		MaxUploadSize: 10485760, // 10MB

		DBLogLevel:           getEnv("DB_LOG_LEVEL", "warn"),
		DBMaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetimeMin: getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 30),
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию.
//
// Параметры:
//   - key: имя переменной окружения
//   - def: значение по умолчанию, если переменная не установлена
//
// Возвращает значение переменной или def, если переменная пуста.
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getEnvInt возвращает целочисленное значение переменной окружения или значение по умолчанию.
//
// Параметры:
//   - key: имя переменной окружения
//   - def: значение по умолчанию, если переменная не установлена или не является числом
//
// Если преобразование в int не удается, возвращается def.
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
