package departments

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupDepartmentsRoutes(app *fiber.App) {
	departments := app.Group("/departments")
	departments.Use(auth.AuthMiddleware)

	// Routes
	departments.Get("/", DepartmentsPage)

	// API routes
	api := app.Group("/api/departments")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetDepartmentsAPI)
	api.Post("/", CreateDepartmentAPI)
	api.Put("/:id", UpdateDepartmentAPI)
	api.Delete("/:id", DeleteDepartmentAPI)
}

func DepartmentsPage(c *fiber.Ctx) error {
	departments, err := database.GetAllDepartments(config.GetDB())
	if err != nil {
		// Log the error for debugging
		println("Error getting departments:", err.Error())
		// Initialize empty slice if there's an error
		departments = []*models.Department{}
	}

	// Ensure departments is never nil
	if departments == nil {
		departments = []*models.Department{}
	}

	user := c.Locals("user").(*models.User)
	return c.Render("departments/index", fiber.Map{
		"Title":       "Departments Management - Swadiq Schools",
		"CurrentPage": "departments",
		"departments": departments,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
