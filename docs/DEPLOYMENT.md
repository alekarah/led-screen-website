# 🚀 Deployment Guide

> Инструкция по развертыванию LED Screen Website на production сервере

## 📋 Содержание

- [Требования](#требования)
- [Локальное развертывание](#локальное-развертывание)
- [Production развертывание](#production-развертывание)
- [Docker развертывание](#docker-развертывание)
- [Nginx настройка](#nginx-настройка)
- [SSL сертификаты](#ssl-сертификаты)
- [Системная служба](#системная-служба)
- [Бэкапы базы данных](#бэкапы-базы-данных)
- [Мониторинг](#мониторинг)
- [Обновление приложения](#обновление-приложения)
- [Troubleshooting](#troubleshooting)

---

## Требования

### Минимальные требования сервера

**Hardware**:
- CPU: 2 cores (рекомендуется 4)
- RAM: 2 GB (рекомендуется 4 GB)
- Disk: 20 GB SSD (рекомендуется 50 GB)
- Network: 100 Mbps

**Software**:
- OS: Ubuntu 20.04+ / Debian 11+ / CentOS 8+
- Go: 1.21 или выше
- PostgreSQL: 15
- Nginx: 1.18+ (для reverse proxy)
- Git: 2.25+
- Docker & Docker Compose: latest (опционально)

### Проверка требований

```bash
# Проверка версий
go version          # go1.21 или выше
psql --version      # PostgreSQL 15
nginx -v            # nginx/1.18 или выше
docker --version
docker compose version
```

---

## Локальное развертывание

### 1. Клонирование репозитория

```bash
git clone https://github.com/yourusername/led-screen-website.git
cd led-screen-website
```

### 2. Настройка переменных окружения

```bash
cd backend
# Создать .env файл
cp .env.example .env

# Редактировать .env
nano .env
```

**Минимальная конфигурация**:
```env
ENVIRONMENT=development
PORT=8080
DATABASE_URL=postgres://postgres:password123@localhost:5432/led_display_db?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production
```

**Генерация JWT секрета**:
```bash
openssl rand -base64 32
```

### 3. Запуск PostgreSQL через Docker

```bash
docker compose up -d postgres
```

Или установка напрямую:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql-15 postgresql-contrib

# CentOS/RHEL
sudo dnf install postgresql15-server
sudo postgresql-setup --initdb
sudo systemctl start postgresql
```

**Примечание:** Инициализация базы данных выполняется автоматически через GORM миграции при первом запуске приложения (см. `backend/main.go`).

### 4. Установка зависимостей Go

```bash
cd backend
go mod download
go mod verify
```

### 5. Создание первого администратора

```bash
cd cmd/create-admin
go run main.go
# Введите username и password
```

Или через интерактивный режим:
```bash
go run cmd/create-admin/main.go
```

### 6. Запуск приложения

```bash
cd backend
go run main.go
```

**Проверка**:
- Публичная часть: http://localhost:8080
- Админ панель: http://localhost:8080/admin
- Healthcheck: http://localhost:8080/healthz

---

## Проверка перед деплоем

**⚠️ ВАЖНО:** Перед деплоем на production обязательно запустите smoke tests!

### Автоматическое тестирование (Smoke Tests)

Проект включает автоматические smoke tests которые за 30 секунд проверят все критичные функции.

**Windows (PowerShell):**
```powershell
# В корне проекта
.\test-smoke.ps1
```

**Linux / Mac / Git Bash:**
```bash
chmod +x test-smoke.sh
./test-smoke.sh
```

**Что проверяют тесты:**
- ✅ Зависимости (Go, PostgreSQL)
- ✅ Сборка проекта без ошибок
- ✅ Запуск сервера
- ✅ Все критичные HTTP endpoints
- ✅ Публичные страницы (главная, портфолио, услуги, контакты)
- ✅ API endpoints
- ✅ Защита админ-панели

**Результат должен быть:**
```
✅ ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО!
   Проект готов к деплою
```

Если есть проваленные тесты - исправьте их перед деплоем. Подробная документация: [docs/TESTING.md](TESTING.md)

### Чек-лист безопасности

Перед деплоем также проверьте: [docs/SECURITY_CHECKLIST.md](SECURITY_CHECKLIST.md)

**Критически важно:**
- [ ] JWT_SECRET изменен с дефолтного
- [ ] Пароль PostgreSQL сложный
- [ ] .env не в Git репозитории
- [ ] ENVIRONMENT=production
- [ ] Репозиторий приватный

---

## Production развертывание

### Подготовка сервера

#### 1. Обновление системы

```bash
sudo apt update && sudo apt upgrade -y  # Ubuntu/Debian
sudo dnf update -y                       # CentOS/RHEL
```

#### 2. Установка Go

```bash
# Загрузка Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# Установка
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Добавить в PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Проверка
go version
```

#### 3. Установка PostgreSQL 15

```bash
# Ubuntu/Debian
sudo apt install -y postgresql-15 postgresql-contrib

# CentOS/RHEL
sudo dnf install -y postgresql15-server postgresql15-contrib
sudo postgresql-setup --initdb
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

#### 4. Настройка PostgreSQL

```bash
# Войти в psql
sudo -u postgres psql

# Создать базу и пользователя
CREATE DATABASE led_display_db;
CREATE USER led_user WITH ENCRYPTED PASSWORD 'strong_password_here';
GRANT ALL PRIVILEGES ON DATABASE led_display_db TO led_user;
\q
```

**Настройка pg_hba.conf** (разрешить локальные подключения):
```bash
sudo nano /etc/postgresql/15/main/pg_hba.conf
```

Добавить:
```
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   led_display_db  led_user                                md5
host    led_display_db  led_user        127.0.0.1/32            md5
```

Перезапустить PostgreSQL:
```bash
sudo systemctl restart postgresql
```

#### 5. Развертывание приложения

```bash
# Создать директорию для приложения
sudo mkdir -p /opt/led-website
sudo chown $USER:$USER /opt/led-website

# Клонировать репозиторий
cd /opt/led-website
git clone https://github.com/yourusername/led-screen-website.git .

# Настроить .env
cd backend
cp .env.example .env
nano .env
```

**Production .env**:
```env
ENVIRONMENT=production
PORT=8080
APP_VERSION=1.0.0

# Database
DATABASE_URL=postgres://led_user:strong_password_here@localhost:5432/led_display_db?sslmode=disable
DB_LOG_LEVEL=error
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME_MIN=30

# Security
JWT_SECRET=<generated_secret_from_openssl>

# Uploads
UPLOAD_PATH=../frontend/static/uploads
MAX_UPLOAD_SIZE=10485760
```

#### 6. Сборка приложения

```bash
cd /opt/led-website/backend
go mod download

# Сборка бинарника
go build -o led-website -ldflags="-s -w" main.go

# Проверка
./led-website
```

**Примечание:** База данных будет инициализирована автоматически через GORM миграции при первом запуске приложения.

#### 7. Создание администратора

```bash
cd /opt/led-website/backend/cmd/create-admin
go run main.go
```

---

## Docker развертывание

### Создание Dockerfile

Создать `backend/Dockerfile`:

```dockerfile
# Builder stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходников
COPY . .

# Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копирование бинарника
COPY --from=builder /app/main .

# Создание директорий
RUN mkdir -p ../frontend/static/uploads

# Порт
EXPOSE 8080

# Запуск
CMD ["./main"]
```

### Обновление docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:15
    container_name: led-postgres
    environment:
      POSTGRES_DB: led_display_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    networks:
      - led-network

  app:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: led-app
    environment:
      - ENVIRONMENT=production
      - PORT=8080
      - DATABASE_URL=postgres://postgres:password123@postgres:5432/led_display_db?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8080:8080"
    volumes:
      - ./frontend:/root/frontend:ro
      - uploads:/root/frontend/static/uploads
    depends_on:
      - postgres
    restart: unless-stopped
    networks:
      - led-network

  nginx:
    image: nginx:alpine
    container_name: led-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - ./frontend:/usr/share/nginx/html:ro
      - uploads:/usr/share/nginx/html/static/uploads:ro
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - led-network

volumes:
  postgres_data:
  uploads:

networks:
  led-network:
    driver: bridge
```

### Запуск через Docker

```bash
# Сборка и запуск всех сервисов
docker compose up -d --build

# Проверка логов
docker compose logs -f app

# Проверка статуса
docker compose ps

# Остановка
docker compose down

# Полная очистка (с удалением volumes)
docker compose down -v
```

---

## Nginx настройка

### Установка Nginx

```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo dnf install nginx

# Запуск
sudo systemctl enable nginx
sudo systemctl start nginx
```

### Конфигурация Nginx (Reverse Proxy)

Создать `/etc/nginx/sites-available/led-website`:

```nginx
upstream led_backend {
    server localhost:8080;
    keepalive 64;
}

# HTTP -> HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name yourdomain.com www.yourdomain.com;

    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # SSL configuration (Mozilla Modern)
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript
               application/json application/javascript application/xml+rss;

    # Static files
    location /static/ {
        alias /opt/led-website/frontend/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Uploads (user-generated content)
    location /static/uploads/ {
        alias /opt/led-website/frontend/static/uploads/;
        expires 30d;
        add_header Cache-Control "public";
    }

    # Proxy to Go backend
    location / {
        proxy_pass http://led_backend;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_buffering off;
        proxy_request_buffering off;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }

    # Rate limiting for contact form
    location /api/contact {
        limit_req zone=contact_limit burst=5 nodelay;
        proxy_pass http://led_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Access logs
    access_log /var/log/nginx/led-website-access.log;
    error_log /var/log/nginx/led-website-error.log warn;
}

# Rate limiting zone
limit_req_zone $binary_remote_addr zone=contact_limit:10m rate=5r/h;
```

**Активация конфигурации**:
```bash
# Создать симлинк
sudo ln -s /etc/nginx/sites-available/led-website /etc/nginx/sites-enabled/

# Проверка конфигурации
sudo nginx -t

# Перезагрузка
sudo systemctl reload nginx
```

---

## SSL сертификаты

### Let's Encrypt (Certbot)

#### 1. Установка Certbot

```bash
# Ubuntu/Debian
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo dnf install certbot python3-certbot-nginx
```

#### 2. Получение сертификата

```bash
# Автоматическая настройка Nginx
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Или только получить сертификат
sudo certbot certonly --webroot -w /var/www/html \
    -d yourdomain.com -d www.yourdomain.com
```

#### 3. Автообновление сертификатов

```bash
# Проверка автообновления
sudo certbot renew --dry-run

# Cron job (уже создается автоматически)
sudo crontab -e
```

Добавить:
```
0 0,12 * * * /usr/bin/certbot renew --quiet --deploy-hook "systemctl reload nginx"
```

---

## Системная служба

### Создание systemd service

Создать `/etc/systemd/system/led-website.service`:

```ini
[Unit]
Description=LED Screen Website
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/led-website/backend
ExecStart=/opt/led-website/backend/led-website
Restart=on-failure
RestartSec=5s

# Environment
Environment="ENVIRONMENT=production"
EnvironmentFile=/opt/led-website/backend/.env

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/led-website/frontend/static/uploads

# Limits
LimitNOFILE=65536
LimitNPROC=4096

# Logs
StandardOutput=journal
StandardError=journal
SyslogIdentifier=led-website

[Install]
WantedBy=multi-user.target
```

**Управление службой**:
```bash
# Перезагрузка конфигурации
sudo systemctl daemon-reload

# Запуск
sudo systemctl start led-website

# Автозапуск
sudo systemctl enable led-website

# Статус
sudo systemctl status led-website

# Логи
sudo journalctl -u led-website -f

# Перезапуск
sudo systemctl restart led-website

# Остановка
sudo systemctl stop led-website
```

---

## Бэкапы базы данных

### Автоматический бэкап (ежедневный)

Создать `/opt/led-website/scripts/backup.sh`:

```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/opt/led-website/backups"
DB_NAME="led_display_db"
DB_USER="led_user"
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup filename with timestamp
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_$(date +%Y%m%d_%H%M%S).sql.gz"

# Create backup
PGPASSWORD="your_password" pg_dump -h localhost -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"

# Check if backup was successful
if [ $? -eq 0 ]; then
    echo "Backup successful: $BACKUP_FILE"

    # Remove old backups
    find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +$RETENTION_DAYS -delete
    echo "Old backups removed (older than $RETENTION_DAYS days)"
else
    echo "Backup failed!" >&2
    exit 1
fi
```

**Сделать исполняемым**:
```bash
chmod +x /opt/led-website/scripts/backup.sh
```

**Cron job** (ежедневно в 2:00):
```bash
sudo crontab -e
```

Добавить:
```
0 2 * * * /opt/led-website/scripts/backup.sh >> /var/log/led-website-backup.log 2>&1
```

### Восстановление из бэкапа

```bash
# Восстановление из gzip backup
gunzip < /opt/led-website/backups/led_display_db_20241102_020000.sql.gz | \
    psql -h localhost -U led_user -d led_display_db

# Или из обычного .sql файла
psql -h localhost -U led_user -d led_display_db < backup.sql
```

---

## Мониторинг

### 1. Healthcheck endpoint

```bash
# Проверка доступности
curl http://localhost:8080/healthz
# Должен вернуть: ok
```

### 2. Мониторинг логов

```bash
# Системные логи приложения
sudo journalctl -u led-website -f

# Nginx access logs
sudo tail -f /var/log/nginx/led-website-access.log

# Nginx error logs
sudo tail -f /var/log/nginx/led-website-error.log

# PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

### 3. Мониторинг ресурсов

```bash
# CPU и память
htop

# Дисковое пространство
df -h

# Использование БД
sudo -u postgres psql -d led_display_db -c "
SELECT
    pg_size_pretty(pg_database_size('led_display_db')) AS db_size,
    (SELECT count(*) FROM projects) AS projects_count,
    (SELECT count(*) FROM contact_forms) AS contacts_count;
"
```

### 4. Скрипт мониторинга

Создать `/opt/led-website/scripts/monitor.sh`:

```bash
#!/bin/bash

# Check if app is running
if ! systemctl is-active --quiet led-website; then
    echo "ERROR: Application is not running!"
    systemctl restart led-website
fi

# Check healthcheck
HEALTH=$(curl -s http://localhost:8080/healthz)
if [ "$HEALTH" != "ok" ]; then
    echo "ERROR: Healthcheck failed!"
fi

# Check disk space (alert if > 80%)
DISK_USAGE=$(df -h /opt/led-website | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "WARNING: Disk usage is at ${DISK_USAGE}%"
fi

echo "Monitoring check completed at $(date)"
```

**Cron job** (каждые 5 минут):
```
*/5 * * * * /opt/led-website/scripts/monitor.sh >> /var/log/led-website-monitor.log 2>&1
```

---

## Обновление приложения

### Через Git (рекомендуется)

```bash
# Перейти в директорию приложения
cd /opt/led-website

# Создать резервную копию БД
/opt/led-website/scripts/backup.sh

# Остановить службу
sudo systemctl stop led-website

# Обновить код
git pull origin main

# Обновить зависимости
cd backend
go mod download

# Пересобрать приложение
go build -o led-website -ldflags="-s -w" main.go

# Применить миграции (если есть)
# psql -h localhost -U led_user -d led_display_db -f migrations/new_migration.sql

# Перезапустить службу
sudo systemctl start led-website

# Проверить статус
sudo systemctl status led-website
sudo journalctl -u led-website -f
```

### Через Docker

```bash
cd /opt/led-website

# Бэкап БД
docker compose exec postgres pg_dump -U postgres led_display_db > backup.sql

# Остановить контейнеры
docker compose down

# Обновить код
git pull origin main

# Пересобрать и запустить
docker compose up -d --build

# Проверить логи
docker compose logs -f app
```

### Откат к предыдущей версии

```bash
# Git откат
git log --oneline  # Найти хеш нужного коммита
git checkout <commit-hash>

# Пересобрать
cd backend && go build -o led-website main.go

# Перезапустить
sudo systemctl restart led-website
```

---

## Troubleshooting

### Проблема: Приложение не запускается

**Проверка логов**:
```bash
sudo journalctl -u led-website -n 50 --no-pager
```

**Возможные причины**:
1. БД не доступна:
   ```bash
   sudo systemctl status postgresql
   psql -h localhost -U led_user -d led_display_db -c "SELECT 1;"
   ```

2. Неверные права доступа:
   ```bash
   ls -la /opt/led-website/backend/led-website
   sudo chown www-data:www-data /opt/led-website/backend/led-website
   ```

3. Порт уже занят:
   ```bash
   sudo lsof -i :8080
   sudo netstat -tulpn | grep 8080
   ```

### Проблема: 502 Bad Gateway (Nginx)

**Проверка**:
```bash
# Проверка, что приложение запущено
curl http://localhost:8080/healthz

# Проверка Nginx конфигурации
sudo nginx -t

# Проверка логов Nginx
sudo tail -f /var/log/nginx/led-website-error.log
```

### Проблема: База данных переполнена

**Очистка**:
```bash
# Удалить старые архивные заявки (старше 1 года)
psql -h localhost -U led_user -d led_display_db -c "
DELETE FROM contact_forms
WHERE archived_at IS NOT NULL
  AND archived_at < NOW() - INTERVAL '1 year';
"

# Очистить статистику просмотров (старше 3 месяцев)
psql -h localhost -U led_user -d led_display_db -c "
DELETE FROM project_view_dailies
WHERE day < CURRENT_DATE - INTERVAL '90 days';
"

# VACUUM для освобождения места
psql -h localhost -U led_user -d led_display_db -c "VACUUM FULL;"
```

### Проблема: Медленные запросы

**Анализ**:
```bash
# Включить логирование медленных запросов
sudo nano /etc/postgresql/15/main/postgresql.conf
```

Добавить:
```
log_min_duration_statement = 1000  # Логировать запросы > 1сек
```

Перезапустить PostgreSQL:
```bash
sudo systemctl restart postgresql
```

**Проверка индексов**:
```sql
-- Найти таблицы без индексов
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Неиспользуемые индексы
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;
```

---

## Security Checklist

### Production checklist

- [ ] Сгенерирован сложный `JWT_SECRET`
- [ ] Изменены пароли БД
- [ ] Настроены SSL сертификаты (HTTPS)
- [ ] Настроены security headers в Nginx
- [ ] Включен firewall (ufw/firewalld)
- [ ] Закрыт прямой доступ к PostgreSQL извне
- [ ] Настроены автоматические бэкапы
- [ ] Включено логирование
- [ ] Настроен мониторинг
- [ ] Обновлены системные пакеты
- [ ] Отключен root login по SSH
- [ ] Настроен fail2ban (опционально)

### Firewall настройка

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --reload
```

---

## Performance Tuning

### PostgreSQL оптимизация

```bash
sudo nano /etc/postgresql/15/main/postgresql.conf
```

```ini
# Memory
shared_buffers = 256MB              # 25% RAM
effective_cache_size = 1GB          # 50-75% RAM
work_mem = 4MB
maintenance_work_mem = 64MB

# Connections
max_connections = 100

# WAL
wal_buffers = 16MB
checkpoint_completion_target = 0.9

# Query planner
random_page_cost = 1.1              # Для SSD
effective_io_concurrency = 200      # Для SSD
```

Перезапустить:
```bash
sudo systemctl restart postgresql
```

### Nginx кеширование

Добавить в `nginx.conf`:
```nginx
# Cache zone
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=1g
                 inactive=60m use_temp_path=off;

# В location /
location / {
    proxy_cache my_cache;
    proxy_cache_valid 200 10m;
    proxy_cache_use_stale error timeout http_500 http_502 http_503 http_504;

    add_header X-Cache-Status $upstream_cache_status;

    # ... остальная proxy конфигурация
}
```

---

## Версии и совместимость

| Компонент      | Минимальная версия | Рекомендуемая | Проверено |
|----------------|--------------------|---------------|-----------|
| Go             | 1.21               | 1.21+         | 1.21.0    |
| PostgreSQL     | 15                 | 15            | 15.3      |
| Nginx          | 1.18               | 1.22+         | 1.22.1    |
| Ubuntu         | 20.04              | 22.04         | 22.04 LTS |
| Debian         | 11                 | 12            | 12        |
| Docker         | 20.10              | 24.0+         | 24.0.5    |
| Docker Compose | 2.0                | 2.20+         | 2.21.0    |

---

**Версия документа**: 1.0
**Последнее обновление**: Ноябрь 2024

