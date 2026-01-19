package departments

import (
	"swadiq-schools/app/config"
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
	api.Get("/stats", GetDepartmentStatsAPI)
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
	db := config.GetDB()

	// Fetch stats for the dashboard cards
	var totalDepts, activeDepts, totalTeachers, totalSubjects int
	db.QueryRow("SELECT COUNT(*) FROM departments WHERE deleted_at IS NULL").Scan(&totalDepts)
	db.QueryRow("SELECT COUNT(*) FROM departments WHERE is_active = true AND deleted_at IS NULL").Scan(&activeDepts)
	db.QueryRow("SELECT COUNT(DISTINCT ud.user_id) FROM user_departments ud JOIN users u ON ud.user_id = u.id WHERE u.is_active = true").Scan(&totalTeachers)
	db.QueryRow("SELECT COUNT(*) FROM subjects WHERE deleted_at IS NULL").Scan(&totalSubjects)

	return c.Render("departments/index", fiber.Map{
		"Title":             "Departments Management - Swadiq Schools",
		"CurrentPage":       "departments",
		"user":              user,
		"FirstName":         user.FirstName,
		"LastName":          user.LastName,
		"Email":             user.Email,
		"TotalDepartments":  totalDepts,
		"ActiveDepartments": activeDepts,
		"TotalTeachers":     totalTeachers,
		"TotalSubjects":     totalSubjects,
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
