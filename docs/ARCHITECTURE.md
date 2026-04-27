# 🏗️ Архитектура проекта

> Описание архитектуры LED Screen Website - корпоративного сайта для компании по продаже и обслуживанию LED дисплеев

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

**Технологический стек:** См. [README.md](../README.md#-технологический-стек)

---

## Backend архитектура

### 📁 Структура пакетов

```
backend/
├── cmd/                           # Утилиты командной строки
│   ├── create-admin/              # Создание администраторов
│   ├── regenerate-thumbnails/     # Регенерация thumbnails для всех изображений
│   └── check-db/                  # Проверка состояния БД
│       └── main.go
├── internal/                      # Внутренние пакеты (private)
│   ├── config/                    # Конфигурация приложения
│   │   └── config.go              # Загрузка переменных окружения
│   ├── database/                  # Работа с базой данных
│   │   └── database.go            # Подключение, миграции, seed
│   ├── handlers/                  # HTTP обработчики
│   │   ├── handlers.go            # Публичные страницы
│   │   ├── seo.go                 # SEO handlers (sitemap.xml, robots.txt)
│   │   ├── admin_auth.go          # Аутентификация
│   │   ├── admin_dashboard.go     # Dashboard с аналитикой (топ-5 проектов/прайсов)
│   │   ├── admin_projects_crud.go # CRUD проектов
│   │   ├── admin_prices_crud.go   # CRUD позиций прайс-листа
│   │   ├── admin_actions.go       # Действия админа (контакты)
│   │   ├── admin_images.go        # Загрузка изображений и автогенерация thumbnails
│   │   ├── image_processor.go     # Обработка изображений (thumbnails, crop, WebP)
│   │   ├── admin_sorting.go       # Сортировка проектов
│   │   ├── admin_map_points.go    # CRUD точек на карте + импорт из Яндекс.Карт
│   │   ├── calculator.go          # Калькулятор: курс ЦБ, кэш, данные для шаблона
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

**Инициализация (main.go):**
```
Load .env → Config → DB Connect → Migrate → Handlers (DI) → Routes → Run Server
```

**Обработка запроса:**
```
HTTP Request → Gin Router → Middleware (auth) → Handler → GORM → PostgreSQL → Response
```

**Типы маршрутов:**
- Публичные: `/`, `/projects`, `/services`, `/prices`, `/contact`
- API: `/api/projects`, `/api/contact`, `/api/track/project-view/:id`, `/api/track/price-view/:id`, `/api/calculator`
- Админ: `/admin/login` (открытый), `/admin/*` (JWT защита)
- Карта: `/admin/map-points/*` (CRUD точек + импорт из Яндекс.Карт)

### 🔐 Система авторизации

**Аутентификация:**
```
POST /admin/login → bcrypt verify → JWT token (HS256) → HTTP-only cookie → redirect /admin/
```

**Защищенные роуты:**
```
Request → AuthMiddleware → validate JWT → extract claims (admin_id, username) → gin.Context → Handler
```

**JWT Claims:** `{ user_id, username, exp }`

### 📦 Модели данных

**Основные модели** (internal/models/models.go):

| Модель | Описание | Ключевые поля |
|--------|----------|---------------|
| `Project` | Проекты портфолио | Title, Slug, Description, Location, Size, Featured, SortOrder, ViewCount |
| `Category` | Категории проектов | Name, Slug, Description |
| `Image` | Изображения проектов | ProjectID, Filename, FilePath, ThumbnailSmallPath, ThumbnailMediumPath, CropX/Y/Scale, IsPrimary |
| `PriceItem` | Позиции прайс-листа | Title, Description, PriceFrom, HasSpecifications, IsActive, SortOrder, Category (indoor/outdoor/innovative/other), IsLight |
| `PriceImage` | Изображения позиций прайса | PriceItemID, Filename, FilePath, ThumbnailSmallPath, ThumbnailMediumPath, CropX/Y/Scale |
| `PriceSpecification` | Характеристики позиций прайса | PriceItemID, SpecGroup, SpecKey, SpecValue, SpecOrder (группировка) |
| `ContactForm` | Заявки клиентов | Name, Phone, Email, Status, ArchivedAt, RemindAt |
| `ContactNote` | Заметки по заявкам | ContactID, Text, Author |
| `Admin` | Администраторы | Username, PasswordHash, IsActive, LastLoginAt |
| `ProjectViewDaily` | Просмотры проектов по дням | ProjectID, Day, Views (аналитика) |
| `PriceViewDaily` | Просмотры позиций прайса по дням | PriceItemID, Day, Views (аналитика) |
| `Service` | Услуги компании | Name, Slug, Description, Icon, Featured |
| `Settings` | Настройки сайта | Key, Value, Type |
| `MapPoint` | Точки на карте (география работы) | Title, Description, Latitude, Longitude, PanoramaURL, SortOrder, IsActive |
| `CalculatorSettings` | Настройки калькулятора | CabWidth, CabHeight, Commutation, Card, Power, UsdRate, UsdRateAt (кэш курса ЦБ) |
| `CalculatorPixelPitch` | Шаги пикселя для калькулятора | ScreenType (indoor/outdoor), Name, ModulePrice, SortOrder, IsActive |

**Связи между таблицами:**
- `Project` ↔ `Category` (many-to-many через `project_categories`)
- `Project` → `Image` (one-to-many)
- `Project` → `ProjectViewDaily` (one-to-many с CASCADE DELETE)
- `PriceItem` → `PriceImage` (one-to-many с CASCADE DELETE)
- `PriceItem` → `PriceSpecification` (one-to-many с CASCADE DELETE)
- `PriceItem` → `PriceViewDaily` (one-to-many с CASCADE DELETE)
- `ContactForm` → `ContactNote` (one-to-many)

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
│   │   ├── admin-prices.css     # Прайс-лист админки
│   │   ├── admin-contacts.css   # Контакты админки
│   │   ├── admin-login.css      # Страница входа
│   │   ├── admin-modals.css     # Модальные окна админки
│   │   ├── crop-editor.css      # Редактор обрезки изображений
│   │   ├── public-prices.css    # Прайс-лист (публичная)
│   │   └── public-responsive.css # Медиа-запросы публичной части
│   ├── js/                      # JavaScript модули
│   │   ├── admin-base.js        # Базовая функциональность админки
│   │   ├── admin-projects-*.js  # Модули управления проектами
│   │   ├── admin-prices.js      # CRUD прайс-листа
│   │   ├── admin-prices-images.js # Управление изображениями прайса
│   │   ├── admin-contacts-*.js  # Модули управления контактами
│   │   ├── crop-editor.js       # Редактор обрезки
│   │   ├── price-modal.js       # Модальные окна прайс-листа (публичная)
│   │   ├── prices-accordion.js  # Аккордеон характеристик (публичная)
│   │   ├── grid-center.js       # Универсальное центрирование неполных рядов в гридах
│   │   ├── projects-load-more.js # «Показать ещё» для проектов
│   │   ├── prices-filter.js     # Фильтрация позиций прайса по категории и Light-исполнению (JS, без перезагрузки)
│   │   ├── prices-load-more.js  # «Показать ещё» для прайс-листа
│   │   ├── projects-filter.js   # Фильтрация проектов по категориям
│   │   ├── admin-map-points.js  # CRUD точек на карте (админка)
│   │   ├── contact-map.js       # Яндекс.Карта на странице контактов
│   │   └── vendor/              # Сторонние библиотеки (Sortable.js, Chart.js)
│   ├── images/                  # Статические изображения
│   └── uploads/                 # Загруженные файлы (gitignore)
└── templates/                   # HTML шаблоны (Go templates)
    ├── public_base.html         # Базовый layout (публичный)
    ├── admin_base.html          # Базовый layout (админка)
    ├── index.html               # Главная страница
    ├── projects.html            # Портфолио
    ├── prices.html              # Прайс-лист с модальными окнами
    ├── services.html            # Услуги
    ├── contact.html             # Контакты
    ├── admin_dashboard.html     # Dashboard админки (топ-5 проектов/прайсов)
    ├── admin_projects.html      # Управление проектами
    ├── admin_prices.html        # Управление прайс-листом
    ├── admin_contacts.html      # Управление заявками
    ├── admin_calculator.html    # Управление калькулятором (курс, константы, шаги пикселя)
    ├── admin_map_points.html    # Управление точками на карте
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
- **Module Pattern**: `const ContactsAPI = { updateStatus: async (id, status) => {...} }`
- **Event-Driven**: `DOMContentLoaded → initFilters() + initBulkActions() + initModal()`
- **API Abstraction**: `fetchWithAuth(url, {credentials: 'include'}) → response.json()`

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
// Frontend: fetch(`/admin/contacts/${id}/status`, {method: 'POST', credentials: 'include', body: JSON.stringify({status})})
// Backend: c.BindJSON(&input) → h.db.Update("status", input.Status) → c.JSON(200, gin.H{"success": true})
```

---

## База данных

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

## Система обработки изображений

**Автоматическая генерация thumbnails:**
- **Small** (400×300px) - карточки проектов, главная страница
- **Medium** (1200×900px) - модальное окно галереи
- **Original** - кнопка "Открыть изображение"
- Формат: JPEG (quality 85%), PNG с оптимизацией
- Библиотека: `github.com/disintegration/imaging` (Lanczos resampling)

**Crop Editor:**
- Live preview через CSS transform
- Параметры: CropX/Y (0-100%), CropScale (0.5-3.0x)
- Регенерация thumbnails с применением кропа на сервере
- Fallback к оригиналу для старых изображений

**API:**
- `POST /admin/upload-images` - загрузка и генерация thumbnails
- `POST /admin/images/{id}/crop` - обновление кропа и регенерация
- `DELETE /admin/images/{id}` - удаление оригинала + thumbnails

**Утилита:** `backend/cmd/regenerate-thumbnails` - массовая регенерация

---

## Паттерны проектирования

- **MVC**: Models (`models.go`) + Views (`templates/*.html`) + Controllers (`handlers/*.go`)
- **Dependency Injection**: `handlers.New(db)` - DI базы данных в handlers
- **Middleware**: `admin.Use(middleware.AuthMiddleware())` - цепочка middleware
- **Repository**: GORM абстрагирует SQL (`h.db.Where().Find()`, `h.db.Preload()`)
- **Configuration**: `config.Load()` - централизованная конфигурация
- **Factory**: `handlers.New(db)` - фабрика handlers

---

## Безопасность

**Реализованные меры:**
- JWT токены (HTTP-only cookies, expiration)
- bcrypt хеширование паролей (cost 10)
- Middleware защита всех админских роутов
- SQL Injection защита (GORM prepared statements)
- XSS защита (Go templates auto-escaping)
- Валидация файлов (размер 10MB, MIME типы, UUID имена)
- Логирование всех запросов (Gin Logger + GORM)

**Production рекомендации:**
- HTTPS с TLS сертификатами
- CORS ограничения
- Rate Limiting (DDoS защита)
- Secure + HttpOnly флаги cookies
- Регулярные бэкапы БД

---

## Производительность

**База данных:**
- Connection pooling (MaxOpenConns: 20, MaxIdleConns: 10)
- Индексы на часто запрашиваемые колонки
- Eager loading (Preload) для связанных данных
- Агрегация просмотров по дням (ProjectViewDaily)

**Backend:**
- Gin Release Mode (production)
- Recovery middleware (graceful panic recovery)

**Frontend:**
- Defer загрузка JavaScript
- Автоматическая генерация thumbnails (400×300 для карточек, 1200×900 для галереи)
- Lazy loading изображений с fallback к оригиналу
- Debounce для поиска/фильтрации

**Кеширование:**
- Статические файлы через Nginx (production)
- Browser cache (CSS/JS/изображения)
- GORM PrepareStmt (SQL кеширование)

**Мониторинг:**
- Healthcheck: `GET /healthz → 200 OK`
- Метрики: Response time, DB pool usage, просмотры проектов

---

## 📚 Дополнительные ресурсы

- [API.md](API.md) - Документация API эндпоинтов
- [DEPLOYMENT.md](DEPLOYMENT.md) - Инструкции по деплою
- [README.md](../README.md) - Общая информация о проекте

---

**Версия документа**: 1.0
**Последнее обновление**: Февраль 2026
