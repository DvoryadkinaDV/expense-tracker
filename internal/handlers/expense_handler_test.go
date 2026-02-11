package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dvoryadkinadv/expense-tracker/internal/models"
	"github.com/dvoryadkinadv/expense-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

// mockRepo - мок репозитория для тестов хэндлеров
type mockRepo struct {
	expenses map[int64]*models.Expense
	lastID   int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		expenses: make(map[int64]*models.Expense),
		lastID:   0,
	}
}

func (m *mockRepo) Create(ctx context.Context, expense *models.Expense) error {
	m.lastID++
	expense.ID = m.lastID
	expense.CreatedAt = time.Now()
	m.expenses[expense.ID] = expense
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	if e, ok := m.expenses[id]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockRepo) GetAll(ctx context.Context, filter models.ExpenseFilter) ([]models.Expense, error) {
	var result []models.Expense
	for _, e := range m.expenses {
		result = append(result, *e)
	}
	return result, nil
}

func (m *mockRepo) Update(ctx context.Context, id int64, req models.UpdateExpenseRequest) (*models.Expense, error) {
	e, ok := m.expenses[id]
	if !ok {
		return nil, nil
	}
	if req.Description != nil {
		e.Description = *req.Description
	}
	if req.Amount != nil {
		e.Amount = *req.Amount
	}
	return e, nil
}

func (m *mockRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := m.expenses[id]; !ok {
		return errors.New("not found")
	}
	delete(m.expenses, id)
	return nil
}

func (m *mockRepo) GetStats(ctx context.Context) (*models.ExpenseStats, error) {
	return &models.ExpenseStats{
		TotalAmount:  1000,
		ExpenseCount: 5,
		ByCategory:   map[string]float64{"Еда": 500},
	}, nil
}

func (m *mockRepo) GetCategories(ctx context.Context) ([]string, error) {
	return []string{"Еда", "Транспорт"}, nil
}

func setupTestRouter() (*gin.Engine, *mockRepo) {
	gin.SetMode(gin.TestMode)

	repo := newMockRepo()
	svc := service.NewExpenseService(repo)
	handler := NewExpenseHandler(svc)

	router := gin.New()
	api := router.Group("/api")
	{
		api.POST("/expenses", handler.CreateExpense)
		api.GET("/expenses", handler.GetExpenses)
		api.GET("/expenses/:id", handler.GetExpense)
		api.PUT("/expenses/:id", handler.UpdateExpense)
		api.DELETE("/expenses/:id", handler.DeleteExpense)
		api.GET("/stats", handler.GetStats)
		api.GET("/categories", handler.GetCategories)
	}
	router.GET("/health", HealthCheck)

	return router, repo
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "ok" {
		t.Error("Health check должен возвращать status: ok")
	}
}

func TestCreateExpense_Handler(t *testing.T) {
	router, _ := setupTestRouter()

	expense := models.CreateExpenseRequest{
		Description: "Тестовый расход",
		Amount:      100.50,
		Category:    "Тест",
		Date:        "2024-01-15",
	}

	body, _ := json.Marshal(expense)
	req, _ := http.NewRequest("POST", "/api/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Ожидали статус 201, получили %d. Body: %s", w.Code, w.Body.String())
	}

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response.Success {
		t.Error("Ответ должен быть успешным")
	}
}

func TestCreateExpense_ValidationError(t *testing.T) {
	router, _ := setupTestRouter()

	// Пустое описание - должна быть ошибка валидации
	expense := models.CreateExpenseRequest{
		Description: "",
		Amount:      100,
		Category:    "Тест",
		Date:        "2024-01-15",
	}

	body, _ := json.Marshal(expense)
	req, _ := http.NewRequest("POST", "/api/expenses", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали статус 400, получили %d", w.Code)
	}
}

func TestGetExpenses_Handler(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/expenses", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}
}

func TestGetExpense_NotFound(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/expenses/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидали статус 404, получили %d", w.Code)
	}
}

func TestGetExpense_InvalidID(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/expenses/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали статус 400, получили %d", w.Code)
	}
}

func TestGetStats_Handler(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response.Success {
		t.Error("Ответ должен быть успешным")
	}
}

func TestGetCategories_Handler(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}
}

func TestDeleteExpense_Handler(t *testing.T) {
	router, repo := setupTestRouter()

	// Сначала создаём расход напрямую в репо
	expense := &models.Expense{
		Description: "Для удаления",
		Amount:      50,
		Category:    "Тест",
		Date:        time.Now(),
	}
	repo.Create(context.Background(), expense)

	// Теперь удаляем
	deleteReq, _ := http.NewRequest("DELETE", "/api/expenses/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, deleteReq)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}
}

func TestUpdateExpense_Handler(t *testing.T) {
	router, repo := setupTestRouter()

	// Создаём расход напрямую
	expense := &models.Expense{
		Description: "Старое",
		Amount:      100,
		Category:    "Тест",
		Date:        time.Now(),
	}
	repo.Create(context.Background(), expense)

	// Обновляем
	newDesc := "Новое описание"
	update := models.UpdateExpenseRequest{
		Description: &newDesc,
	}

	updateBody, _ := json.Marshal(update)
	updateReq, _ := http.NewRequest("PUT", "/api/expenses/1", bytes.NewBuffer(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, updateReq)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestGetExpense_Success(t *testing.T) {
	router, repo := setupTestRouter()

	// Создаём расход
	expense := &models.Expense{
		Description: "Тестовый",
		Amount:      100,
		Category:    "Тест",
		Date:        time.Now(),
	}
	repo.Create(context.Background(), expense)

	req, _ := http.NewRequest("GET", "/api/expenses/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидали статус 200, получили %d", w.Code)
	}

	var response APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response.Success {
		t.Error("Ответ должен быть успешным")
	}
}
