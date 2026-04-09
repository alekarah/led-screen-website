# 🖥️ LED Screen Website

> Корпоративный веб-сайт для компании по продаже, ремонту и обслуживанию LED дисплеев

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)
[![Production](https://img.shields.io/badge/Production-Live-success)](https://s-n-r.ru)
[![CI](https://github.com/alekarah/led-screen-website/actions/workflows/ci.yml/badge.svg)](https://github.com/alekarah/led-screen-website/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/alekarah/led-screen-website/branch/main/graph/badge.svg)](https://codecov.io/gh/alekarah/led-screen-website)

**🌐 Production сайт:** [https://s-n-r.ru](https://s-n-r.ru)

## 📋 О проекте

Современный веб-сайт для компании, специализирующейся на LED дисплеях. Проект включает в себя публичную часть для клиентов и полнофункциональную административную панель для управления контентом.

> **⚠️ Примечание:** Этот проект выполнен на коммерческой основе для клиента. Публикация согласована с заказчиком и предназначена исключительно для демонстрации технических навыков в портфолио.

**Особенности:**
- 🎨 Современный адаптивный дизайн
- 🔐 Безопасная система авторизации (JWT)
- 📊 CRM-система для управления заявками
- 📱 Telegram бот с интерактивными кнопками (быстрая обработка заявок)
- ⏰ Автоматические напоминания через Telegram (проверка каждые 5 минут)
- 🖼️ Система оптимизации изображений с WebP thumbnails (90% качество, экономия 25-35% размера)
- ✂️ Продвинутый редактор обрезки изображений с live preview
- 📈 Яндекс.Метрика (вебвизор, карта кликов, отслеживание источников трафика)
- 💾 Экспорт данных в CSV
- 🔍 SEO-оптимизация (sitemap.xml, robots.txt, Open Graph, Schema.org)
- 🧪 Unit тесты и CI/CD (GitHub Actions)
- 🛡️ Production-grade безопасность (SSH ключи, fail2ban, автообновления)
- 📦 Автоматические ежедневные бэкапы БД и файлов
- 🚀 Automated deployment скрипты

## ✨ Возможности

### Публичная часть
- **Портфолио проектов** с фильтрацией по категориям
- **Страница прайс-листа** с модальными окнами и характеристиками
- **Калькулятор стоимости LED экрана** — интерактивный расчёт на странице `/prices`: выбор типа (интерьерный/уличный), шага пикселя, размера в кабинетах, исполнения Light. Цены указаны в долларах, курс USD автоматически подтягивается с сайта ЦБ РФ раз в сутки, итоговая стоимость выводится в рублях
- **Страница услуг** с интерактивными элементами
- **Форма обратной связи** с валидацией
- **Интерактивная карта** с точками на Яндекс.Картах (кластеризация, панорамы)
- **Адаптивный дизайн** для всех устройств
- **SEO-оптимизация**: автоматический sitemap.xml, robots.txt для поисковых систем

### Административная панель
- 🔐 **Авторизация** с JWT токенами
- 📁 **Управление проектами**: CRUD операции, загрузка изображений, редактор обрезки
- 💰 **Управление прайс-листом**: CRUD позиций, множественная загрузка изображений, характеристики с группировкой, drag & drop сортировка
- 🖼️ **Оптимизация изображений**: автогенерация WebP thumbnails (400×300, 1200×900) при загрузке с настраиваемым кроппингом
- 📧 **CRM для заявок**: фильтрация, статусы, напоминания, заметки, архив
- 📲 **Telegram уведомления**: интерактивные кнопки (обработано, напоминание, открыть в админке)
- ⏰ **Автоматические напоминания**: фоновая задача отправляет уведомления в Telegram в установленное время
- 📊 **Dashboard** с аналитикой и статистикой (топ-5 проектов, топ-5 позиций прайса за 30 дней)
- 🎯 **Система категорий** для проектов
- 📤 **Экспорт контактов** в CSV
- 🗺️ **Управление картой**: точки на Яндекс.Картах, импорт из ссылок, панорамы
- 🔄 **Drag & Drop** сортировка проектов и позиций прайса
- 📱 **Адаптивный интерфейс** для мобильных устройств

## 🛠 Технологический стек

### Backend
- **Go 1.21+** - основной язык бэкенда
- **Gin** - веб-фреймворк
- **GORM** - ORM для работы с БД
- **JWT** - авторизация
- **bcrypt** - хеширование паролей
- **WebP** - оптимизация изображений (github.com/chai2010/webp)

### Frontend
- **HTML5, CSS3, JavaScript** (Vanilla JS)
- **Адаптивная верстка** (медиа-запросы)
- **CSS переменные** для темизации
- **Fetch API** для взаимодействия с backend

### Telegram Bot (Python)
- **Python 3.x** - язык для Telegram бота
- **FastAPI** - веб-фреймворк для API
- **python-telegram-bot** - библиотека для Telegram Bot API с callback handlers
- **Uvicorn** - ASGI сервер
- **httpx** - асинхронные HTTP запросы к Go backend

### База данных
- **PostgreSQL 15** - основная БД
- **pgAdmin** - для администрирования

### DevOps & Infrastructure
- **Docker & Docker Compose** - контейнеризация PostgreSQL
- **Nginx** - reverse proxy с SSL/TLS (Let's Encrypt)
- **systemd** - service management с оптимизированным бинарником
- **fail2ban** - защита от брутфорс-атак SSH
- **UFW Firewall** - настройка сетевой безопасности
- **Automated Backups** - ежедневные бэкапы БД и файлов с ротацией
- **Unattended Upgrades** - автоматические обновления безопасности

## 🚀 Установка и запуск

### Требования
- Go 1.24 или выше
- Docker и Docker Compose (опционально)
- PostgreSQL 15 (или через Docker)

### Быстрый старт

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/yourusername/led-screen-website.git
cd led-screen-website
```

2. **Создайте файл `.env` в директории `backend/`:**
```bash
cd backend
cp .env.example .env
```

3. **Настройте переменные окружения в `backend/.env`:**
```env
DATABASE_URL=postgres://postgres:your_password@localhost:5432/led_display_db?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production
PORT=8080
ENVIRONMENT=development
```

**⚠️ Важно:** Если пароль БД содержит специальные символы (например, `/`), их нужно URL-кодировать:
- `/` → `%2F`
- `@` → `%40`
- `:` → `%3A`

Пример: `postgres://postgres:pass%2Fword@localhost:5432/db`

Полный список переменных см. в [backend/.env.example](backend/.env.example)

4. **Запустите PostgreSQL через Docker:**
```bash
cd ..  # вернуться в корень проекта
docker-compose up -d postgres
```

5. **Установите зависимости Go и запустите приложение:**
```bash
cd backend
go mod download
go run main.go
```

Миграции базы данных выполнятся автоматически при первом запуске.

6. **Создайте администратора:**
```bash
cd backend/cmd/create-admin
go run main.go
```

Следуйте инструкциям для создания первого администратора.

7. **Откройте браузер:**
- Публичная часть: http://localhost:8080
- Админ панель: http://localhost:8080/admin/login

### Запуск через Docker Compose

```bash
docker-compose up -d
```

Приложение будет доступно по адресу http://localhost:8080

## 📁 Структура проекта

```
led-screen-website/
├── backend/               # Go backend
│   ├── cmd/               # Утилиты CLI (create-admin, etc.)
│   ├── internal/          # Пакеты (config, database, handlers, middleware, models, routes)
│   ├── main.go
│   └── .env.example
├── frontend/
│   ├── static/            # CSS, JS, images, uploads
│   └── templates/         # HTML шаблоны (Go templates)
├── telegram-bot/          # Python Telegram уведомления
├── deployment/            # Production deployment
│   ├── led-website.service  # systemd service file
│   ├── deploy.sh          # Automated deployment script
│   └── README.md          # Deployment instructions
├── docs/                  # Документация (API, Architecture, Testing, etc.)
├── .github/workflows/     # CI/CD конфигурация
├── docker-compose.yml
├── check-code.sh/ps1      # Локальная проверка кода
└── test-smoke.sh/ps1      # Smoke tests
```

## 🗄️ База данных

### Основные таблицы:
- **projects** - портфолио проектов
- **gallery_images** - изображения проектов с кроппингом (WebP thumbnails)
- **categories** - категории проектов
- **project_categories** - связь проектов и категорий
- **price_items** - позиции прайс-листа
- **price_images** - изображения позиций прайса (WebP thumbnails, кроппинг)
- **price_specifications** - характеристики позиций прайса с группировкой
- **contacts** - заявки от клиентов
- **contact_notes** - заметки по заявкам
- **admins** - администраторы системы
- **project_view_dailies** - аналитика просмотров проектов по дням
- **price_view_dailies** - аналитика просмотров позиций прайса по дням
- **map_points** - точки на карте (география работы, панорамы)
- **calculator_settings** - настройки калькулятора (размеры кабинета, стоимость комплектующих, кэш курса USD)
- **calculator_pixel_pitches** - шаги пикселя для калькулятора (тип экрана, название, цена модуля)

Миграции БД выполняются автоматически через GORM при первом запуске приложения. Схемы моделей находятся в [backend/internal/models/](backend/internal/models/)

## 🔐 Безопасность

### Аутентификация и Авторизация
- ✅ **JWT токены** с истечением для авторизации
- ✅ **bcrypt** хеширование паролей (cost factor 10)
- ✅ **Middleware** защита всех админских роутов
- ✅ **CORS** настройки с whitelist доменов

### Безопасность данных
- ✅ **SQL injection** защита через GORM prepared statements
- ✅ **Валидация** всех входных данных на бэкенде
- ✅ **Sanitization** пользовательского ввода
- ✅ **Сложные пароли** для БД с URL-кодированием

### Инфраструктурная безопасность
- ✅ **SSH ключи** вместо паролей (Ed25519)
- ✅ **fail2ban** защита от брутфорс-атак (автобан IP)
- ✅ **UFW Firewall** - открыты только 22, 80, 443 порты
- ✅ **Автообновления** безопасности Ubuntu (unattended-upgrades)
- ✅ **Security Headers** в Nginx (X-Frame-Options, X-Content-Type-Options, XSS Protection)
- ✅ **Docker** контейнеры без привилегированного режима
- ✅ **Go приложение** слушает только localhost:8080 (доступ только через Nginx)
- ✅ **PostgreSQL** доступен только с localhost (не из интернета)
- ✅ **pgAdmin** отключен на production (закрыта основная уязвимость)

### Мониторинг и Восстановление
- ✅ **Автоматические бэкапы** БД и файлов (ежедневно в 2:00 UTC)
- ✅ **Ротация бэкапов** (30 дней хранения)
- ✅ **Логирование** всех событий безопасности
- ✅ **Healthcheck** эндпоинты для мониторинга

## 🔍 SEO Оптимизация

- **sitemap.xml** - автоматическая карта сайта (https://s-n-r.ru/sitemap.xml)
- **robots.txt** - правила индексации (https://s-n-r.ru/robots.txt)
- **Open Graph** - мета-теги для красивых превью в соцсетях (VK, Telegram, WhatsApp)
- **Schema.org** - структурированные данные (Organization, LocalBusiness) для rich snippets
- **Meta description & keywords** - для поисковых систем

Добавлен sitemap в [Google Search Console](https://search.google.com/search-console) и [Яндекс.Вебмастер](https://webmaster.yandex.ru)

## 📊 Аналитика и отслеживание

- **Встроенная аналитика**:
  - Отслеживание просмотров проектов и позиций прайса
  - Топ-5 популярных проектов за 30 дней с возможностью сброса статистики
  - Топ-5 просматриваемых позиций прайса за 30 дней
  - Агрегация по дням в отдельных таблицах для оптимизации
  - Client-side трекинг с TTL (10 минут) для предотвращения дублирования
- **Яндекс.Метрика** - вебвизор, карты кликов и скроллинга, аналитика форм, отслеживание источников трафика

## 🧪 Тестирование

**Локальная проверка кода:**
```bash
.\check-code.ps1  # Windows
./check-code.sh   # Linux/macOS
```

**Unit тесты:** 253 автоматических тестов (handlers 51.7%, middleware 100%)
- Middleware 100%, Admin Auth, Admin CRM, Admin Projects (+ Duplicate), Admin Prices CRUD, Admin Images, Admin Sorting, Admin Settings, Admin Promo, Map Points, Helpers, Telegram API, SEO, Helper Functions
```bash
cd backend && go test ./... -v -cover
```

**CI/CD:** GitHub Actions автоматически тестирует, линтует и собирает при push в `main`/`develop`

Подробнее: [docs/TESTING.md](docs/TESTING.md), [docs/LOCAL_CHECKS.md](docs/LOCAL_CHECKS.md)

## 📚 Документация

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Описание архитектуры проекта
- [API.md](docs/API.md) - Документация API эндпоинтов
- [OPERATIONS_GUIDE.md](docs/OPERATIONS_GUIDE.md) - Руководство по операциям на production
- [TESTING.md](docs/TESTING.md) - Документация по тестированию
- [LOCAL_CHECKS.md](docs/LOCAL_CHECKS.md) - Локальная проверка кода перед push
- [CODING_STYLE.md](docs/CODING_STYLE.md) - Соглашения о стиле кодирования

## 🌐 Production

Проект успешно развернут и работает на [https://s-n-r.ru](https://s-n-r.ru)

### Инфраструктура
- **Хостинг:** Beget VPS (2GB RAM, 20GB SSD)
- **ОС:** Ubuntu 22.04 LTS
- **База данных:** PostgreSQL 15 (Docker)
- **Веб-сервер:** Nginx 1.18 с SSL/TLS (Let's Encrypt)
- **Process Manager:** systemd с оптимизированным бинарником
- **Мониторинг:** systemd журналы + healthcheck эндпоинты

### Безопасность Production
- 🛡️ **SSH ключи** вместо паролей (Ed25519)
- 🛡️ **fail2ban** активен (защита от брутфорса)
- 🛡️ **UFW Firewall** настроен (только 22, 80, 443)
- 🛡️ **Автообновления** безопасности (unattended-upgrades)
- 🛡️ **Security Headers** (защита от clickjacking, XSS, MIME-sniffing)
- 🛡️ **Go приложение** слушает только 127.0.0.1:8080
- 🛡️ **PostgreSQL** изолирован (только localhost)
- 🛡️ **Сложные пароли** для всех сервисов

### Бэкапы
- 📦 **Автоматические** ежедневные бэкапы (2:00 UTC)
- 📦 **База данных** + загруженные файлы (изображения)
- 📦 **Ротация** 30 дней (автоматическое удаление старых)
- 📦 **Сжатие** gzip для экономии места

### Deployment
- 🚀 **Automated deployment** через `deployment/deploy.sh`
- 🚀 **Zero-downtime** updates (systemd restart)
- 🚀 **Rollback** возможность через git checkout
- 🚀 **Healthcheck** после деплоя

Подробнее: [docs/OPERATIONS_GUIDE.md](docs/OPERATIONS_GUIDE.md) | [deployment/README.md](deployment/README.md)


## 📝 Лицензия

Proprietary - All Rights Reserved

Этот проект является собственностью и защищен авторским правом. Использование, копирование, модификация или распространение без явного письменного разрешения владельца запрещено.

## 💼 Разработчик

Проект разработан [@alekarah](https://github.com/alekarah) на коммерческой основе.

**Контакт для вопросов:** alekarah.all@gmail.com

## 🙏 Благодарности

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [JWT-Go](https://github.com/golang-jwt/jwt)
- [WebP for Go](https://github.com/chai2010/webp)

---

⭐ Если вам понравился проект, поставьте звездочку!
