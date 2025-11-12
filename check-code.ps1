# Скрипт для проверки кода перед push
# Запуск: .\check-code.ps1

# Устанавливаем кодировку UTF-8 для корректного отображения эмодзи
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$PSDefaultParameterValues['*:Encoding'] = 'utf8'

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Code Check Before Push" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Переходим в директорию backend
Set-Location backend

# 1. Проверка тестов
Write-Host "[1/3] Running tests..." -ForegroundColor Yellow
go test ./... -v
if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "FAILED: Tests failed!" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Write-Host "OK: All tests passed!" -ForegroundColor Green
Write-Host ""

# 2. Форматирование кода
Write-Host "[2/3] Formatting code (gofmt)..." -ForegroundColor Yellow
go fmt ./...
Write-Host "OK: Code formatted!" -ForegroundColor Green
Write-Host ""

# 3. Проверка линтером (если установлен)
Write-Host "[3/3] Running linter..." -ForegroundColor Yellow
if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
    golangci-lint run ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "FAILED: Linter found errors!" -ForegroundColor Red
        Set-Location ..
        exit 1
    }
    Write-Host "OK: No linter errors!" -ForegroundColor Green
} else {
    Write-Host "SKIPPED: golangci-lint not installed (will check in CI/CD)" -ForegroundColor Yellow
    Write-Host "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" -ForegroundColor Gray
}
Write-Host ""

# Возвращаемся в корень
Set-Location ..

# Итог
Write-Host "================================" -ForegroundColor Cyan
Write-Host "SUCCESS: All checks passed!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "You can now safely:" -ForegroundColor White
Write-Host "  git add ." -ForegroundColor Gray
Write-Host "  git commit -m `"Your message`"" -ForegroundColor Gray
Write-Host "  git push" -ForegroundColor Gray
Write-Host ""
