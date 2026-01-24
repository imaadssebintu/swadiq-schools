package expenses

import (
	"database/sql"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupExpensesRoutes(app *fiber.App, db *sql.DB) {
	// Initialize database tables
	InitExpensesDB(db)
	// Web Routes
	web := app.Group("/expenses")
	web.Use(auth.AuthMiddleware)
	web.Get("/", ExpensesPageHandler)

	// API Routes
	api := app.Group("/api/expenses")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetExpensesAPI)
	api.Post("/", CreateExpenseAPI)
	api.Put("/:id", UpdateExpenseAPI)
	api.Delete("/:id", DeleteExpenseAPI)

	// Category API
	catAPI := app.Group("/api/expense-categories")
	catAPI.Use(auth.AuthMiddleware)
	catAPI.Get("/", GetCategoriesAPI)
	catAPI.Post("/", CreateCategoryAPI)
	catAPI.Put("/:id", UpdateCategoryAPI)
	catAPI.Delete("/:id", DeleteCategoryAPI)
}
