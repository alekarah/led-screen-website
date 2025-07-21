package config

import "os"

type Config struct {
	DatabaseURL   string
	Port          string
	Environment   string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	UploadPath    string
	MaxUploadSize int64
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/led_display_db?sslmode=disable"),
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		SMTPHost:      getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		UploadPath:    getEnv("UPLOAD_PATH", "../frontend/static/uploads"),
		MaxUploadSize: 10485760, // 10MB
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
