# Makefile для expense-tracker
# Упрощает типичные операции разработки

# Переменные
APP_NAME = expense-tracker
MAIN_PATH = ./cmd/server
DOCKER_COMPOSE = docker-compose

# Цвета для красивого вывода (работает в большинстве терминалов)
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

.PHONY: help build run test clean docker-build docker-up docker-down migrate lint

# По умолчанию показываем справку
help:
	@echo "$(GREEN)Доступные команды:$(NC)"
	@echo "  $(YELLOW)make build$(NC)        - Собрать бинарник"
	@echo "  $(YELLOW)make run$(NC)          - Запустить локально (нужен PostgreSQL)"
	@echo "  $(YELLOW)make test$(NC)         - Запустить тесты"
	@echo "  $(YELLOW)make test-cover$(NC)   - Тесты с отчётом о покрытии"
	@echo "  $(YELLOW)make lint$(NC)         - Проверить код линтером"
	@echo "  $(YELLOW)make docker-build$(NC) - Собрать Docker образ"
	@echo "  $(YELLOW)make docker-up$(NC)    - Запустить в Docker (с БД)"
	@echo "  $(YELLOW)make docker-down$(NC)  - Остановить Docker контейнеры"
	@echo "  $(YELLOW)make docker-logs$(NC)  - Показать логи контейнеров"
	@echo "  $(YELLOW)make clean$(NC)        - Очистить артефакты сборки"
	@echo "  $(YELLOW)make deps$(NC)         - Скачать зависимости"

# Скачать зависимости
deps:
	@echo "$(GREEN)Скачиваю зависимости...$(NC)"
	go mod download
	go mod tidy

# Собрать бинарник
build:
	@echo "$(GREEN)Собираю $(APP_NAME)...$(NC)"
	go build -o $(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Готово! Бинарник: ./$(APP_NAME)$(NC)"

# Запустить локально
run: build
	@echo "$(GREEN)Запускаю $(APP_NAME)...$(NC)"
	./$(APP_NAME)

# Запустить в режиме разработки (с hot reload через air, если установлен)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)Air не установлен. Запускаю без hot reload...$(NC)"; \
		go run $(MAIN_PATH)/main.go; \
	fi

# Запустить тесты
test:
	@echo "$(GREEN)Запускаю тесты...$(NC)"
	go test -v ./...

# Тесты с покрытием
test-cover:
	@echo "$(GREEN)Запускаю тесты с покрытием...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчёт сохранён в coverage.html$(NC)"

# Линтер (нужен golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "$(GREEN)Запускаю линтер...$(NC)"; \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint не установлен. Установите: brew install golangci-lint$(NC)"; \
	fi

# Docker команды
docker-build:
	@echo "$(GREEN)Собираю Docker образ...$(NC)"
	docker build -t $(APP_NAME) .

docker-up:
	@echo "$(GREEN)Запускаю контейнеры...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Приложение доступно по адресу: http://localhost:8080$(NC)"

docker-down:
	@echo "$(GREEN)Останавливаю контейнеры...$(NC)"
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-restart: docker-down docker-up

# Полная очистка (включая volumes)
docker-clean:
	@echo "$(YELLOW)Удаляю контейнеры и данные...$(NC)"
	$(DOCKER_COMPOSE) down -v

# Очистить артефакты сборки
clean:
	@echo "$(GREEN)Очищаю...$(NC)"
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	go clean

# Применить миграции (для локальной БД)
migrate:
	@echo "$(GREEN)Применяю миграции...$(NC)"
	@for f in migrations/*.sql; do \
		echo "Выполняю $$f..."; \
		PGPASSWORD=postgres psql -h localhost -U postgres -d expense_tracker -f $$f; \
	done
	@echo "$(GREEN)Миграции применены$(NC)"

# Создать базу данных локально
create-db:
	@echo "$(GREEN)Создаю базу данных...$(NC)"
	PGPASSWORD=postgres createdb -h localhost -U postgres expense_tracker || true
	@echo "$(GREEN)База создана$(NC)"

# Быстрый старт: создать БД + миграции + запуск
quickstart: create-db migrate run
