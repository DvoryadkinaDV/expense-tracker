package models

import (
	"time"
)

// Expense представляет расход пользователя
// Простая структура - без лишних полей,
// чтобы не усложнять жизнь себе и тем, кто будет это читать
type Expense struct {
	ID          int64     `json:"id" db:"id"`
	Description string    `json:"description" db:"description"`
	Amount      float64   `json:"amount" db:"amount"`
	Category    string    `json:"category" db:"category"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CreateExpenseRequest - то, что приходит от клиента при создании расхода
// Валидацию делаю через теги binding - Gin сам всё проверит
type CreateExpenseRequest struct {
	Description string  `json:"description" binding:"required,min=1,max=500"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required,min=1,max=100"`
	Date        string  `json:"date" binding:"required"` // формат: 2024-01-15
}

// UpdateExpenseRequest - для обновления расхода
// Все поля опциональные, обновляем только то, что прислали
type UpdateExpenseRequest struct {
	Description *string  `json:"description,omitempty" binding:"omitempty,min=1,max=500"`
	Amount      *float64 `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Category    *string  `json:"category,omitempty" binding:"omitempty,min=1,max=100"`
	Date        *string  `json:"date,omitempty"`
}

// ExpenseFilter - фильтры для списка расходов
// Сделать фильтрацию гибкой, но не переусложнить
type ExpenseFilter struct {
	Category string
	DateFrom string
	DateTo   string
	Limit    int
	Offset   int
}

// ExpenseStats - статистика по расходам
type ExpenseStats struct {
	TotalAmount   float64            `json:"total_amount"`
	ExpenseCount  int                `json:"expense_count"`
	AverageAmount float64            `json:"average_amount"`
	ByCategory    map[string]float64 `json:"by_category"`
}
