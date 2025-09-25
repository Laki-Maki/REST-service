# 1. Базовый образ с Go
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache git bash

# Создаём рабочую директорию
WORKDIR /app

# Копируем зависимости и подтягиваем их
COPY go.mod go.sum ./
RUN go mod tidy

# Копируем весь код проекта
COPY . ./

# Собираем бинарник
RUN go build -o server ./cmd/server

# -----------------------
# 2. Финальный легковесный образ
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарник из предыдущего шага
COPY --from=builder /app/server .

# Копируем файл .env, если нужно (опционально)
# COPY .env .

# Открываем порт
EXPOSE 8080

# Команда запуска сервиса
CMD ["./server"]
