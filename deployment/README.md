# 🚀 Deployment инструкции

## Файлы

- **led-website.service** - правильный systemd service файл
- **deploy.sh** - автоматический скрипт деплоя
- **backup.sh** - скрипт автоматического бэкапа БД

---

## 📋 Первоначальная настройка сервера

### 1. Установка правильного systemd сервиса

```bash
# На локальной машине
scp deployment/led-website.service root@wzvufjpjcz:/opt/led-website/deployment/

# На сервере
ssh root@wzvufjpjcz

# Скопируйте сервис файл
sudo cp /opt/led-website/deployment/led-website.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/led-website.service

# Соберите бинарник
cd /opt/led-website/backend
go build -o led-website -ldflags="-s -w" main.go
chmod +x led-website

# Перезагрузите systemd
sudo systemctl daemon-reload

# Включите автозапуск
sudo systemctl enable led-website

# Запустите сервис
sudo systemctl start led-website

# Проверьте статус
sudo systemctl status led-website
```

### 2. Настройка автоматических бэкапов

```bash
# Создайте скрипт бэкапа
sudo nano /opt/led-website/backup.sh
```

Содержимое файла:

```bash
#!/bin/bash

BACKUP_DIR="/opt/led-website/backups"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="led_display_db_${DATE}.sql.gz"

# Создаем директорию
mkdir -p "$BACKUP_DIR"

# Создаем бэкап
sudo -u postgres pg_dump led_display_db | gzip > "$BACKUP_DIR/$FILENAME"

# Удаляем старые бэкапы (старше 30 дней)
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete

echo "$(date): Backup created: $FILENAME"
```

Настройте права и cron:

```bash
# Сделайте исполняемым
sudo chmod +x /opt/led-website/backup.sh

# Добавьте в cron (ежедневно в 2:00)
sudo crontab -e

# Добавьте строку:
0 2 * * * /opt/led-website/backup.sh >> /var/log/led-backup.log 2>&1
```

---

## 🔄 Деплой обновлений

### Вариант А: Автоматический деплой (рекомендуется)

```bash
# На локальной машине
cd /path/to/led-screen-website

# Сделайте скрипт исполняемым (только первый раз)
chmod +x deployment/deploy.sh

# Запустите деплой
./deployment/deploy.sh
```

Скрипт автоматически:
- Создаст бэкап БД
- Остановит сервис
- Обновит код (git pull)
- Соберет бинарник
- Обновит systemd сервис
- Запустит сервис
- Проверит healthcheck

### Вариант Б: Ручной деплой

```bash
# Подключитесь к серверу
ssh root@wzvufjpjcz

# 1. Бэкап БД
sudo -u postgres pg_dump led_display_db | gzip > /opt/led-website/backups/backup_$(date +%Y%m%d_%H%M%S).sql.gz

# 2. Остановите сервис
sudo systemctl stop led-website

# 3. Обновите код
cd /opt/led-website
git pull origin main

# 4. Соберите бинарник
cd backend
go build -o led-website -ldflags="-s -w" main.go

# 5. Запустите сервис
sudo systemctl start led-website

# 6. Проверьте статус
sudo systemctl status led-website

# 7. Проверьте логи
sudo journalctl -u led-website -n 50 --no-pager

# 8. Проверьте healthcheck
curl http://localhost:8080/healthz
```

---

## 🔍 Проверки после деплоя

```bash
# Статус сервиса
sudo systemctl status led-website

# Логи в реальном времени
sudo journalctl -u led-website -f

# Последние 50 строк логов
sudo journalctl -u led-website -n 50 --no-pager

# Healthcheck
curl http://localhost:8080/healthz

# Проверка что сайт отвечает
curl -I https://s-n-r.ru

# Проверка базы данных
sudo -u postgres psql -d led_display_db -c "SELECT COUNT(*) FROM projects;"
```

---

## 🆘 Откат к предыдущей версии

```bash
# Подключитесь к серверу
ssh root@wzvufjpjcz

cd /opt/led-website

# Посмотрите последние коммиты
git log --oneline -10

# Откатитесь к предыдущей версии
git checkout <commit-hash>

# Пересоберите бинарник
cd backend
go build -o led-website -ldflags="-s -w" main.go

# Перезапустите сервис
sudo systemctl restart led-website

# Проверьте статус
sudo systemctl status led-website
```

---

## 📊 Мониторинг

### Проверка ресурсов

```bash
# CPU и RAM
htop

# Место на диске
df -h

# Размер базы данных
sudo -u postgres psql -c "SELECT pg_size_pretty(pg_database_size('led_display_db'));"

# Количество записей
sudo -u postgres psql -d led_display_db -c "
SELECT
  (SELECT COUNT(*) FROM projects) as projects,
  (SELECT COUNT(*) FROM categories) as categories,
  (SELECT COUNT(*) FROM contact_forms) as contacts,
  (SELECT COUNT(*) FROM images) as images;
"
```

### Логи

```bash
# Приложение
sudo journalctl -u led-website -f

# Nginx
sudo tail -f /var/log/nginx/led-website-access.log
sudo tail -f /var/log/nginx/led-website-error.log

# PostgreSQL
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

---

## ⚙️ Важные команды

```bash
# Управление сервисом
sudo systemctl start led-website
sudo systemctl stop led-website
sudo systemctl restart led-website
sudo systemctl status led-website

# Перезагрузка конфигурации systemd
sudo systemctl daemon-reload

# Включить/выключить автозапуск
sudo systemctl enable led-website
sudo systemctl disable led-website

# Проверка конфигурации Nginx
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🔐 Безопасность

### Проверка прав доступа

```bash
# .env файл должен быть защищен
ls -l /opt/led-website/backend/.env
# Должно быть: -rw------- (600)

# Если нет - исправьте
chmod 600 /opt/led-website/backend/.env
```

### Обновление секретов

```bash
# Сгенерируйте новый JWT_SECRET
openssl rand -base64 32

# Обновите .env
sudo nano /opt/led-website/backend/.env

# Перезапустите сервис
sudo systemctl restart led-website
```

---

## 📝 Changelog

Перед каждым деплоем создавайте git tag:

```bash
# Локально
git tag -a v1.0.1 -m "Description of changes"
git push origin v1.0.1

# Посмотреть все теги
git tag -l
```

---

**Автор:** @alekarah
**Последнее обновление:** 2025-11-27
