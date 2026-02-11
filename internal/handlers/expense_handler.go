package handlers

import (
	"net/http"
	"strconv"

	"github.com/dvoryadkinadv/expense-tracker/internal/models"
	"github.com/dvoryadkinadv/expense-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

// ExpenseHandler обрабатывает HTTP-запросы для работы с расходами
type ExpenseHandler struct {
	service *service.ExpenseService
}

// NewExpenseHandler создаёт новый хэндлер
func NewExpenseHandler(s *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: s}
}

// APIResponse - стандартный формат ответа API
// Всегда возвращаем структуру с полями success и data/error
// Так фронтенду проще обрабатывать ответы
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateExpense создаёт новый расход
func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var req models.CreateExpenseRequest

	// Gin сам проверит валидацию по тегам binding
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверные данные: " + err.Error(),
		})
		return
	}

	expense, err := h.service.CreateExpense(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    expense,
	})
}

// GetExpense возвращает расход по ID
func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный ID",
		})
		return
	}

	expense, err := h.service.GetExpense(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    expense,
	})
}

// GetExpenses возвращает список расходов с фильтрацией
func (h *ExpenseHandler) GetExpenses(c *gin.Context) {
	filter := models.ExpenseFilter{
		Category: c.Query("category"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}

	// Парсим limit и offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	expenses, err := h.service.GetExpenses(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    expenses,
	})
}

// UpdateExpense обновляет расход
func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный ID",
		})
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверные данные: " + err.Error(),
		})
		return
	}

	expense, err := h.service.UpdateExpense(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    expense,
	})
}

// DeleteExpense удаляет расход
func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный ID",
		})
		return
	}

	if err := h.service.DeleteExpense(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    "Расход успешно удалён",
	})
}

// GetStats возвращает статистику по расходам
func (h *ExpenseHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
}

// GetCategories возвращает список категорий
func (h *ExpenseHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    categories,
	})
}

// HealthCheck проверяет состояние сервиса
// Полезно для kubernetes liveness/readiness probes
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Сервис работает нормально",
	})
}
