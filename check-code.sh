#!/bin/bash
# Скрипт для проверки кода перед push
# Запуск: ./check-code.sh

set -e  # Выход при первой ошибке

echo ""
echo "================================"
echo "Code Check Before Push"
echo "================================"
echo ""

# Переходим в директорию backend
cd backend

# 1. Проверка тестов
echo "[1/3] Running tests..."
if go test ./... -v; then
    echo "OK: All tests passed!"
    echo ""
else
    echo ""
    echo "FAILED: Tests failed!"
    cd ..
    exit 1
fi

# 2. Форматирование кода
echo "[2/3] Formatting code (gofmt)..."
go fmt ./...
echo "OK: Code formatted!"
echo ""

# 3. Проверка линтером (если установлен)
echo "[3/3] Running linter..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./...; then
        echo "OK: No linter errors!"
    else
        echo ""
        echo "FAILED: Linter found errors!"
        cd ..
        exit 1
    fi
else
    echo "SKIPPED: golangci-lint not installed (will check in CI/CD)"
    echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi
echo ""

# Возвращаемся в корень
cd ..

# Итог
echo "================================"
echo "SUCCESS: All checks passed!"
echo "================================"
echo ""
echo "You can now safely:"
echo "  git add ."
echo "  git commit -m \"Your message\""
echo "  git push"
echo ""
