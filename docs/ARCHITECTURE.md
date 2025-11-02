# 🏗️ Архитектура проекта

> Описание архитектуры LED Screen Website - корпоративного сайта для компании по продаже и обслуживанию LED дисплеев

## 📋 Содержание

- [Общий обзор](#общий-обзор)
- [Backend архитектура](#backend-архитектура)
- [Frontend архитектура](#frontend-архитектура)
- [База данных](#база-данных)
- [Паттерны проектирования](#паттерны-проектирования)
- [Безопасность](#безопасность)
- [Производительность](#производительность)

---

## Общий обзор

Проект построен по классической клиент-серверной архитектуре:

```
┌─────────────────────────────────────────────────────────┐
│                    Клиент (Browser)                     │
│  ┌─────────────────────┐  ┌──────────────────────────┐  │
│  │   Публичная часть   │  │    Админ панель (SPA)    │  │
│  │  (HTML/CSS/JS)      │  │     (HTML/CSS/JS)        │  │
│  └─────────────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────┐
│                    Backend (Go + Gin)                   │
│  ┌────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │  Handlers  │──│ Middleware  │──│  Routes          │  │
│  └────────────┘  └─────────────┘  └──────────────────┘  │
│  ┌────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │   Models   │──│  Database   │──│  Config          │  │
│  └────────────┘  └─────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────┐
│                  PostgreSQL Database                    │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │   Projects   │  │   Contacts   │  │    Images     │  │
│  │  Categories  │  │    Admins    │  │   Settings    │  │
│  └──────────────┘  └──────────────┘  └───────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Технологический стек

- **Backend**: Go 1.21+, Gin Web Framework, GORM ORM
- **Frontend**: Vanilla JavaScript (ES6+), HTML5, CSS3
- **Database**: PostgreSQL 15
- **Auth**: JWT (JSON Web Tokens)
- **DevOps**: Docker, Docker Compose

---

## Backend архитектура

### 📁 Структура пакетов

```
backend/
├── cmd/                           # Утилиты командной строки
│   └── create-admin/              # Создание администраторов
│       └── main.go
├── internal/                      # Внутренние пакеты (private)
│   ├── config/                    # Конфигурация приложения
│   │   └── config.go              # Загрузка переменных окружения
│   ├── database/                  # Работа с базой данных
│   │   └── database.go            # Подключение, миграции, seed
│   ├── handlers/                  # HTTP обработчики
│   │   ├── handlers.go            # Публичные страницы
│   │   ├── admin_auth.go          # Аутентификация
│   │   ├── admin_dashboard.go     # Dashboard с аналитикой
│   │   ├── admin_projects_crud.go # CRUD проектов
│   │   ├── admin_actions.go       # Действия админа (контакты)
│   │   ├── admin_images.go        # Загрузка и обработка изображений
│   │   ├── admin_sorting.go       # Сортировка проектов
│   │   ├── admin_pages.go         # Рендеринг админских страниц
│   │   └── admin_helpers.go       # Вспомогательные функции
│   ├── middleware/                # HTTP middleware
│   │   └── auth.go                # JWT авторизация
│   ├── models/                    # Модели данных (ORM)
│   │   └── models.go              # Все структуры БД
│   ├── routes/                    # Маршрутизация
│   │   └── routes.go              # Настройка роутов
│   └── services/                  # Бизнес-логика (будущее)
└── main.go                        # Точка входа приложения
```

### 🔄 Жизненный цикл запроса

1. **Инициализация** (main.go):
   ```go
   godotenv.Load()           // Загрузка .env
   cfg := config.Load()      // Конфигурация
   db := database.Connect()  // Подключение к БД
   database.Migrate()        // Миграции
   handlers.New(db)          // Создание handlers с DI
   routes.Setup()            // Настройка роутов
   router.Run()              // Запуск сервера
   ```

2. **Обработка запроса**:
   ```
   HTTP Request
      ↓
   Gin Router (routes/routes.go)
      ↓
   Middleware (auth.go) - если защищенный роут
      ↓
   Handler (handlers/*.go)
      ↓
   GORM Query (models через database)
      ↓
   PostgreSQL Database
      ↓
   Response (JSON или HTML template)
   ```

3. **Типы маршрутов**:
   - **Публичные**: `/`, `/projects`, `/services`, `/contact`, `/privacy`
   - **API публичное**: `/api/projects`, `/api/contact`
   - **Админ незащищенные**: `/admin/login` (POST/GET)
   - **Админ защищенные**: `/admin/*` (требуют JWT токен)

### 🔐 Система авторизации

**Flow аутентификации:**

```
1. User → POST /admin/login (username, password)
2. Handler проверяет bcrypt hash пароля
3. Генерируется JWT token (HS256)
4. Token сохраняется в HTTP-only cookie "admin_token"
5. Редирект на /admin/

Защищенные роуты:
1. Request → AuthMiddleware
2. Извлечение токена из cookie
3. Валидация JWT (подпись + expiration)
4. Извлечение claims (admin_id, username)
5. Сохранение в gin.Context
6. Передача управления handler'у
```

**JWT Claims структура:**
```go
type JWTClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}
```

### 📦 Модели данных

**Основные модели** (internal/models/models.go):

| Модель | Описание | Ключевые поля |
|--------|----------|---------------|
| `Project` | Проекты портфолио | Title, Slug, Description, Location, Size, Featured, SortOrder, ViewCount |
| `Category` | Категории проектов | Name, Slug, Description |
| `Image` | Изображения проектов | ProjectID, Filename, FilePath, CropX/Y/Scale |
| `ContactForm` | Заявки клиентов | Name, Phone, Email, Status, ArchivedAt, RemindAt |
| `ContactNote` | Заметки по заявкам | ContactID, Text, Author |
| `Admin` | Администраторы | Username, PasswordHash, IsActive, LastLoginAt |
| `ProjectViewDaily` | Просмотры по дням | ProjectID, Day, Views |
| `Service` | Услуги компании | Name, Slug, Description, Icon, Featured |
| `Settings` | Настройки сайта | Key, Value, Type |

**Связи между таблицами:**
- `Project` ↔ `Category` (many-to-many через `project_categories`)
- `Project` → `Image` (one-to-many)
- `ContactForm` → `ContactNote` (one-to-many)
- `Project` → `ProjectViewDaily` (one-to-many с CASCADE DELETE)

---

## Frontend архитектура

### 📁 Структура файлов

```
frontend/
├── static/
│   ├── css/                     # Стили
│   │   ├── public-base.css      # Базовые стили (публичная часть)
│   │   ├── public-vars.css      # CSS переменные (публичная)
│   │   ├── admin-base.css       # Базовые стили (админка)
│   │   ├── admin-vars.css       # CSS переменные (админка)
│   │   ├── admin-forms.css      # Формы админки
│   │   ├── admin-projects.css   # Проекты админки
│   │   ├── admin-contacts.css   # Контакты админки
│   │   ├── admin-login.css      # Страница входа
│   │   ├── crop-editor.css      # Редактор обрезки изображений
│   │   └── modal.css            # Модальные окна
│   ├── js/                      # JavaScript модули
│   │   ├── admin-base.js        # Базовая функциональность админки
│   │   ├── admin-projects-*.js  # Модули управления проектами
│   │   ├── admin-contacts-*.js  # Модули управления контактами
│   │   ├── crop-editor.js       # Редактор обрезки
│   │   └── vendor/              # Сторонние библиотеки (Sortable.js)
│   ├── images/                  # Статические изображения
│   └── uploads/                 # Загруженные файлы (gitignore)
└── templates/                   # HTML шаблоны (Go templates)
    ├── public_base.html         # Базовый layout (публичный)
    ├── admin_base.html          # Базовый layout (админка)
    ├── index.html               # Главная страница
    ├── projects.html            # Портфолио
    ├── services.html            # Услуги
    ├── contact.html             # Контакты
    ├── admin_dashboard.html     # Dashboard админки
    ├── admin_projects.html      # Управление проектами
    ├── admin_contacts.html      # Управление заявками
    └── admin_login.html         # Страница входа
```

### 🎨 CSS архитектура

**Принципы организации стилей:**

1. **CSS Variables (Custom Properties)**:
   - `public-vars.css` - цвета, отступы, размеры для публичной части
   - `admin-vars.css` - переменные для админ панели (brand colors, spacing, shadows)

2. **Модульность**:
   - Каждая секция имеет свой CSS файл
   - Базовые стили отделены от специфичных

3. **Адаптивность**:
   - Mobile-first approach
   - Брейкпоинты: 1150px, 1024px, 900px, 768px, 480px
   - Flexbox и CSS Grid для раскладки

4. **BEM-подобная методология**:
   ```css
   .contacts-toolbar           /* Блок */
   .contacts-toolbar__actions  /* Элемент */
   .project-item--focus        /* Модификатор */
   ```

### 📜 JavaScript архитектура

**Модульная структура** (разделение по ответственности):

**Админ панель - Проекты:**
- `admin-projects-creation.js` - создание нового проекта
- `admin-projects-editing.js` - редактирование проекта
- `admin-projects-drag.js` - drag & drop сортировка
- `crop-editor.js` - редактор обрезки изображений

**Админ панель - Контакты:**
- `admin-contacts-api.js` - API запросы
- `admin-contacts-ui.js` - UI компоненты
- `admin-contacts-filters.js` - фильтрация и поиск
- `admin-contacts-modal.js` - модальные окна
- `admin-contacts-bulk.js` - массовые операции
- `admin-contacts-notes.js` - система заметок
- `admin-contacts-shared.js` - общие функции
- `admin-contacts-init.js` - инициализация

**Архитектурные паттерны в JS:**

1. **Module Pattern**:
   ```javascript
   const ContactsAPI = {
       updateStatus: async (id, status) => { ... },
       archiveContact: async (id) => { ... }
   };
   ```

2. **Event-Driven Architecture**:
   ```javascript
   document.addEventListener('DOMContentLoaded', () => {
       initFilters();
       initBulkActions();
       initModal();
   });
   ```

3. **API Abstraction Layer**:
   ```javascript
   // admin-contacts-api.js
   async function fetchWithAuth(url, options) {
       const response = await fetch(url, {
           ...options,
           credentials: 'include'  // JWT cookie
       });
       return response.json();
   }
   ```

### 🔄 Взаимодействие Frontend ↔ Backend

**Публичная часть** (Server-Side Rendering):
```
Browser Request → Gin Handler → Go Template → HTML Response
```

**Админ панель** (Single Page Application approach):
```
1. Страница рендерится через Go Template
2. JavaScript загружается асинхронно (defer)
3. API запросы через Fetch API (JSON)
4. Динамическое обновление DOM
```

**Пример API взаимодействия:**
```javascript
// Фронтенд (admin-contacts-api.js)
async function updateContactStatus(id, status) {
    const response = await fetch(`/admin/contacts/${id}/status`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        credentials: 'include',  // JWT cookie
        body: JSON.stringify({ status })
    });
    return response.json();
}

// Бэкенд (admin_actions.go)
func (h *Handlers) UpdateContactStatus(c *gin.Context) {
    var input struct { Status string `json:"status"` }
    c.BindJSON(&input)

    contactID := c.Param("id")
    h.db.Model(&models.ContactForm{}).
        Where("id = ?", contactID).
        Update("status", input.Status)

    c.JSON(200, gin.H{"success": true})
}
```

---

## База данных

### 🗄️ Схема таблиц

```sql
-- Основные таблицы
categories           (id, name, slug, description)
projects             (id, title, slug, description, location, size,
                      pixel_pitch, completed, featured, view_count,
                      sort_order, created_at, updated_at)
images               (id, project_id, filename, file_path,
                      crop_x, crop_y, crop_scale, sort_order)
contact_forms        (id, name, phone, email, company, project_type,
                      message, status, created_at, archived_at,
                      remind_at, remind_flag)
contact_notes        (id, contact_id, text, author, created_at)
admins               (id, username, password_hash, email, is_active,
                      last_login_at, created_at, updated_at)
project_view_dailies (id, project_id, day, views, created_at, updated_at)
services             (id, name, slug, short_desc, description,
                      icon, featured, sort_order)
settings             (id, key, value, type, created_at, updated_at)

-- Промежуточные таблицы
project_categories   (project_id, category_id)
```

### 🔗 Связи и ограничения

**Foreign Keys:**
```sql
images.project_id → projects.id (CASCADE DELETE)
contact_notes.contact_id → contact_forms.id
project_view_dailies.project_id → projects.id (CASCADE DELETE)
```

**Индексы** (для оптимизации запросов):
```sql
-- Основные индексы
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_projects_featured ON projects(featured);
CREATE INDEX idx_projects_sort_order ON projects(sort_order);

-- Контакты
CREATE INDEX idx_contacts_status ON contact_forms(status);
CREATE INDEX idx_contacts_created_at ON contact_forms(created_at);
CREATE INDEX idx_contacts_archived_at ON contact_forms(archived_at);
CREATE INDEX idx_contacts_remind_at ON contact_forms(remind_at);

-- Просмотры проектов
CREATE UNIQUE INDEX uniq_project_day
    ON project_view_dailies(project_id, day);
CREATE INDEX idx_pvd_project ON project_view_dailies(project_id);
CREATE INDEX idx_pvd_day ON project_view_dailies(day);
```

### 📊 Миграции и seed данные

**Автоматические миграции** (GORM AutoMigrate):
- Выполняются при каждом запуске приложения
- Создают таблицы и обновляют структуру без потери данных
- Местоположение: `internal/database/database.go` → `Migrate()`

**Seed данные** (начальное наполнение):
- 6 базовых категорий (Рекламные щиты, АЗС, Торговые центры, и т.д.)
- 4 базовые услуги (Продажа интерьерных, уличных, обслуживание, металлоконструкции)
- Настройки сайта (название, телефон, email, SEO meta)
- Местоположение: `internal/database/database.go` → `seedInitialData()`

---

## Паттерны проектирования

### 1. MVC (Model-View-Controller)

```
Models      → internal/models/models.go
Views       → frontend/templates/*.html
Controllers → internal/handlers/*.go
```

### 2. Dependency Injection

```go
// main.go
db := database.Connect(cfg)
handlers := handlers.New(db)  // DI базы данных

// handlers/handlers.go
type Handlers struct {
    db *gorm.DB  // Зависимость
}

func New(db *gorm.DB) *Handlers {
    return &Handlers{db: db}
}
```

### 3. Middleware Pattern

```go
// routes/routes.go
admin := router.Group("/admin")
admin.Use(middleware.AuthMiddleware())  // Цепочка middleware
{
    admin.GET("/", h.AdminDashboard)
    // ...
}
```

### 4. Repository Pattern (частично)

GORM выступает в роли Repository, абстрагируя SQL:
```go
// Вместо SQL запросов
h.db.Where("status = ?", "new").Find(&contacts)
h.db.Preload("Images").Find(&projects)
```

### 5. Configuration Pattern

Централизованная конфигурация через `config/config.go`:
```go
cfg := config.Load()  // Один источник правды
db := database.Connect(cfg)
```

### 6. Factory Pattern

```go
// handlers/handlers.go
func New(db *gorm.DB) *Handlers {
    return &Handlers{db: db}  // Фабрика handlers
}
```

---

## Безопасность

### 🔐 Меры безопасности

1. **Аутентификация и авторизация**:
   - JWT токены с истечением (expiration)
   - HTTP-only cookies (защита от XSS)
   - bcrypt хеширование паролей (cost factor 10)
   - Middleware проверка на всех защищенных роутах

2. **Защита от атак**:
   - **SQL Injection**: GORM использует prepared statements
   - **XSS**: Экранирование в Go templates (auto-escaping)
   - **CSRF**: SameSite cookie attribute (опционально)
   - **Path Traversal**: Валидация путей при загрузке файлов

3. **Валидация данных**:
   ```go
   // Валидация в handlers
   if input.Status != "new" && input.Status != "in_progress" {
       c.JSON(400, gin.H{"error": "Invalid status"})
       return
   }
   ```

4. **Безопасность файлов**:
   - Ограничение размера загружаемых файлов (10MB)
   - Проверка MIME типов изображений
   - Генерация уникальных имен файлов (UUID)
   - Хранение вне корня веб-сервера (`uploads/` не в public)

5. **Логирование**:
   - Gin Logger middleware для всех запросов
   - GORM логирование SQL запросов (уровень настраивается)

### 🔒 Рекомендации для production

1. Использовать HTTPS (TLS сертификаты)
2. Настроить CORS ограничения
3. Добавить Rate Limiting (защита от DDoS)
4. Использовать Secure и HttpOnly флаги для cookies
5. Включить GORM PrepareStmt для кеширования запросов
6. Настроить регулярные бэкапы БД

---

## Производительность

### ⚡ Оптимизации

1. **База данных**:
   - Connection pooling (MaxOpenConns: 20, MaxIdleConns: 10)
   - Индексы на часто запрашиваемые колонки
   - Eager loading (Preload) для связанных данных
   - Агрегация просмотров по дням (ProjectViewDaily)

2. **Backend**:
   - Gin Release Mode для production
   - Gin Recovery middleware (graceful panic recovery)
   - Ленивая загрузка шаблонов

3. **Frontend**:
   - Defer загрузка JavaScript
   - CSS минимизация (production)
   - Lazy loading изображений (intersection observer)
   - Debounce для поиска и фильтрации

4. **Кеширование**:
   - Статические файлы через Nginx (production)
   - Browser cache для CSS/JS/изображений
   - GORM PrepareStmt для кеширования SQL

### 📊 Мониторинг

**Healthcheck endpoint:**
```
GET /healthz → 200 OK
```

**Метрики для отслеживания:**
- Response time (Gin Logger)
- Database connection pool usage
- Количество заявок за период
- Просмотры проектов (ProjectViewDaily)

---

## 🔄 Deployment Flow

```
Development → Testing → Staging → Production

1. Локальная разработка (go run main.go)
2. Docker сборка (docker-compose up)
3. CI/CD pipeline (опционально)
4. Production deployment (см. DEPLOYMENT.md)
```

---

## 📚 Дополнительные ресурсы

- [API.md](API.md) - Документация API эндпоинтов
- [DEPLOYMENT.md](DEPLOYMENT.md) - Инструкции по деплою
- [README.md](../README.md) - Общая информация о проекте

---

**Версия документа**: 1.0
**Последнее обновление**: Ноябрь 2024
