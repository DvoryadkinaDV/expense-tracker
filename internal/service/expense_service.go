package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dvoryadkinadv/expense-tracker/internal/models"
)

// ExpenseRepository описывает интерфейс работы с хранилищем
// Использую интерфейс, чтобы можно было подменить реализацию в тестах
type ExpenseRepository interface {
	Create(ctx context.Context, expense *models.Expense) error
	GetByID(ctx context.Context, id int64) (*models.Expense, error)
	GetAll(ctx context.Context, filter models.ExpenseFilter) ([]models.Expense, error)
	Update(ctx context.Context, id int64, req models.UpdateExpenseRequest) (*models.Expense, error)
	Delete(ctx context.Context, id int64) error
	GetStats(ctx context.Context) (*models.ExpenseStats, error)
	GetCategories(ctx context.Context) ([]string, error)
}

// ExpenseService содержит бизнес-логику работы с расходами
// Пока тут всё просто, но в будущем можно добавить валидацию,
// нотификации, логирование и прочее
type ExpenseService struct {
	repo ExpenseRepository
}

// NewExpenseService создаёт новый сервис
func NewExpenseService(repo ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

// CreateExpense создаёт новый расход
func (s *ExpenseService) CreateExpense(ctx context.Context, req models.CreateExpenseRequest) (*models.Expense, error) {
	// Парсим дату
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты, используйте YYYY-MM-DD: %w", err)
	}

	expense := &models.Expense{
		Description: req.Description,
		Amount:      req.Amount,
		Category:    req.Category,
		Date:        date,
	}

	if err := s.repo.Create(ctx, expense); err != nil {
		return nil, err
	}

	return expense, nil
}

// GetExpense возвращает расход по ID
func (s *ExpenseService) GetExpense(ctx context.Context, id int64) (*models.Expense, error) {
	expense, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if expense == nil {
		return nil, fmt.Errorf("расход с id=%d не найден", id)
	}

	return expense, nil
}

// GetExpenses возвращает список расходов с фильтрацией
func (s *ExpenseService) GetExpenses(ctx context.Context, filter models.ExpenseFilter) ([]models.Expense, error) {
	// Устанавливаем дефолтный лимит, чтобы не выгружать всю базу
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Максимум 100 записей за раз - защита от случайных злоупотреблений
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.repo.GetAll(ctx, filter)
}

// UpdateExpense обновляет расход
func (s *ExpenseService) UpdateExpense(ctx context.Context, id int64, req models.UpdateExpenseRequest) (*models.Expense, error) {
	// Проверяем, существует ли расход
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, fmt.Errorf("расход с id=%d не найден", id)
	}

	return s.repo.Update(ctx, id, req)
}

// DeleteExpense удаляет расход
func (s *ExpenseService) DeleteExpense(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// GetStats возвращает статистику
func (s *ExpenseService) GetStats(ctx context.Context) (*models.ExpenseStats, error) {
	return s.repo.GetStats(ctx)
}

// GetCategories возвращает список категорий
func (s *ExpenseService) GetCategories(ctx context.Context) ([]string, error) {
	return s.repo.GetCategories(ctx)
}
