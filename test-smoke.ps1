# ========================================
# Smoke Test для LED Screen Website
# ========================================
# Быстрая автоматическая проверка критичных функций
# Время выполнения: ~30 секунд
# ========================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SMOKE TEST - LED Screen Website" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Continue"
$testsPassed = 0
$testsFailed = 0
$testsTotal = 0

# Функция для вывода результата теста
function Test-Result {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Details = ""
    )

    $script:testsTotal++

    if ($Success) {
        Write-Host "[PASS]" -ForegroundColor Green -NoNewline
        Write-Host " $TestName" -ForegroundColor White
        if ($Details) {
            Write-Host "       $Details" -ForegroundColor Gray
        }
        $script:testsPassed++
    } else {
        Write-Host "[FAIL]" -ForegroundColor Red -NoNewline
        Write-Host " $TestName" -ForegroundColor White
        if ($Details) {
            Write-Host "       $Details" -ForegroundColor Yellow
        }
        $script:testsFailed++
    }
}

# ========================================
# 1. Проверка зависимостей
# ========================================
Write-Host "`n1. Проверка зависимостей..." -ForegroundColor Yellow

# Go установлен?
$goVersion = & go version 2>&1
if ($LASTEXITCODE -eq 0) {
    Test-Result "Go установлен" $true $goVersion
} else {
    Test-Result "Go установлен" $false "Go не найден в PATH"
}

# PostgreSQL запущен?
try {
    $pgProcess = Get-Process -Name postgres -ErrorAction SilentlyContinue
    if ($pgProcess) {
        Test-Result "PostgreSQL запущен" $true "PID: $($pgProcess.Id)"
    } else {
        Test-Result "PostgreSQL запущен" $false "Процесс postgres не найден"
    }
} catch {
    Test-Result "PostgreSQL запущен" $false "Не удалось проверить"
}

# ========================================
# 2. Сборка проекта
# ========================================
Write-Host "`n2. Сборка проекта..." -ForegroundColor Yellow

Set-Location "$PSScriptRoot\backend"

$buildOutput = & go build -o led-backend-test.exe main.go 2>&1
if ($LASTEXITCODE -eq 0) {
    Test-Result "Проект собирается" $true "Бинарник создан"
} else {
    Test-Result "Проект собирается" $false "Ошибка компиляции"
    Write-Host $buildOutput -ForegroundColor Red
    exit 1
}

# ========================================
# 3. Запуск сервера
# ========================================
Write-Host "`n3. Запуск тестового сервера..." -ForegroundColor Yellow

# Проверяем что .env существует
if (-not (Test-Path ".env")) {
    Test-Result ".env файл существует" $false "Создайте .env файл"
    Remove-Item led-backend-test.exe -ErrorAction SilentlyContinue
    exit 1
} else {
    Test-Result ".env файл существует" $true
}

# Запускаем сервер в фоне
$serverProcess = Start-Process -FilePath ".\led-backend-test.exe" -PassThru -NoNewWindow -RedirectStandardOutput "test-output.log" -RedirectStandardError "test-error.log"

if ($serverProcess) {
    Test-Result "Сервер запущен" $true "PID: $($serverProcess.Id)"
} else {
    Test-Result "Сервер запущен" $false
    Remove-Item led-backend-test.exe -ErrorAction SilentlyContinue
    exit 1
}

# Ждем пока сервер запустится
Write-Host "`n   Ожидание запуска сервера..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# Проверяем что процесс еще жив
if ($serverProcess.HasExited) {
    Test-Result "Сервер продолжает работать" $false "Сервер упал при запуске"
    Get-Content "test-error.log" -ErrorAction SilentlyContinue | Write-Host -ForegroundColor Red
    Remove-Item led-backend-test.exe -ErrorAction SilentlyContinue
    exit 1
} else {
    Test-Result "Сервер продолжает работать" $true
}

# ========================================
# 4. HTTP тесты
# ========================================
Write-Host "`n4. Тестирование HTTP endpoints..." -ForegroundColor Yellow

$baseUrl = "http://localhost:8080"

# Функция для HTTP запроса
function Test-Endpoint {
    param(
        [string]$Url,
        [string]$Name,
        [string]$ExpectedContent = ""
    )

    try {
        $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop

        if ($response.StatusCode -eq 200) {
            if ($ExpectedContent -and $response.Content -notmatch [regex]::Escape($ExpectedContent)) {
                Test-Result $Name $false "Страница не содержит '$ExpectedContent'"
            } else {
                Test-Result $Name $true "HTTP 200 OK"
            }
        } else {
            Test-Result $Name $false "HTTP $($response.StatusCode)"
        }
    } catch {
        Test-Result $Name $false "Ошибка запроса: $($_.Exception.Message)"
    }
}

# Публичные страницы
Test-Endpoint "$baseUrl/" "Главная страница" "LED"
Test-Endpoint "$baseUrl/projects" "Страница проектов"
Test-Endpoint "$baseUrl/services" "Страница услуг"
Test-Endpoint "$baseUrl/contact" "Страница контактов"

# API endpoints
try {
    $apiResponse = Invoke-RestMethod -Uri "$baseUrl/api/projects" -Method Get -TimeoutSec 5
    if ($apiResponse.projects -ne $null) {
        Test-Result "API /api/projects" $true "Получены данные проектов"
    } else {
        Test-Result "API /api/projects" $true "Нет проектов в БД (это нормально)"
    }
} catch {
    Test-Result "API /api/projects" $false $_.Exception.Message
}

# Админ логин страница
Test-Endpoint "$baseUrl/admin/login" "Админ логин страница" "Вход"

# Проверка что админ панель защищена (должен редиректить на логин)
try {
    $adminResponse = Invoke-WebRequest -Uri "$baseUrl/admin/dashboard" -UseBasicParsing -MaximumRedirection 0 -ErrorAction SilentlyContinue
    # Если редирект (302) или Unauthorized (401) - это правильно
    if ($adminResponse.StatusCode -in @(302, 401)) {
        Test-Result "Админ панель защищена" $true "Требуется авторизация"
    } elseif ($adminResponse.StatusCode -eq 200) {
        Test-Result "Админ панель защищена" $false "Доступ без авторизации!"
    }
} catch {
    # 302 редирект выбросит исключение - это нормально
    if ($_.Exception.Response.StatusCode -in @(302, 'Redirect')) {
        Test-Result "Админ панель защищена" $true "Требуется авторизация"
    } else {
        Test-Result "Админ панель защищена" $false $_.Exception.Message
    }
}

# ========================================
# 5. Остановка сервера и очистка
# ========================================
Write-Host "`n5. Очистка..." -ForegroundColor Yellow

# Останавливаем сервер
try {
    Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
    Test-Result "Сервер остановлен" $true
} catch {
    Test-Result "Сервер остановлен" $false "Не удалось остановить процесс"
}

# Удаляем тестовый бинарник
Remove-Item led-backend-test.exe -ErrorAction SilentlyContinue
Remove-Item test-output.log -ErrorAction SilentlyContinue
Remove-Item test-error.log -ErrorAction SilentlyContinue

Test-Result "Временные файлы удалены" $true

# ========================================
# РЕЗУЛЬТАТЫ
# ========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "           РЕЗУЛЬТАТЫ ТЕСТОВ" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nВсего тестов: $testsTotal" -ForegroundColor White
Write-Host "Успешно:      " -NoNewline; Write-Host $testsPassed -ForegroundColor Green
Write-Host "Провалено:    " -NoNewline; Write-Host $testsFailed -ForegroundColor Red

if ($testsFailed -eq 0) {
    Write-Host "`n✅ ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО!" -ForegroundColor Green
    Write-Host "   Проект готов к деплою`n" -ForegroundColor Gray
    exit 0
} else {
    Write-Host "`n❌ ЕСТЬ ПРОВАЛЕННЫЕ ТЕСТЫ!" -ForegroundColor Red
    Write-Host "   Исправьте ошибки перед деплоем`n" -ForegroundColor Yellow
    exit 1
}
