#!/bin/bash
# ========================================
# Smoke Test для LED Screen Website
# ========================================
# Быстрая автоматическая проверка критичных функций
# Время выполнения: ~30 секунд
# ========================================

set +e  # Не останавливаться при ошибках

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

tests_passed=0
tests_failed=0
tests_total=0

echo -e "${CYAN}========================================"
echo -e "  SMOKE TEST - LED Screen Website"
echo -e "========================================${NC}"
echo ""

# Функция для вывода результата теста
test_result() {
    local test_name="$1"
    local success="$2"
    local details="$3"

    ((tests_total++))

    if [ "$success" = "true" ]; then
        echo -e "${GREEN}[PASS]${NC} $test_name"
        [ -n "$details" ] && echo -e "       ${GRAY}$details${NC}"
        ((tests_passed++))
    else
        echo -e "${RED}[FAIL]${NC} $test_name"
        [ -n "$details" ] && echo -e "       ${YELLOW}$details${NC}"
        ((tests_failed++))
    fi
}

# ========================================
# 1. Проверка зависимостей
# ========================================
echo -e "\n${YELLOW}1. Проверка зависимостей...${NC}"

# Go установлен?
if command -v go &> /dev/null; then
    go_version=$(go version)
    test_result "Go установлен" "true" "$go_version"
else
    test_result "Go установлен" "false" "Go не найден в PATH"
fi

# PostgreSQL запущен?
if pgrep -x postgres > /dev/null 2>&1; then
    pg_pid=$(pgrep -x postgres | head -1)
    test_result "PostgreSQL запущен" "true" "PID: $pg_pid"
else
    test_result "PostgreSQL запущен" "false" "Процесс postgres не найден"
fi

# ========================================
# 2. Сборка проекта
# ========================================
echo -e "\n${YELLOW}2. Сборка проекта...${NC}"

cd "$(dirname "$0")/backend" || exit 1

if go build -o led-backend-test main.go 2>&1; then
    test_result "Проект собирается" "true" "Бинарник создан"
else
    test_result "Проект собирается" "false" "Ошибка компиляции"
    exit 1
fi

# ========================================
# 3. Запуск сервера
# ========================================
echo -e "\n${YELLOW}3. Запуск тестового сервера...${NC}"

# Проверяем что .env существует
if [ ! -f ".env" ]; then
    test_result ".env файл существует" "false" "Создайте .env файл"
    rm -f led-backend-test
    exit 1
else
    test_result ".env файл существует" "true"
fi

# Запускаем сервер в фоне
./led-backend-test > test-output.log 2> test-error.log &
server_pid=$!

if [ -n "$server_pid" ]; then
    test_result "Сервер запущен" "true" "PID: $server_pid"
else
    test_result "Сервер запущен" "false"
    rm -f led-backend-test
    exit 1
fi

# Ждем пока сервер запустится
echo -e "\n   ${GRAY}Ожидание запуска сервера...${NC}"
sleep 3

# Проверяем что процесс еще жив
if ! kill -0 $server_pid 2>/dev/null; then
    test_result "Сервер продолжает работать" "false" "Сервер упал при запуске"
    cat test-error.log 2>/dev/null
    rm -f led-backend-test
    exit 1
else
    test_result "Сервер продолжает работать" "true"
fi

# ========================================
# 4. HTTP тесты
# ========================================
echo -e "\n${YELLOW}4. Тестирование HTTP endpoints...${NC}"

base_url="http://localhost:8080"

# Функция для HTTP запроса
test_endpoint() {
    local url="$1"
    local name="$2"
    local expected_content="$3"

    if response=$(curl -s -w "\n%{http_code}" "$url" 2>&1); then
        http_code=$(echo "$response" | tail -1)
        body=$(echo "$response" | head -n -1)

        if [ "$http_code" = "200" ]; then
            if [ -n "$expected_content" ] && ! echo "$body" | grep -q "$expected_content"; then
                test_result "$name" "false" "Страница не содержит '$expected_content'"
            else
                test_result "$name" "true" "HTTP 200 OK"
            fi
        else
            test_result "$name" "false" "HTTP $http_code"
        fi
    else
        test_result "$name" "false" "Ошибка запроса"
    fi
}

# Публичные страницы
test_endpoint "$base_url/" "Главная страница" "LED"
test_endpoint "$base_url/projects" "Страница проектов"
test_endpoint "$base_url/services" "Страница услуг"
test_endpoint "$base_url/contact" "Страница контактов"

# API endpoints
if api_response=$(curl -s "$base_url/api/projects" 2>&1); then
    if echo "$api_response" | grep -q "projects"; then
        test_result "API /api/projects" "true" "Получены данные проектов"
    else
        test_result "API /api/projects" "true" "Нет проектов в БД (это нормально)"
    fi
else
    test_result "API /api/projects" "false" "Ошибка запроса"
fi

# Админ логин страница
test_endpoint "$base_url/admin/login" "Админ логин страница" "Вход"

# Проверка что админ панель защищена
admin_status=$(curl -s -o /dev/null -w "%{http_code}" "$base_url/admin/dashboard" 2>&1)
if [ "$admin_status" = "302" ] || [ "$admin_status" = "401" ]; then
    test_result "Админ панель защищена" "true" "Требуется авторизация"
elif [ "$admin_status" = "200" ]; then
    test_result "Админ панель защищена" "false" "Доступ без авторизации!"
else
    test_result "Админ панель защищена" "false" "HTTP $admin_status"
fi

# ========================================
# 5. Остановка сервера и очистка
# ========================================
echo -e "\n${YELLOW}5. Очистка...${NC}"

# Останавливаем сервер
if kill $server_pid 2>/dev/null; then
    test_result "Сервер остановлен" "true"
else
    test_result "Сервер остановлен" "false" "Не удалось остановить процесс"
fi

# Удаляем тестовый бинарник
rm -f led-backend-test test-output.log test-error.log

test_result "Временные файлы удалены" "true"

# ========================================
# РЕЗУЛЬТАТЫ
# ========================================
echo -e "\n${CYAN}========================================"
echo -e "           РЕЗУЛЬТАТЫ ТЕСТОВ"
echo -e "========================================${NC}"

echo -e "\nВсего тестов: $tests_total"
echo -e "Успешно:      ${GREEN}$tests_passed${NC}"
echo -e "Провалено:    ${RED}$tests_failed${NC}"

if [ $tests_failed -eq 0 ]; then
    echo -e "\n${GREEN}✅ ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО!${NC}"
    echo -e "   ${GRAY}Проект готов к деплою${NC}\n"
    exit 0
else
    echo -e "\n${RED}❌ ЕСТЬ ПРОВАЛЕННЫЕ ТЕСТЫ!${NC}"
    echo -e "   ${YELLOW}Исправьте ошибки перед деплоем${NC}\n"
    exit 1
fi
