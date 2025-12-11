package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ledsite/internal/models"
	"log"
	"net/http"
	"os"
	"time"
)

// TelegramNotification представляет структуру данных для отправки в Telegram бот
type TelegramNotification struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email,omitempty"`
	Company     string `json:"company,omitempty"`
	ProjectType string `json:"project_type,omitempty"`
	Message     string `json:"message,omitempty"`
	ContactID   uint   `json:"contact_id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// SendTelegramNotification отправляет уведомление о новой заявке в Telegram
func SendTelegramNotification(contact *models.ContactForm) {
	// Проверяем что URL Telegram бота настроен
	telegramBotURL := os.Getenv("TELEGRAM_BOT_URL")
	if telegramBotURL == "" {
		log.Println("TELEGRAM_BOT_URL не настроен, пропускаем отправку уведомления")
		return
	}

	// Формируем данные для отправки
	notification := TelegramNotification{
		Name:        contact.Name,
		Phone:       contact.Phone,
		Email:       contact.Email,
		Company:     contact.Company,
		ProjectType: contact.ProjectType,
		Message:     contact.Message,
		ContactID:   contact.ID,
		Timestamp:   time.Now().Format("02.01.2006 15:04"),
	}

	// Преобразуем в JSON
	jsonData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Ошибка сериализации данных для Telegram: %v", err)
		return
	}

	// Отправляем POST запрос
	resp, err := http.Post(
		telegramBotURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Ошибка отправки уведомления в Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram бот вернул ошибку: статус %d", resp.StatusCode)
		return
	}

	log.Printf("Уведомление успешно отправлено в Telegram для контакта: %s", contact.Name)
}

// SendTelegramNotificationAsync отправляет уведомление асинхронно в горутине
func SendTelegramNotificationAsync(contact *models.ContactForm) {
	go func() {
		// Восстановление после паники
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Паника при отправке Telegram уведомления: %v", r)
			}
		}()

		SendTelegramNotification(contact)
	}()
}

// ValidateTelegramBotConnection проверяет доступность Telegram бота
func ValidateTelegramBotConnection() error {
	telegramBotURL := os.Getenv("TELEGRAM_BOT_URL")
	if telegramBotURL == "" {
		return fmt.Errorf("TELEGRAM_BOT_URL не настроен")
	}

	// Проверяем healthcheck endpoint
	healthURL := telegramBotURL[:len(telegramBotURL)-len("/api/send-notification")] + "/health"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Telegram боту: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram бот вернул статус: %d", resp.StatusCode)
	}

	log.Println("✓ Telegram бот доступен и работает")
	return nil
}
