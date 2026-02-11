-- Миграция для создания таблицы расходов
--  id, описание, сумма, категория, дата

CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    category VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для ускорения часто используемых запросов
-- Категория - для фильтрации
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);

-- Дата - для фильтрации по периоду и сортировки
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date DESC);

-- Комментарий: в реальном проекте тут была бы еще таблица users
-- и поле user_id для привязки расходов к пользователю
