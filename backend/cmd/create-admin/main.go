package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"ledsite/internal/config"
	"ledsite/internal/database"
	"ledsite/internal/models"
)

func main() {
	fmt.Println("=== Утилита создания администратора ===")

	// Загружаем .env файл из backend/.env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Выполняем миграции (чтобы таблица admins существовала)
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Ввод имени пользователя
	var username, email string
	fmt.Print("Введите имя пользователя (username): ")
	username, err = reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading username: %v", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		log.Fatal("Имя пользователя не может быть пустым")
	}

	// Проверяем, существует ли уже такой пользователь
	var existing models.Admin
	if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
		fmt.Printf("\nПользователь '%s' уже существует!\n", username)
		fmt.Print("Хотите обновить пароль? (y/n): ")
		answer, readErr := reader.ReadString('\n')
		if readErr != nil {
			log.Fatalf("Error reading answer: %v", readErr)
		}
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" && answer != "да" {
			fmt.Println("Отменено.")
			return
		}

		// Обновляем пароль существующего пользователя
		var pwd, pwdConfirm string
		var pwdErr error
		fmt.Print("Введите новый пароль: ")
		pwd, pwdErr = readPassword()
		if pwdErr != nil {
			log.Fatalf("Error reading password: %v", pwdErr)
		}

		if len(pwd) < 8 {
			log.Fatal("Пароль должен содержать минимум 8 символов")
		}

		fmt.Print("\nПовторите пароль: ")
		pwdConfirm, pwdErr = readPassword()
		if pwdErr != nil {
			log.Fatalf("Error reading password: %v", pwdErr)
		}

		if pwd != pwdConfirm {
			log.Fatal("\nПароли не совпадают")
		}

		// Хешируем пароль
		hashedPassword, pwdErr := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if pwdErr != nil {
			log.Fatalf("Error hashing password: %v", pwdErr)
		}

		existing.PasswordHash = string(hashedPassword)
		existing.IsActive = true

		if err := db.Save(&existing).Error; err != nil {
			log.Fatalf("Failed to update admin: %v", err)
		}

		fmt.Printf("\n✓ Пароль для пользователя '%s' успешно обновлен!\n", username)
		return
	}

	// Создаём нового пользователя
	fmt.Print("Введите email (необязательно): ")
	email, err = reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading email: %v", err)
	}
	email = strings.TrimSpace(email)

	// Ввод пароля (скрытый)
	var password, passwordConfirm string
	fmt.Print("Введите пароль (минимум 8 символов): ")
	password, err = readPassword()
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}

	if len(password) < 8 {
		log.Fatal("\nПароль должен содержать минимум 8 символов")
	}

	fmt.Print("\nПовторите пароль: ")
	passwordConfirm, err = readPassword()
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}

	if password != passwordConfirm {
		log.Fatal("\nПароли не совпадают")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}

	// Создаём администратора
	admin := models.Admin{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		IsActive:     true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("\n✓ Администратор '%s' успешно создан!\n", username)
	fmt.Println("\nТеперь вы можете войти в админ-панель по адресу: http://localhost:8080/admin/login")
}

// readPassword читает пароль без отображения на экране
func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}
