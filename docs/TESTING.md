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

**Покрытие: 119 unit тестов (handlers 31.5%, middleware 100%)**
- ✅ **Middleware (JWT)** - 100% (6 тестов)
- ✅ **Handlers (API)** - основные endpoints (9 тестов)
- ✅ **Admin CRM Actions** - управление заявками, заметки, напоминания (30 тестов)
- ✅ **Admin Projects CRUD** - создание, редактирование, удаление проектов (14 тестов)
- ✅ **Admin Map Points** - CRUD точек, bulk import, парсинг URL Яндекс.Карт (22 теста)
- ✅ **Admin Helpers** - mustID, parseStatus, пагинация, JSON-ответы (22 теста)
- ✅ **Telegram API** - интеграция с Telegram ботом (12 тестов)
- ✅ **SEO** - sitemap.xml, robots.txt, HTTPS (7 тестов)

**Что тестируется:**
- **Public API:** GetProjects, SubmitContact, TrackProjectView (пагинация, валидация)
- **Admin CRM:** UpdateContactStatus, BulkUpdateContacts, ArchiveContact, RestoreContact, DeleteContact, заметки, напоминания (security tests)
- **Admin Projects:** CreateProject (slug generation), GetProject, UpdateProject (many-to-many categories), DeleteProject (cascade, transactions)
- **Telegram Integration:** update status, add note, set reminder, due reminders, mark sent
- **SEO:** HTTPS для production, X-Forwarded-Proto, корректность форматов
- **Admin Map Points:** CRUD (create, get, update, delete), сортировка, bulk import из Яндекс.Карт, парсинг координат, извлечение адреса из URL
- **Admin Helpers:** mustID (валидация/невалидные ID), parseStatus, buildPageNumbers (пагинация), jsonOK/jsonErr, pageMeta, getPageQuery, NowMSK
- **Auth:** валидные/невалидные/истекшие токены, редиректы

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
- ✅ 119 unit тестов (Middleware 100%, Handlers 31.5%, Map Points + Helpers полностью покрыты)
- ✅ 14 smoke tests
- ✅ CI/CD pipeline (GitHub Actions + Codecov)
- ✅ SEO HTTPS оптимизировано для Google/Yandex

**Планы улучшений:**
- 🎯 Handlers покрытие → 50%+ (достигнуто: admin_actions 73-87%, admin_projects 50-88%)
- 🎯 Integration тесты (database CRUD) - частично покрыто в admin tests
- 🎯 E2E тесты (Playwright/Cypress для админ-панели)
- 🎯 Performance тесты (k6, Go benchmarks)

**Дополнительная документация:**
- [LOCAL_CHECKS.md](LOCAL_CHECKS.md) - Локальная проверка кода
- [DEPLOYMENT.md](DEPLOYMENT.md) - Деплой на production

---

**Тестирование - инвестиция в качество!** ✅
