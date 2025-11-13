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
**Note:** Агрегирует просмотры по дням (UTC) в `project_view_dailies`

### 4. Статистика заявок за 7 дней

`GET /api/admin/contacts-7d`

**Response** (200): `[{day: "2024-11-01", count: 3}, ...]`

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
- `DELETE /admin/images/:id` - удалить (удаляет из БД и файловой системы)
- `POST /admin/images/:id/crop` - обновить кроппинг (Request: {crop_x: 0-100, crop_y: 0-100, crop_scale: 0.5-3.0})

**Note:** Имена файлов: `project_{id}_{timestamp}_{index}.ext`, путь: `../frontend/static/uploads/`

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

- `GET /admin/` - dashboard (HTML: статистика, заявки 7д, напоминания, топ-5 проектов 30д, график просмотров, system info)
- `POST /admin/analytics/reset` - сбросить всю статистику просмотров (TRUNCATE project_view_dailies)

---

## Коды ошибок

**HTTP Status:** `200` (OK), `302` (redirect), `400` (bad request/validation), `401` (unauthorized), `404` (not found), `500` (server error)

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

// Админ: POST обновить статус (важно: credentials: 'include' для JWT cookie!)
await fetch('/admin/contacts/10/status', {method: 'POST', credentials: 'include',
  headers: {'Content-Type': 'application/json'}, body: JSON.stringify({status: 'processed'})})

// Админ: POST загрузка изображений
const fd = new FormData(); fd.append('project_id', '5'); fd.append('images', file);
await fetch('/admin/upload-images', {method: 'POST', credentials: 'include', body: fd})
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

**v1.0** (Ноябрь 2024)

