## Анализ бизнеса заказчика

Компания Дмитрия Рудкина занимается:

- Продажей, ремонтом и обслуживанием LED дисплеев
- Работает в Санкт-Петербурге и Ленинградской области
- Изготавливает металлоконструкции и фундаментные блоки
- Использует систему управления Novastar
- Обслуживает 172 рекламных конструкции (204 LED дисплея)
- Обслуживает 83 АЗС (124 LED дисплея)
- Специализируется на интерьерных экранах для Питера и уличных для регионов

### Технологический стек:

- **Backend**: Go (Gin фреймворк) - для API, маршрутизации, обработки форм
- **Python**: для обработки изображений, аналитики, возможно ML для рекомендаций (не реализовано)
- **Frontend**: HTML5, CSS3, JavaScript
- **База данных**: PostgreSQL
- **Деплой**: Docker + возможно Nginx
  

## Детальный план разработки

### Этап 1: Архитектура и настройка проекта

**1.1 Структура проекта**

```
led-screen-website/
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── database/
│   │   │   └── database.go
│   │   ├── handlers/
│   │   │   ├── admin_contacts.go
│   │   │   ├── admin_dashboard.go
│   │   │   ├── admin_images.go
│   │   │   ├── admin_projects_crud.go
│   │   │   ├── admin_sorting.go
│   │   │   └── handlers.go
│   │   ├── models/
│   │   │   └── models.go
│   │   ├── routes/
│   │   │   └── routes.go
│   │   └── services/
│   ├── main.go
│   ├── go.sum
│   └── go.mod
├── python-services/
│   ├── image_processor/
│   └── analytics/
├── frontend/
│   ├── static/
│   │   ├── uploads/
│   │   ├── css/
│   │   │   ├── admin-base.css
│   │   │   ├── admin-forms.css
│   │   │   ├── admin-modals.css
│   │   │   ├── admin-projects.css
│   │   │   ├── crop-editor.css
│   │   │   ├── led-effect.css
│   │   │   ├── public-base.css
│   │   │   ├── public-contact.css
│   │   │   ├── public-footer.css
│   │   │   ├── public-forms.css
│   │   │   ├── public-home.css
│   │   │   ├── public-privacy.css
│   │   │   ├── public-projects.css
│   │   │   ├── public-responsive.css
│   │   │   └── public-services.css
│   │   ├── js/
│   │   │   ├── admin-base.js
│   │   │   ├── admin-contacts.js
│   │   │   ├── admin-projects-creation.js
│   │   │   ├── admin-projects-editing.js
│   │   │   ├── admin-projects-images.js
│   │   │   ├── admin-projects-init.js
│   │   │   ├── contact-form.js
│   │   │   ├── crop-editor-api.js
│   │   │   ├── crop-editor-core.js
│   │   │   ├── crop-editor-presets.js
│   │   │   ├── crop-editor-ui.js
│   │   │   ├── project-modal.js
│   │   │   ├── project-sorting.js
│   │   │   └── services-steps.js
│   │   └── images/
│   │   │   ├── diod.png
│   │   │   └── favicon.png
│   └── templates/
│   │   ├── admin_base.html
│   │   ├── admin_contacts.html
│   │   ├── admin_dashboard.html
│   │   ├── admin_projects.html
│   │   ├── contact.html
│   │   ├── index.html
│   │   ├── privacy.html
│   │   ├── projects.html
│   │   ├── public_base.html
│   │   └── services.html
├── init.sql
├── .env.example
├── docker-compose.yml
└── README.md
```

**1.2 База данных (PostgreSQL)** Основные таблицы:

- `projects` (портфолио проектов)
- `services` (услуги)
- `contacts` (заявки)
- `gallery_images` (фотографии)
- `categories` (категории проектов)

### Этап 2: Backend на Go

**2.1 Основные компоненты:**

- REST API для получения проектов, услуг
- Обработка форм обратной связи
- Админка для управления контентом
- Загрузка и обработка изображений

**2.2 Эндпоинты:**

```
GET /api/projects - список проектов с фильтрацией
GET /api/projects/:id - детали проекта
GET /api/services - список услуг
POST /api/contact - отправка заявки
GET /admin/* - админ панель
```

### Этап 3: Python сервисы

**3.1 Обработка изображений:**

- Автоматическое создание превью
- Оптимизация размеров
- Создание водяных знаков

**3.2 Аналитика:**

- Подсчет просмотров проектов
- Анализ популярных услуг
- Метрики конверсии

### Этап 4: Frontend

**4.1 Дизайн-система:**

- Светлая тема (в отличие от темной ekranika.ru)
- Корпоративные цвета: синий (`#4A90E2`), белый, светло-серый
- Современная типографика
- Адаптивная верстка

**4.2 Ключевые компоненты:**

- Героический блок с призывом к действию
- Галерея проектов с фильтрацией
- Интерактивная карта услуг
- Форма заявки с валидацией
- Слайдеры для портфолио

### Этап 5: Контент и наполнение

**5.1 Обработка фотографий заказчика:**

- Категоризация по типам проектов
- Создание описаний для каждого проекта
- Оптимизация для веба

**5.2 Текстовый контент:**

- SEO-оптимизированные описания услуг
- Преимущества компании
- Технические характеристики
