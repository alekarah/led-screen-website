# 🧪 Тестирование проекта

Документация по автоматическому и ручному тестированию LED Screen Website.

---

## 🧪 Unit Tests

**Запуск:**
```bash
cd backend
go test ./... -v                                    # Все тесты
go test ./... -v -cover -coverprofile=coverage.out # С покрытием
go test ./internal/handlers -run TestGetProjects -v # Конкретный тест
```

**Покрытие: 268 unit тестов (handlers 51.7%, middleware 100%)**
- ✅ **Middleware (JWT)** - 100% (6 тестов)
- ✅ **Handlers (API)** - основные endpoints (13 тестов)
- ✅ **Admin CRM Actions** - управление заявками, заметки, напоминания (30 тестов)
- ✅ **Admin Projects CRUD** - создание, редактирование, удаление, дублирование проектов (19 тестов)
- ✅ **Admin Prices CRUD** - позиции прайса, спецификации, дублирование, сортировка (27 тестов)
- ✅ **Admin Images** - удаление, SetPrimary (транзакции), crop валидация, вспомогательные функции (16 тестов)
- ✅ **Admin Sorting** - ReorderProject, BulkReorderProjects (транзакции, множественные проекты) (11 тестов)
- ✅ **Admin Map Points** - CRUD точек, bulk import, парсинг URL Яндекс.Карт (22 теста)
- ✅ **Admin Helpers** - mustID, parseStatus, пагинация, JSON-ответы, applyDateFilter (27 тестов)
- ✅ **Admin Auth** - Login/Logout, JWT, bcrypt, cookies (23 теста)
- ✅ **Admin Settings** - formatPhone, AdminSettingsUpdate (12 тестов)
- ✅ **Admin Promo** - GetActivePromo, AdminPromoUpdate, getAllPages (15 тестов)
- ✅ **Calculator** - replaceComma, fetchUSDRateFromURL (мок HTTP), getOrRefreshUSDRate (кэш, надбавка, fallback) (15 тестов)
- ✅ **Telegram API** - интеграция с Telegram ботом (12 тестов)
- ✅ **SEO** - sitemap.xml, robots.txt, HTTPS (7 тестов)
- ✅ **Helper Functions** - isImageFile, generateSlug (4 теста)

**Что тестируется:**
- **Public API:** GetProjects (пагинация, валидация, фильтры), SubmitContact (honeypot, spam check), TrackProjectView, TrackPriceView
- **Admin CRM:** UpdateContactStatus, BulkUpdateContacts, ArchiveContact, RestoreContact, DeleteContact, заметки, напоминания (security tests)
- **Admin Projects:** CreateProject (slug generation), GetProject, UpdateProject (many-to-many categories), DeleteProject (cascade, transactions), **DuplicateProject** (slug uniqueness, categories copy)
- **Admin Prices:** CreatePriceItem (с/без спецификаций, валидация), GetPriceItem (с спецификациями), UpdatePriceItem (обновление спецификаций), DeletePriceItem, DuplicatePriceItem (копирование спецификаций, sort_order), UpdatePriceItemsSorting, convertToWebPath (Windows paths fix)
- **Admin Images:** DeleteImage (not found, invalid ID), SetPrimaryImage (транзакции, единственное главное, no project_id), UpdateImageCrop (валидация JSON), validateCropData (граничные значения), createImageRecord, generateImageFilename
- **Admin Sorting:** ReorderProject (валидация позиции, negative values), BulkReorderProjects (транзакции, единственный/множественные проекты, пустой список, GORM Update behavior)
- **Admin Auth:** Login (success, валидация, неверные credentials, неактивный админ), Logout (clear cookie), JWT (генерация, валидация, истечение, подписи), bcrypt (хеширование), cookies (Secure/HttpOnly flags)
- **Admin Settings:** formatPhone (форматирование номеров, граничные случаи), AdminSettingsUpdate (все поля, невалидные числа, форматирование телефона)
- **Admin Promo:** GetActivePromo (активность, страницы, invalid JSON), AdminPromoUpdate (все поля, is_active, invalid TTL/delay, пустые pages, 404), getAllPages (состав и полнота списка)
- **Telegram Integration:** update status, add note, set reminder, due reminders, mark sent
- **SEO:** HTTPS для production, X-Forwarded-Proto, корректность форматов
- **Admin Map Points:** CRUD (create, get, update, delete), сортировка, bulk import из Яндекс.Карт, парсинг координат, извлечение адреса из URL
- **Admin Helpers:** mustID (валидация/невалидные ID), parseStatus, buildPageNumbers (пагинация), jsonOK/jsonErr, pageMeta, getPageQuery, NowMSK, **applyDateFilter** (today, 7d, month)
- **Calculator:** replaceComma (запятая/точка/пустая строка), fetchUSDRateFromURL (успех, запятая в числе, USD не найден, ошибка сервера, невалидный XML), getOrRefreshUSDRate (свежий кэш, расчёт надбавки, fallback при недоступном ЦБ, пустая БД)
- **Helper Functions:** isImageFile (valid/invalid extensions), generateSlug (транслитерация)

---

## 🚀 CI/CD Pipeline

GitHub Actions автоматически при каждом push в `main`/`develop`:
1. **Test** - запуск тестов + coverage → Codecov
2. **Lint** - golangci-lint (стиль, безопасность)
3. **Build** - компиляция бинарника

**Результаты:** GitHub → вкладка "Checks" или badges в README

---

## 🚀 Smoke Tests

**Быстрая проверка критичных функций (~30 сек):**

```bash
# Windows
.\test-smoke.ps1

# Linux/Mac/Git Bash
./test-smoke.sh
```

**14 автоматических проверок:**
1. Зависимости (Go, PostgreSQL)
2. Сборка проекта
3. Запуск тестового сервера
4. HTTP endpoints (/, /projects, /services, /contact, /api/projects)
5. Админ панель (login доступен, dashboard защищен)
6. Очистка

**Когда запускать:**
- ✅ Перед каждым коммитом
- ✅ Перед деплоем на production
- ✅ После изменения handlers/routes
- ✅ После обновления зависимостей

---

## 📝 Ручное тестирование

**Публичная часть (~15 мин):**
- Главная: навигация, избранные проекты, услуги
- Портфолио: фильтр по категориям, пагинация, изображения
- Контакты: форма, валидация, отправка
- Адаптивность: desktop/tablet/mobile (1920px/768px/375px)

**Админ-панель (~30 мин):**
- Авторизация: вход/выход, "запомнить меня"
- Dashboard: статистика, график просмотров, напоминания
- Проекты: CRUD, загрузка изображений, crop editor, drag&drop сортировка
- Заявки: статусы, заметки, напоминания, фильтры, экспорт CSV

---

## 🐛 Troubleshooting

**PostgreSQL не запущен:**
- Windows: Services → PostgreSQL → Start
- Linux: `sudo systemctl start postgresql`
- Mac: `brew services start postgresql`

**Проект не собирается:**
```bash
cd backend
go mod tidy
go build main.go  # Смотрите вывод ошибки
```

**Сервер не запускается:**
- Проверьте `.env` существует: `ls backend/.env`
- Проверьте DATABASE_URL: `cat backend/.env`
- Смотрите логи: `cat backend/test-error.log`

**Порт 8080 занят:**
- Windows: `Get-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess`
- Linux/Mac: `lsof -i :8080`

**Админ панель не защищена:**
- Проверьте JWT_SECRET в `.env`
- Проверьте middleware в `routes/routes.go`

---

## 📈 Статистика и планы

**Текущее состояние:**
- ✅ **268 unit тестов** (Middleware 100%, Handlers 51.7%)
- ✅ 14 smoke tests
- ✅ CI/CD pipeline (GitHub Actions + Codecov)
- ✅ SEO HTTPS оптимизировано для Google/Yandex
- ✅ **Handlers покрытие 51.7%** (admin_actions 73-87%, admin_projects 50-88%, admin_auth 91-100%, admin_helpers 100%, admin_promo ~80%, admin_settings ~85%)

**Планы улучшений:**
- 🎯 E2E тесты (Playwright/Cypress для админ-панели)
- 🎯 Performance тесты (k6, Go benchmarks)
- ⏸️ Handlers покрытие 60%+ — остаток (HTML-рендеринг, файловые операции, внешние API) нецелесообразно покрывать unit-тестами

**Дополнительная документация:**
- [LOCAL_CHECKS.md](LOCAL_CHECKS.md) - Локальная проверка кода
- [DEPLOYMENT.md](DEPLOYMENT.md) - Деплой на production

---

**Тестирование - инвестиция в качество!** ✅
