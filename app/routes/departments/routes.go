package departments

import (
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
	api.Get("/overview", GetDepartmentOverviewAPI)
	api.Post("/", CreateDepartmentAPI)
	api.Put("/:id", UpdateDepartmentAPI)
	api.Delete("/:id", DeleteDepartmentAPI)
}

func DepartmentsPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("departments/index", fiber.Map{
		"Title":       "Departments Management - Swadiq Schools",
		"CurrentPage": "departments",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
