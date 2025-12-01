# 🚀 Deployment инструкции

## Файлы

- **led-website.service** - systemd service файл для запуска приложения на сервере

---

## 📋 Первоначальная настройка сервера

### Установка systemd сервиса

```bash
# Подключитесь к серверу
ssh root@YOUR_SERVER_IP

# Скопируйте сервис файл
cd /opt/led-website
sudo cp deployment/led-website.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/led-website.service

# Соберите бинарник
cd /opt/led-website/backend
go build -o led-website -ldflags="-s -w" main.go
chmod +x led-website

# Перезагрузите systemd и запустите сервис
sudo systemctl daemon-reload
sudo systemctl enable led-website
sudo systemctl start led-website

# Проверьте статус
sudo systemctl status led-website
```

---

## 🔄 Деплой обновлений

```bash
# Подключитесь к серверу
ssh root@YOUR_SERVER_IP

# Перейдите в директорию проекта
cd /opt/led-website

# Обновите код
git pull

# Соберите бинарник
cd backend
go build -o led-website

# Перезапустите сервис
systemctl restart led-website

# Проверьте статус
systemctl status led-website
```

### Опционально: Бэкап базы перед деплоем

```bash
# Создайте бэкап (опционально, перед важными обновлениями)
mkdir -p /opt/led-website/backups
sudo -u postgres pg_dump led_display_db | gzip > /opt/led-website/backups/backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

---

## 🔍 Проверки после деплоя

```bash
# Статус сервиса
systemctl status led-website

# Последние логи
journalctl -u led-website -n 50 --no-pager

# Логи в реальном времени
journalctl -u led-website -f

# Healthcheck
curl http://localhost:8080/healthz

# Проверка что сайт отвечает
curl -I https://your-domain.com
```

---

## 🆘 Откат к предыдущей версии

```bash
cd /opt/led-website

# Посмотрите последние коммиты
git log --oneline -10

# Откатитесь к нужному коммиту
git checkout <commit-hash>

# Пересоберите бинарник
cd backend
go build -o led-website

# Перезапустите сервис
systemctl restart led-website
```

---

## ⚙️ Управление сервисом

```bash
# Запуск/остановка/перезапуск
systemctl start led-website
systemctl stop led-website
systemctl restart led-website
systemctl status led-website

# Перезагрузка конфигурации systemd (после изменения .service файла)
systemctl daemon-reload

# Включить/выключить автозапуск
systemctl enable led-website
systemctl disable led-website
```

---

## 📊 Мониторинг

### Логи

```bash
# Логи приложения
journalctl -u led-website -f
journalctl -u led-website -n 100 --no-pager

# Nginx логи
tail -f /var/log/nginx/led-website-access.log
tail -f /var/log/nginx/led-website-error.log

# PostgreSQL логи
tail -f /var/log/postgresql/postgresql-*-main.log
```

### Ресурсы

```bash
# CPU и RAM
htop

# Место на диске
df -h

# Размер базы данных
sudo -u postgres psql -c "SELECT pg_size_pretty(pg_database_size('led_display_db'));"

# Количество записей в таблицах
sudo -u postgres psql -d led_display_db -c "
SELECT
  (SELECT COUNT(*) FROM projects) as projects,
  (SELECT COUNT(*) FROM categories) as categories,
  (SELECT COUNT(*) FROM contact_forms) as contacts,
  (SELECT COUNT(*) FROM images) as images;
"
```

### Nginx

```bash
# Проверка конфигурации
nginx -t

# Перезагрузка конфигурации (без downtime)
systemctl reload nginx

# Перезапуск Nginx
systemctl restart nginx
```

---

## 🔐 Безопасность

### Проверка прав доступа к .env

```bash
# .env файл должен быть защищен
ls -l /opt/led-website/backend/.env
# Должно быть: -rw------- (600)

# Если права неправильные - исправьте
chmod 600 /opt/led-website/backend/.env
```

### Обновление JWT секрета

```bash
# Сгенерируйте новый секрет
openssl rand -base64 32

# Обновите .env файл
nano /opt/led-website/backend/.env

# Перезапустите сервис
systemctl restart led-website
```

---

## 🔧 Troubleshooting

### Сервис не запускается

```bash
# Проверьте детальные логи
journalctl -u led-website -n 100 --no-pager

# Проверьте что бинарник существует и исполняемый
ls -la /opt/led-website/backend/led-website

# Проверьте что PostgreSQL запущен
systemctl status postgresql

# Попробуйте запустить бинарник вручную для отладки
cd /opt/led-website/backend
./led-website
```

### База данных недоступна

```bash
# Проверьте статус PostgreSQL
systemctl status postgresql

# Проверьте подключение
sudo -u postgres psql -d led_display_db -c "SELECT 1;"

# Проверьте настройки в .env
cat /opt/led-website/backend/.env | grep DATABASE_URL
```

### Сайт не открывается

```bash
# Проверьте что приложение слушает на порту 8080
netstat -tulpn | grep 8080

# Проверьте Nginx
systemctl status nginx
nginx -t

# Проверьте healthcheck
curl http://localhost:8080/healthz

# Проверьте логи Nginx
tail -n 50 /var/log/nginx/led-website-error.log
```

---

## 📝 Конфигурация

Перед деплоем настройте следующие параметры:

- **Сервер:** YOUR_SERVER_IP
- **Пользователь:** root (или другой пользователь с sudo правами)
- **Директория:** /opt/led-website
- **Порт приложения:** 8080
- **Домен:** your-domain.com
