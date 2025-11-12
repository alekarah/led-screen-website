# Multi-stage build для уменьшения размера финального образа

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

WORKDIR /app

# Копируем go mod файлы и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./main.go

# Stage 2: Runtime
FROM alpine:latest

# Устанавливаем CA certificates для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/server .

# Создаем директории для статики и загрузок
RUN mkdir -p /root/static/uploads

# Копируем статические файлы (если нужны в образе)
# COPY --from=builder /app/static ./static

# Порт приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./server"]
