# 📡 API Документация

> Документация API эндпоинтов LED Screen Website

---

## Общая информация

**Base URL**: `http://localhost:8080` (development) / `https://yourdomain.com` (production)

**Форматы данных**:
- Request: `application/json` или `multipart/form-data` (для загрузки файлов)
- Response: `application/json` или `text/html` (для страниц)

**Часовой пояс**: Europe/Moscow (MSK, UTC+3)

**Кодировка**: UTF-8

---

## Аутентификация

**JWT токены** хранятся в HTTP-only cookies (`admin_token`).

**Вход:** `POST /admin/login` (username, password) → JWT cookie → redirect `/admin/`
**Выход:** `GET /admin/logout` → clear cookie → redirect `/admin/login`

**Middleware:** Все `/admin/*` (кроме `/admin/login`) проверяют JWT автоматически.

**Errors:** `401` - неверные credentials / деактивирован / истек токен

---

## Публичные API

### 1. Получить список проектов

`GET /api/projects`

**Query:** `page` (default: 1), `limit` (default: 12), `category` (slug)

**Response** (200): `{projects: [{id, title, slug, description, location, size, pixel_pitch, featured, view_count, categories: [], images: []}], total, page, limit}`

### 2. Отправить заявку

`POST /api/contact`

**Request:** `{name*, phone*, email, company, project_type, message}` (* required)
**Response** (200): `{message: "Заявка успешно отправлена!"}`
**Errors:** `400` - имя/телефон обязательны

### 3. Трекинг просмотра проекта

`POST /api/track/project-view/:id`

**Response** (200): `{ok: true}`
**Note:** Агрегирует просмотры по дням (MSK) в `project_view_dailies`. Client-side TTL 10 минут.

### 4. Трекинг просмотра позиции прайса

`POST /api/track/price-view/:id`

**Response** (200): `{ok: true}`
**Note:** Агрегирует просмотры по дням (MSK) в `price_view_dailies`. Client-side TTL 10 минут.

### 5. Статистика заявок за 7 дней

`GET /api/admin/contacts-7d`

**Response** (200): `[{day: "2024-11-01", count: 3}, ...]`

### 6. Данные калькулятора стоимости

`GET /api/calculator`

**Response** (200):
```json
{
  "settings": {
    "cab_width": 640, "cab_height": 640,
    "commutation": 25.0, "card": 30.0, "power": 15.0,
    "usd_rate": 92.5, "usd_rate_at": "2025-04-09T10:00:00Z"
  },
  "indoor_pitches": [
    {"id": 1, "name": "P1.25", "module_price": 180.0, "screen_type": "indoor", "is_active": true}
  ],
  "outdoor_pitches": [
    {"id": 5, "name": "P4", "module_price": 45.0, "screen_type": "outdoor", "is_active": true}
  ],
  "usd_rate": 92.5
}
```
**Note:** Курс USD кэшируется в БД и обновляется с сайта ЦБ РФ раз в сутки. Все цены в долларах, итоговый расчёт на клиенте умножается на курс.

---

## Админ API: Проекты

**Auth:** Все эндпоинты требуют JWT (`admin_token` cookie)

**CRUD операции:**
- `POST /admin/projects` - создать (Request: title*, description, location, size, pixel_pitch, featured, categories[])
- `GET /admin/projects/:id` - получить (Response: project + categories, Headers: no-cache)
- `POST /admin/projects/:id/update` - обновить (Request: аналогично создать)
- `DELETE /admin/projects/:id` - удалить (CASCADE: categories, images, views)

**Сортировка:**
- `POST /admin/projects/:id/reorder` - изменить позицию (Request: {position})
- `POST /admin/projects/bulk-reorder` - массовая сортировка (Request: {projects: [{id, sort_order}]})
- `POST /admin/projects/reset-order` - сброс к алфавитному

**Аналитика:**
- `POST /admin/projects/:id/reset-views` - сбросить просмотры (Response: {ok: true})

**Note:** Slug автогенерируется с транслитерацией + уникальный суффикс

---

## Админ API: Изображения

- `POST /admin/upload-images` - загрузить (Request: project_id*, images[], Formats: jpg/png/gif/webp, Max: 10MB)
  - **Автогенерация thumbnails:** Small (400×300px) + Medium (1200×900px) с дефолтным кропом (50%, 50%, 1.0x)
  - **Response:** `{message, images: [{id, filename, thumbnail_small_path, thumbnail_medium_path, is_primary}]}`
- `DELETE /admin/images/:id` - удалить (удаляет оригинал + thumbnails из файловой системы и БД)
- `POST /admin/images/:id/crop` - обновить кроппинг (Request: {crop_x: 0-100, crop_y: 0-100, crop_scale: 0.5-3.0})
  - **Регенерация:** автоматически пересоздает thumbnails с новыми параметрами кропа
- `POST /admin/images/:id/set-primary` - установить как главное изображение проекта (автоматически сбрасывает у других)

**Note:** Имена файлов: `project_{id}_{timestamp}_{index}.ext`, Thumbnails: `*_small.ext`, `*_medium.ext`, Путь: `../frontend/static/uploads/`

---

## Админ API: Прайс-лист

**Auth:** Все эндпоинты требуют JWT (`admin_token` cookie)

**Страницы:**
- `GET /admin/prices` - страница управления (HTML с модальными окнами)
- `GET /admin/prices/:id` - получить позицию (Response: {price_item, images[], specifications[]}, Headers: no-cache)

**CRUD позиций:**
- `POST /admin/prices` - создать (Request: multipart/form-data: title*, description, price_from, has_specifications, is_active, image, specifications JSON)
  - **Response:** `{success: true, price_id: 123, message}`
  - **Note:** После создания открывается модалка редактирования для добавления изображений
- `POST /admin/prices/:id/update` - обновить (Request: аналогично создать)
- `DELETE /admin/prices/:id` - удалить (CASCADE: images, specifications, views)

**Сортировка:**
- `POST /admin/prices/sort` - сохранить порядок drag & drop (Request: {ids: [17, 20, 18]})

**Изображения позиций прайса:**
- `POST /admin/prices/:id/upload-images` - загрузить (Request: multipart/form-data: images[], Formats: jpg/png/gif/webp, Max: 10MB)
  - **Автогенерация WebP thumbnails:** Small (400×300px) + Medium (1200×900px) с дефолтным кропом (50%, 50%, 1.0x)
  - **Response:** `{message, images: [{id, filename, thumbnail_small_path, thumbnail_medium_path}]}`
- `POST /admin/prices/images/:id/crop` - обновить кроппинг (Request: {crop_x: 0-100, crop_y: 0-100, crop_scale: 0.5-3.0})
  - **Регенерация:** автоматически пересоздает WebP thumbnails с новыми параметрами кропа
- `DELETE /admin/prices/images/:id` - удалить (удаляет оригинал + WebP thumbnails из файловой системы и БД)

**Характеристики:**
- Передаются как JSON в поле `specifications` при создании/обновлении
- **Формат:** `[{group: "Параметры экрана", key: "Разрешение", value: "1920x1080", order: 0}, ...]`
- **Группировка:** автоматическая по полю `group` с сортировкой по `order`

**Аналитика:**
- `POST /admin/prices/:id/reset-views` - сбросить просмотры позиции (Response: {ok: true})
- `POST /admin/analytics/reset-prices` - сбросить всю статистику просмотров (TRUNCATE price_view_dailies)

**Note:** Имена файлов: `price_{id}_{timestamp}_{index}.ext`, Thumbnails: `*_small.webp`, `*_medium.webp`, Путь: `../frontend/static/uploads/`

---

## Админ API: Контакты

**Страницы (HTML):**
- `GET /admin/contacts` - список (Query: page, limit, search, status: new/processed, date: today/7d/month, reminder: today/overdue/upcoming)
- `GET /admin/contacts/archive` - архив (Query: аналогично, без status)
- `GET /admin/contacts/export.csv` - экспорт (Format: UTF-8 BOM, delimiter: `;`, date: DD.MM.YYYY HH:MM MSK)

**Статусы:**
- `POST /admin/contacts/:id/status` - изменить (Request: {status: new/processed/archived})
- `POST /admin/contacts/bulk` - массово (Request: {action: new/processed/archived, ids: []})
- `PATCH /admin/contacts/:id/archive` - архивировать (устанавливает archived_at)
- `PATCH /admin/contacts/:id/restore` - восстановить (Request: {to: new/processed}, очищает archived_at)
- `DELETE /admin/contacts/:id` - удалить (Query: ?hard=true для hard delete, иначе soft delete в архив)

---

## Админ API: Заметки

- `GET /admin/contacts/:id/notes` - получить (Response: {notes: [{id, contact_id, text, author, created_at}]}, Sort: created_at DESC)
- `POST /admin/contacts/:id/notes` - создать (Request: {text*, author})
- `DELETE /admin/contacts/:id/notes/:note_id` - удалить (Security: проверяет принадлежность)

**Напоминания:**
- `PATCH /admin/contacts/:id/reminder` - установить (Request: {remind_at: "YYYY-MM-DD HH:MM" MSK или RFC3339, remind_flag}, очистка: remind_at="", UTC storage)

---

## Админ API: Аналитика

- `GET /admin/` - dashboard (HTML: статистика, заявки 7д, напоминания, **топ-5 проектов 30д**, **топ-5 позиций прайса 30д**, график просмотров, system info)
- `POST /admin/analytics/reset` - сбросить всю статистику просмотров проектов (TRUNCATE project_view_dailies)
- `POST /admin/analytics/reset-prices` - сбросить всю статистику просмотров позиций прайса (TRUNCATE price_view_dailies)
- `POST /admin/projects/:id/reset-views` - сбросить просмотры конкретного проекта (DELETE WHERE project_id)
- `POST /admin/prices/:id/reset-views` - сбросить просмотры конкретной позиции прайса (DELETE WHERE price_item_id)

**Фичи топ-5:**
- Клик на название проекта/позиции → переход на страницу управления с автоматическим фокусом и желтой подсветкой (2.5 сек)
- Кнопки сброса статистики для всей таблицы или отдельной записи
- Тултипы с описанием при наведении на название

---

## Админ API: Точки на карте

**Auth:** Все эндпоинты требуют JWT (`admin_token` cookie)

**Страницы:**
- `GET /admin/map-points` - страница управления точками (HTML)
- `GET /admin/map-points/:id` - получить точку (Response: {map_point}, Headers: no-cache)

**CRUD:**
- `POST /admin/map-points` - создать (Request: {title*, latitude*, longitude*, description, panorama_url, is_active})
  - **Дубликаты:** проверка по координатам (±0.0001° ≈ 11м). При совпадении → `409 Conflict`
- `POST /admin/map-points/:id/update` - обновить (Request: аналогично создать)
- `DELETE /admin/map-points/:id` - удалить

**Сортировка:**
- `POST /admin/map-points/sort` - сохранить порядок drag & drop (Request: {ids: [1, 3, 2]})

**Импорт:**
- `POST /admin/map-points/bulk-import` - массовый импорт из ссылок Яндекс.Карт (Request: {links: ["https://yandex.ru/maps/..."]})
  - Автоматически извлекает координаты из параметра `ll=longitude,latitude`
  - Извлекает название из slug `/house/address_slug/` в URL
  - **Дубликаты:** точки с существующими координатами (±11м) пропускаются с ошибкой в отчёте
  - **Response:** `{success: true, created: 3, errors: [...], message}`

**Яндекс.Карты на публичной странице:**
- Страница `/contact` отображает Яндекс.Карту с метками всех активных точек
- Кластеризация меток (Clusterer) при большом количестве точек
- Автоматическое масштабирование карты под все точки
- Балуны с названием, описанием и кнопкой «Панорама» (если указан panorama_url)

**Note:** API ключ Яндекс.Карт подключается в `public_base.html` и `admin_base.html` (страница карты)

---

## Коды ошибок

**HTTP Status:** `200` (OK), `302` (redirect), `400` (bad request/validation), `401` (unauthorized), `404` (not found), `409` (conflict/duplicate), `500` (server error)

**Format:** `{error: "Описание ошибки"}`

---

## Примеры использования

**JavaScript (Fetch API):**
```javascript
// Публичный: GET проекты
await fetch('/api/projects?page=1&limit=12&category=shopping-centers').then(r => r.json())

// Публичный: POST заявка
await fetch('/api/contact', {method: 'POST', headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({name: 'Иван', phone: '+79211234567', email: 'ivan@example.com'})})

// Публичный: POST трекинг просмотра проекта
await fetch('/api/track/project-view/5', {method: 'POST'})

// Публичный: POST трекинг просмотра позиции прайса
await fetch('/api/track/price-view/17', {method: 'POST'})

// Админ: POST обновить статус (важно: credentials: 'include' для JWT cookie!)
await fetch('/admin/contacts/10/status', {method: 'POST', credentials: 'include',
  headers: {'Content-Type': 'application/json'}, body: JSON.stringify({status: 'processed'})})

// Админ: POST загрузка изображений проекта
const fd = new FormData(); fd.append('project_id', '5'); fd.append('images', file);
await fetch('/admin/upload-images', {method: 'POST', credentials: 'include', body: fd})

// Админ: POST загрузка изображений позиции прайса
const fd2 = new FormData(); fd2.append('images', file1); fd2.append('images', file2);
await fetch('/admin/prices/17/upload-images', {method: 'POST', credentials: 'include', body: fd2})
```

**cURL:**
```bash
# GET проекты
curl "http://localhost:8080/api/projects?page=1&limit=12"

# POST заявка
curl -X POST http://localhost:8080/api/contact -H "Content-Type: application/json" \
  -d '{"name":"Иван","phone":"+79211234567"}'

# POST админ (с JWT cookie)
curl -X POST http://localhost:8080/admin/contacts/10/status \
  -H "Content-Type: application/json" -H "Cookie: admin_token=JWT_TOKEN" \
  -d '{"status":"processed"}'
```

---

## Rate Limiting

**Не реализовано.** Production рекомендации: Публичные API - 100 req/min per IP, Админ - 300 req/min per token, Форма - 5 req/hour per IP

---

**v1.1** (Февраль 2026)

