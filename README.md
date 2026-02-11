# Expense Tracker API

Простой REST API для трекинга личных расходов.

## Что умеет?

- Добавлять расходы с описанием, суммой, категорией и датой
- Получать список расходов с фильтрацией по категории и дате
- Редактировать и удалять расходы
- Показывать статистику: общая сумма, средний расход, расходы по категориям
- Возвращать список используемых категорий

## Технологии

- **Go 1.21** 
- **Gin** - веб-фреймворк (лёгкий, быстрый, документация хорошая)
- **PostgreSQL** - надёжная БД
- **Docker** - для деплоя
- **sqlx** - удобная обёртка над database/sql

## Быстрый старт

### Вариант 1: Docker 

```bash
# Клонируем репозиторий
git clone https://github.com/dvoryadkinadv/expense-tracker.git
cd expense-tracker

# Запускаем
make docker-up

# или без make:
docker-compose up -d
```

API доступен по адресу: http://localhost:8080

### Вариант 2: Локальная разработка

Если запускать без Docker (например, для дебага):

1. **Установите PostgreSQL** и создайте базу:
```bash
createdb expense_tracker
```

2. **Примените миграции:**
```bash
psql -d expense_tracker -f migrations/001_create_expenses_table.sql
```

3. **Установите зависимости и запустите:**
```bash
make deps
make run
```

Или вручную:
```bash
go mod download
go run ./cmd/server/main.go
```

### Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DB_HOST` | localhost | Хост базы данных |
| `DB_PORT` | 5432 | Порт базы данных |
| `DB_USER` | postgres | Пользователь БД |
| `DB_PASSWORD` | postgres | Пароль БД |
| `DB_NAME` | expense_tracker | Название базы |
| `DB_SSLMODE` | disable | SSL режим |
| `PORT` | 8080 | Порт API сервера |
| `GIN_MODE` | debug | Режим Gin (debug/release) |

## API Endpoints

### Health Check
```
GET /health
```
Проверка, что сервис работает.

### Расходы

#### Создать расход
```
POST /api/expenses
Content-Type: application/json

{
  "description": "Кофе в Старбаксе",
  "amount": 350.00,
  "category": "Еда",
  "date": "2024-01-15"
}
```

#### Получить все расходы
```
GET /api/expenses
GET /api/expenses?category=Еда
GET /api/expenses?date_from=2024-01-01&date_to=2024-01-31
GET /api/expenses?limit=10&offset=0
```

#### Получить расход по ID
```
GET /api/expenses/{id}
```

#### Обновить расход
```
PUT /api/expenses/{id}
Content-Type: application/json

{
  "description": "Кофе и круассан",
  "amount": 500.00
}
```
*Можно обновлять только нужные поля*

#### Удалить расход
```
DELETE /api/expenses/{id}
```

### Статистика
```
GET /api/stats
```

Возвращает:
```json
{
  "success": true,
  "data": {
    "total_amount": 15000.50,
    "expense_count": 42,
    "average_amount": 357.15,
    "by_category": {
      "Еда": 5000.00,
      "Транспорт": 3000.00,
      "Развлечения": 7000.50
    }
  }
}
```

### Категории
```
GET /api/categories
```

## Примеры использования (curl)

```bash
# Создать расход
curl -X POST http://localhost:8080/api/expenses \
  -H "Content-Type: application/json" \
  -d '{"description":"Обед в кафе","amount":450,"category":"Еда","date":"2024-01-15"}'

# Получить все расходы
curl http://localhost:8080/api/expenses

# Получить расходы по категории
curl "http://localhost:8080/api/expenses?category=Еда"

# Получить статистику
curl http://localhost:8080/api/stats

# Обновить расход
curl -X PUT http://localhost:8080/api/expenses/1 \
  -H "Content-Type: application/json" \
  -d '{"amount":500}'

# Удалить расход
curl -X DELETE http://localhost:8080/api/expenses/1
```

## Makefile команды

```bash
make help          # Показать все команды
make build         # Собрать бинарник
make run           # Запустить локально
make test          # Запустить тесты
make test-cover    # Тесты с покрытием
make lint          # Проверить линтером
make docker-up     # Запустить в Docker
make docker-down   # Остановить Docker
make docker-logs   # Логи контейнеров
make clean         # Очистить артефакты
```

## Тестирование

```bash
# Запустить все тесты
make test

# Тесты с отчётом о покрытии
make test-cover
# Откроется файл coverage.html с детальным отчётом
```

## Что можно улучшить

- [ ] Аутентификация пользователей (JWT)
- [ ] Привязка расходов к пользователям
- [ ] Swagger документация
- [ ] Экспорт в CSV/Excel
- [ ] Графики и визуализация
- [ ] Бюджеты и лимиты по категориям
- [ ] Повторяющиеся расходы
- [ ] Интеграция с банками


