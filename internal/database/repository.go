package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dvoryadkinadv/expense-tracker/internal/models"
	"github.com/jmoiron/sqlx"
)

// ExpenseRepository - репозиторий для работы с расходами
// Использую паттерн Repository, чтобы отделить логику работы с БД
// от бизнес-логики и HTTP-обработчиков
type ExpenseRepository struct {
	db *sqlx.DB
}

// NewExpenseRepository создаёт новый репозиторий
func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create добавляет новый расход в БД
func (r *ExpenseRepository) Create(ctx context.Context, expense *models.Expense) error {
	query := `
		INSERT INTO expenses (description, amount, category, date, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	expense.CreatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx, query,
		expense.Description, expense.Amount, expense.Category,
		expense.Date, expense.CreatedAt,
	).Scan(&expense.ID)

	if err != nil {
		return fmt.Errorf("ошибка создания расхода: %w", err)
	}

	return nil
}

// GetByID возвращает расход по ID
func (r *ExpenseRepository) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	var expense models.Expense

	query := `SELECT id, description, amount, category, date, created_at FROM expenses WHERE id = $1`

	err := r.db.GetContext(ctx, &expense, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // расход не найден - это нормально, не ошибка
		}
		return nil, fmt.Errorf("ошибка получения расхода: %w", err)
	}

	return &expense, nil
}

// GetAll возвращает список расходов с фильтрацией
// Тут немного магии со строками, но зато гибко!
func (r *ExpenseRepository) GetAll(ctx context.Context, filter models.ExpenseFilter) ([]models.Expense, error) {
	var expenses []models.Expense
	var args []interface{}
	var conditions []string

	query := `SELECT id, description, amount, category, date, created_at FROM expenses`
	argNum := 1

	// Собираем условия фильтрации
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argNum))
		args = append(args, filter.Category)
		argNum++
	}

	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argNum))
		args = append(args, filter.DateFrom)
		argNum++
	}

	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argNum))
		args = append(args, filter.DateTo)
		argNum++
	}

	// Добавляем WHERE если есть условия
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Сортировка по дате (новые сверху)
	query += " ORDER BY date DESC, id DESC"

	// Пагинация
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	err := r.db.SelectContext(ctx, &expenses, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка расходов: %w", err)
	}

	// Если ничего не нашли - возвращаем пустой слайс, а не nil
	if expenses == nil {
		expenses = []models.Expense{}
	}

	return expenses, nil
}

// Update обновляет расход
func (r *ExpenseRepository) Update(ctx context.Context, id int64, req models.UpdateExpenseRequest) (*models.Expense, error) {
	var sets []string
	var args []interface{}
	argNum := 1

	// Обновляем только те поля, которые переданы
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argNum))
		args = append(args, *req.Description)
		argNum++
	}

	if req.Amount != nil {
		sets = append(sets, fmt.Sprintf("amount = $%d", argNum))
		args = append(args, *req.Amount)
		argNum++
	}

	if req.Category != nil {
		sets = append(sets, fmt.Sprintf("category = $%d", argNum))
		args = append(args, *req.Category)
		argNum++
	}

	if req.Date != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты: %w", err)
		}
		sets = append(sets, fmt.Sprintf("date = $%d", argNum))
		args = append(args, parsedDate)
		argNum++
	}

	// Если нечего обновлять - просто возвращаем текущую запись
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(
		`UPDATE expenses SET %s WHERE id = $%d
		 RETURNING id, description, amount, category, date, created_at`,
		strings.Join(sets, ", "), argNum,
	)
	args = append(args, id)

	var expense models.Expense
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&expense.ID, &expense.Description, &expense.Amount,
		&expense.Category, &expense.Date, &expense.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка обновления расхода: %w", err)
	}

	return &expense, nil
}

// Delete удаляет расход по ID
func (r *ExpenseRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM expenses WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления расхода: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("расход с id=%d не найден", id)
	}

	return nil
}

// GetStats возвращает статистику по расходам
func (r *ExpenseRepository) GetStats(ctx context.Context) (*models.ExpenseStats, error) {
	stats := &models.ExpenseStats{
		ByCategory: make(map[string]float64),
	}

	// Общая статистика
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0), COUNT(*), COALESCE(AVG(amount), 0)
		FROM expenses
	`).Scan(&stats.TotalAmount, &stats.ExpenseCount, &stats.AverageAmount)

	if err != nil {
		return nil, fmt.Errorf("ошибка получения общей статистики: %w", err)
	}

	// Статистика по категориям
	rows, err := r.db.QueryContext(ctx, `
		SELECT category, COALESCE(SUM(amount), 0)
		FROM expenses
		GROUP BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики по категориям: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			return nil, err
		}
		stats.ByCategory[category] = amount
	}

	return stats, nil
}

// GetCategories возвращает список уникальных категорий
func (r *ExpenseRepository) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string

	err := r.db.SelectContext(ctx, &categories, `
		SELECT DISTINCT category FROM expenses ORDER BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения категорий: %w", err)
	}

	if categories == nil {
		categories = []string{}
	}

	return categories, nil
}
