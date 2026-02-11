# Dockerfile для expense-tracker
# Используем multi-stage build для минимального размера образа

# === Этап 1: Сборка ===
FROM golang:1.24-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum отдельно для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
# CGO_ENABLED=0 - статическая линковка, без зависимостей от libc
# -ldflags="-w -s" - уменьшаем размер бинарника, убирая отладочную информацию
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /expense-tracker ./cmd/server

# === Этап 2: Финальный образ ===
FROM alpine:3.19

# Добавляем сертификаты для HTTPS и tzdata для работы с временными зонами
RUN apk --no-cache add ca-certificates tzdata

# Создаём непривилегированного пользователя (best practice)
RUN adduser -D -g '' appuser

WORKDIR /app

# Копируем только бинарник из builder-а
COPY --from=builder /expense-tracker .

# Копируем миграции (понадобятся при запуске)
COPY --from=builder /app/migrations ./migrations

# Переключаемся на непривилегированного пользователя
USER appuser

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./expense-tracker"]
