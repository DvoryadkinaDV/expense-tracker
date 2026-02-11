package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dvoryadkinadv/expense-tracker/internal/models"
)

// MockExpenseRepository - мок репозитория для тестов
// Вместо реальной БД храним данные в памяти
type MockExpenseRepository struct {
	expenses map[int64]*models.Expense
	lastID   int64
}

func NewMockRepository() *MockExpenseRepository {
	return &MockExpenseRepository{
		expenses: make(map[int64]*models.Expense),
		lastID:   0,
	}
}

func (m *MockExpenseRepository) Create(ctx context.Context, expense *models.Expense) error {
	m.lastID++
	expense.ID = m.lastID
	expense.CreatedAt = time.Now()
	m.expenses[expense.ID] = expense
	return nil
}

func (m *MockExpenseRepository) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	if expense, ok := m.expenses[id]; ok {
		return expense, nil
	}
	return nil, nil
}

func (m *MockExpenseRepository) GetAll(ctx context.Context, filter models.ExpenseFilter) ([]models.Expense, error) {
	var result []models.Expense
	for _, e := range m.expenses {
		// Простая фильтрация по категории
		if filter.Category != "" && e.Category != filter.Category {
			continue
		}
		result = append(result, *e)
	}
	return result, nil
}

func (m *MockExpenseRepository) Update(ctx context.Context, id int64, req models.UpdateExpenseRequest) (*models.Expense, error) {
	expense, ok := m.expenses[id]
	if !ok {
		return nil, nil
	}

	if req.Description != nil {
		expense.Description = *req.Description
	}
	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.Category != nil {
		expense.Category = *req.Category
	}

	return expense, nil
}

func (m *MockExpenseRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := m.expenses[id]; !ok {
		return errors.New("not found")
	}
	delete(m.expenses, id)
	return nil
}

func (m *MockExpenseRepository) GetStats(ctx context.Context) (*models.ExpenseStats, error) {
	stats := &models.ExpenseStats{
		ByCategory: make(map[string]float64),
	}

	for _, e := range m.expenses {
		stats.TotalAmount += e.Amount
		stats.ExpenseCount++
		stats.ByCategory[e.Category] += e.Amount
	}

	if stats.ExpenseCount > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.ExpenseCount)
	}

	return stats, nil
}

func (m *MockExpenseRepository) GetCategories(ctx context.Context) ([]string, error) {
	categoryMap := make(map[string]bool)
	for _, e := range m.expenses {
		categoryMap[e.Category] = true
	}

	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories, nil
}

// Тесты

func TestCreateExpense_Success(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	req := models.CreateExpenseRequest{
		Description: "Кофе в Старбаксе",
		Amount:      350.0,
		Category:    "Еда",
		Date:        "2024-01-15",
	}

	expense, err := svc.CreateExpense(ctx, req)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if expense.ID == 0 {
		t.Error("ID должен быть установлен")
	}

	if expense.Description != req.Description {
		t.Errorf("Description: ожидали %s, получили %s", req.Description, expense.Description)
	}

	if expense.Amount != req.Amount {
		t.Errorf("Amount: ожидали %f, получили %f", req.Amount, expense.Amount)
	}
}

func TestCreateExpense_InvalidDate(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	req := models.CreateExpenseRequest{
		Description: "Что-то",
		Amount:      100.0,
		Category:    "Разное",
		Date:        "некорректная-дата",
	}

	_, err := svc.CreateExpense(ctx, req)

	if err == nil {
		t.Error("Ожидали ошибку при некорректной дате")
	}
}

func TestGetExpense_NotFound(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	_, err := svc.GetExpense(ctx, 999)

	if err == nil {
		t.Error("Ожидали ошибку при запросе несуществующего расхода")
	}
}

func TestGetExpense_Success(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Сначала создаём расход
	req := models.CreateExpenseRequest{
		Description: "Обед",
		Amount:      500.0,
		Category:    "Еда",
		Date:        "2024-01-15",
	}

	created, _ := svc.CreateExpense(ctx, req)

	// Теперь получаем его
	expense, err := svc.GetExpense(ctx, created.ID)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if expense.ID != created.ID {
		t.Errorf("ID: ожидали %d, получили %d", created.ID, expense.ID)
	}
}

func TestGetExpenses_WithFilter(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Создаём несколько расходов в разных категориях
	expenses := []models.CreateExpenseRequest{
		{Description: "Кофе", Amount: 200, Category: "Еда", Date: "2024-01-15"},
		{Description: "Такси", Amount: 500, Category: "Транспорт", Date: "2024-01-15"},
		{Description: "Обед", Amount: 400, Category: "Еда", Date: "2024-01-16"},
	}

	for _, req := range expenses {
		svc.CreateExpense(ctx, req)
	}

	// Фильтруем по категории "Еда"
	filter := models.ExpenseFilter{Category: "Еда"}
	result, err := svc.GetExpenses(ctx, filter)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Ожидали 2 расхода в категории 'Еда', получили %d", len(result))
	}
}

func TestUpdateExpense_PartialUpdate(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Создаём расход
	req := models.CreateExpenseRequest{
		Description: "Старое описание",
		Amount:      100.0,
		Category:    "Разное",
		Date:        "2024-01-15",
	}

	created, _ := svc.CreateExpense(ctx, req)

	// Обновляем только описание
	newDescription := "Новое описание"
	updateReq := models.UpdateExpenseRequest{
		Description: &newDescription,
	}

	updated, err := svc.UpdateExpense(ctx, created.ID, updateReq)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if updated.Description != newDescription {
		t.Errorf("Description: ожидали %s, получили %s", newDescription, updated.Description)
	}

	// Сумма должна остаться прежней
	if updated.Amount != created.Amount {
		t.Errorf("Amount не должен был измениться: ожидали %f, получили %f", created.Amount, updated.Amount)
	}
}

func TestDeleteExpense_Success(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Создаём расход
	req := models.CreateExpenseRequest{
		Description: "Для удаления",
		Amount:      100.0,
		Category:    "Тест",
		Date:        "2024-01-15",
	}

	created, _ := svc.CreateExpense(ctx, req)

	// Удаляем
	err := svc.DeleteExpense(ctx, created.ID)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	// Проверяем, что расход больше не существует
	_, err = svc.GetExpense(ctx, created.ID)
	if err == nil {
		t.Error("Расход должен был быть удалён")
	}
}

func TestGetStats(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Создаём расходы
	expenses := []models.CreateExpenseRequest{
		{Description: "Раз", Amount: 100, Category: "Еда", Date: "2024-01-15"},
		{Description: "Два", Amount: 200, Category: "Еда", Date: "2024-01-15"},
		{Description: "Три", Amount: 300, Category: "Транспорт", Date: "2024-01-15"},
	}

	for _, req := range expenses {
		svc.CreateExpense(ctx, req)
	}

	stats, err := svc.GetStats(ctx)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if stats.TotalAmount != 600 {
		t.Errorf("TotalAmount: ожидали 600, получили %f", stats.TotalAmount)
	}

	if stats.ExpenseCount != 3 {
		t.Errorf("ExpenseCount: ожидали 3, получили %d", stats.ExpenseCount)
	}

	if stats.AverageAmount != 200 {
		t.Errorf("AverageAmount: ожидали 200, получили %f", stats.AverageAmount)
	}

	if stats.ByCategory["Еда"] != 300 {
		t.Errorf("ByCategory[Еда]: ожидали 300, получили %f", stats.ByCategory["Еда"])
	}
}

func TestGetExpenses_DefaultLimit(t *testing.T) {
	repo := NewMockRepository()
	svc := NewExpenseService(repo)
	ctx := context.Background()

	// Проверяем, что при пустом фильтре устанавливается дефолтный лимит
	filter := models.ExpenseFilter{}

	// Этот тест просто проверяет, что метод не падает
	_, err := svc.GetExpenses(ctx, filter)

	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}
}
