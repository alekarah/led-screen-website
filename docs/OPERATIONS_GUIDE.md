# 🚀 Operations Guide

> Руководство по операциям и обслуживанию LED Screen Website (production: https://s-n-r.ru)

**Статус:** Приложение развернуто на Beget VPS (Ubuntu 22.04, 2GB RAM, PostgreSQL 15, Nginx + SSL)

---

## Управление службой

**Systemd service:** `/etc/systemd/system/led-website.service`

```bash
# Статус
sudo systemctl status led-website

# Перезапуск
sudo systemctl restart led-website

# Остановка/запуск
sudo systemctl stop led-website
sudo systemctl start led-website

# Логи
sudo journalctl -u led-website -f
sudo journalctl -u led-website -n 100 --no-pager
```

---

## Бэкапы базы данных

**Ручной бэкап:**
```bash
# Создать бэкап
PGPASSWORD="your_password" pg_dump -h localhost -U led_user led_display_db | gzip > backup_$(date +%Y%m%d).sql.gz
```

**Восстановление:**
```bash
# Из gzip
gunzip < backup_20241102.sql.gz | psql -h localhost -U led_user -d led_display_db

# Из .sql
psql -h localhost -U led_user -d led_display_db < backup.sql
```

**Автоматизация:** Cron job в `/opt/led-website/scripts/backup.sh` (ежедневно 2:00, retention 30 дней)

---

## Мониторинг

**Healthcheck:**
```bash
curl http://localhost:8080/healthz  # Должен вернуть: ok
```

**Логи:**
```bash
# Приложение
sudo journalctl -u led-website -f

# Nginx
sudo tail -f /var/log/nginx/led-website-access.log
sudo tail -f /var/log/nginx/led-website-error.log
```

**Ресурсы:**
```bash
# CPU/RAM
htop

# Диск
df -h

# БД статистика
sudo -u postgres psql -d led_display_db -c "SELECT pg_size_pretty(pg_database_size('led_display_db')) AS db_size, (SELECT count(*) FROM projects) AS projects, (SELECT count(*) FROM contact_forms) AS contacts;"
```

---

## Обновление приложения

```bash
cd /opt/led-website && /opt/led-website/scripts/backup.sh  # Бэкап БД

sudo systemctl stop led-website

git pull origin main
cd backend && go mod download && go build -o led-website -ldflags="-s -w" main.go

# Миграции (если есть): psql -h localhost -U led_user -d led_display_db -f migrations/new.sql

sudo systemctl start led-website && sudo systemctl status led-website
```

**Откат:** `git log --oneline` → `git checkout <hash>` → rebuild → restart

---

## Troubleshooting

### Приложение не запускается
```bash
sudo journalctl -u led-website -n 50 --no-pager  # Проверка логов
sudo systemctl status postgresql                  # БД доступна?
curl http://localhost:8080/healthz               # Healthcheck
sudo lsof -i :8080                               # Порт занят?
```

### 502 Bad Gateway (Nginx)
```bash
curl http://localhost:8080/healthz               # App работает?
sudo nginx -t                                    # Nginx конфигурация
sudo tail -f /var/log/nginx/led-website-error.log
```

### База данных переполнена
```bash
# Удалить старые архивные заявки (>1 года) и просмотры (>90 дней)
psql -h localhost -U led_user -d led_display_db -c "
DELETE FROM contact_forms WHERE archived_at < NOW() - INTERVAL '1 year';
DELETE FROM project_view_dailies WHERE day < CURRENT_DATE - INTERVAL '90 days';
VACUUM FULL;
"
```

---

**Версия документа**: 1.0
**Последнее обновление**: Ноябрь 2024

