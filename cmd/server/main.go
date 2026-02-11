package main

import (
	"log"
	"os"

	"github.com/dvoryadkinadv/expense-tracker/internal/database"
	"github.com/dvoryadkinadv/expense-tracker/internal/handlers"
	"github.com/dvoryadkinadv/expense-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Читаем настройки из переменных окружения
	// В продакшене лучше использовать что-то типа viper,
	// но для простоты пока так
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "expense_tracker"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Подключаемся к базе
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()
	log.Println("Подключение к БД установлено")

	// Создаём слои приложения
	repo := database.NewExpenseRepository(db)
	expenseService := service.NewExpenseService(repo)
	expenseHandler := handlers.NewExpenseHandler(expenseService)

	// Настраиваем роутер
	router := setupRouter(expenseHandler)

	// Запускаем сервер
	port := getEnv("PORT", "8080")
	log.Printf("Сервер запускается на порту %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// setupRouter настраивает все маршруты
func setupRouter(h *handlers.ExpenseHandler) *gin.Engine {
	// В продакшене можно использовать gin.ReleaseMode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// Middleware для CORS (если будет фронтенд)
	router.Use(corsMiddleware())

	// Health check - для мониторинга
	router.GET("/health", handlers.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Расходы
		expenses := api.Group("/expenses")
		{
			expenses.POST("", h.CreateExpense)
			expenses.GET("", h.GetExpenses)
			expenses.GET("/:id", h.GetExpense)
			expenses.PUT("/:id", h.UpdateExpense)
			expenses.DELETE("/:id", h.DeleteExpense)
		}

		// Статистика
		api.GET("/stats", h.GetStats)

		// Категории
		api.GET("/categories", h.GetCategories)
	}

	return router
}

// corsMiddleware добавляет заголовки CORS
// Без этого браузер не даст фронтенду делать запросы
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getEnv возвращает значение переменной окружения или дефолт
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
