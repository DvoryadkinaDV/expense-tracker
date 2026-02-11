package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config - настройки подключения к БД
// Вынесла в отдельную структуру, чтобы было проще тестировать
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection создаёт подключение к PostgreSQL
// Возвращаю sqlx.DB, потому что он удобнее стандартного sql.DB
// (есть Get, Select и прочие плюшки)
func NewConnection(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	// Проверяем, что соединение реально работает
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("БД не отвечает: %w", err)
	}

	// Настройки пула соединений
	// Для простого приложения
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
