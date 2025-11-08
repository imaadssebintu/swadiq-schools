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
	departments.Get("/teachers", DepartmentTeachersPage)

	// API routes
	api := app.Group("/api/departments")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetDepartmentsAPI)
	api.Get("/:id/teachers", GetDepartmentTeachersAPI)
	api.Post("/:id/teachers", AddTeacherToDepartmentAPI)
	api.Put("/:id/leadership", SetDepartmentLeadershipAPI)
	api.Delete("/:id/teachers/:teacherId", RemoveTeacherFromDepartmentAPI)
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

func DepartmentTeachersPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("departments/department_teachers", fiber.Map{
		"Title":       "Department Teachers - Swadiq Schools",
		"CurrentPage": "departments",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
