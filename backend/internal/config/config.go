package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	Port        string
	Environment string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string

	UploadPath    string
	MaxUploadSize int64

	// Настройки БД
	DBLogLevel           string // silent|error|warn|info
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin int // минуты
}

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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
