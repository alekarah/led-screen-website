#!/bin/bash

# ========================================================================
# СКРИПТ ДЕПЛОЯ LED SCREEN WEBSITE
# ========================================================================
#
# Использование:
#   ./deployment/deploy.sh
#
# Что делает скрипт:
#   1. Делает бэкап базы данных
#   2. Останавливает сервис
#   3. Делает git pull
#   4. Собирает бинарник
#   5. Обновляет systemd сервис
#   6. Запускает сервис
#   7. Проверяет healthcheck
#
# ========================================================================

set -e  # Остановка при первой ошибке

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Настройки
SERVER_HOST="root@wzvufjpjcz"
APP_DIR="/opt/led-website"
BACKUP_DIR="/opt/led-website/backups"
SERVICE_NAME="led-website"

# Функции для вывода
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка SSH подключения
log_info "Проверка подключения к серверу..."
if ! ssh -q "$SERVER_HOST" exit; then
    log_error "Не удалось подключиться к серверу $SERVER_HOST"
    exit 1
fi
log_info "Подключение установлено ✓"

# 1. Создание бэкапа базы данных
log_info "Создание бэкапа базы данных..."
ssh "$SERVER_HOST" "bash -s" << 'EOF'
    set -e
    BACKUP_DIR="/opt/led-website/backups"
    DATE=$(date +%Y%m%d_%H%M%S)
    FILENAME="led_display_db_${DATE}.sql.gz"

    # Создаем директорию если не существует
    mkdir -p "$BACKUP_DIR"

    # Создаем бэкап
    sudo -u postgres pg_dump led_display_db 2>/dev/null | gzip > "$BACKUP_DIR/$FILENAME"

    if [ -f "$BACKUP_DIR/$FILENAME" ]; then
        echo "Бэкап создан: $FILENAME ($(du -h "$BACKUP_DIR/$FILENAME" | cut -f1))"
    else
        echo "ПРЕДУПРЕЖДЕНИЕ: Не удалось создать бэкап (возможно база не существует)"
    fi

    # Удаляем старые бэкапы (старше 30 дней)
    find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete 2>/dev/null || true
EOF

# 2. Остановка сервиса
log_info "Остановка сервиса..."
ssh "$SERVER_HOST" "sudo systemctl stop $SERVICE_NAME" || true

# 3. Git pull
log_info "Обновление кода (git pull)..."
ssh "$SERVER_HOST" "cd $APP_DIR && git pull origin main"

# 4. Сборка бинарника
log_info "Сборка бинарника..."
ssh "$SERVER_HOST" "bash -s" << 'EOF'
    set -e
    cd /opt/led-website/backend

    # Загружаем зависимости (если нужно)
    go mod download

    # Собираем бинарник с оптимизацией
    go build -o led-website -ldflags="-s -w" main.go

    # Делаем исполняемым
    chmod +x led-website

    echo "Бинарник собран: $(du -h led-website | cut -f1)"
EOF

# 5. Обновление systemd сервиса (если файл изменился)
log_info "Проверка systemd сервиса..."
ssh "$SERVER_HOST" "bash -s" << 'EOF'
    set -e

    SOURCE_FILE="/opt/led-website/deployment/led-website.service"
    DEST_FILE="/etc/systemd/system/led-website.service"

    if [ -f "$SOURCE_FILE" ]; then
        # Копируем новый сервис файл
        sudo cp "$SOURCE_FILE" "$DEST_FILE"
        sudo chmod 644 "$DEST_FILE"

        # Перезагружаем systemd
        sudo systemctl daemon-reload

        # Включаем автозапуск
        sudo systemctl enable led-website

        echo "Systemd сервис обновлен"
    else
        echo "ПРЕДУПРЕЖДЕНИЕ: Файл $SOURCE_FILE не найден, сервис не обновлен"
    fi
EOF

# 6. Запуск сервиса
log_info "Запуск сервиса..."
ssh "$SERVER_HOST" "sudo systemctl start $SERVICE_NAME"

# Ждем запуска
sleep 3

# 7. Проверка статуса
log_info "Проверка статуса сервиса..."
if ssh "$SERVER_HOST" "sudo systemctl is-active --quiet $SERVICE_NAME"; then
    log_info "Сервис запущен ✓"
else
    log_error "Сервис не запустился!"
    ssh "$SERVER_HOST" "sudo journalctl -u $SERVICE_NAME -n 30 --no-pager"
    exit 1
fi

# 8. Healthcheck
log_info "Проверка healthcheck..."
sleep 2

if ssh "$SERVER_HOST" "curl -s http://localhost:8080/healthz | grep -q 'ok'"; then
    log_info "Healthcheck прошел успешно ✓"
else
    log_warn "Healthcheck не прошел, проверьте логи:"
    ssh "$SERVER_HOST" "sudo journalctl -u $SERVICE_NAME -n 20 --no-pager"
fi

# 9. Показать статус
echo ""
log_info "================================"
log_info "ДЕПЛОЙ ЗАВЕРШЕН УСПЕШНО!"
log_info "================================"
echo ""
ssh "$SERVER_HOST" "sudo systemctl status $SERVICE_NAME --no-pager | head -n 15"

# Показать последние логи
echo ""
log_info "Последние логи приложения:"
ssh "$SERVER_HOST" "sudo journalctl -u $SERVICE_NAME -n 10 --no-pager"

echo ""
log_info "Сайт доступен по адресу: https://s-n-r.ru"
echo ""
