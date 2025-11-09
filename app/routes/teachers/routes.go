package teachers

import (
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupTeachersRoutes(app *fiber.App) {
	teachers := app.Group("/teachers")
	teachers.Use(auth.AuthMiddleware)

	// Routes
	teachers.Get("/", TeachersPage)
	teachers.Get("/:id", TeacherViewPage)

	// API routes
	api := app.Group("/api/teachers")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetTeachersAPI)
	api.Get("/selection", GetTeachersForSelectionAPI) // Fast endpoint for selection
	api.Get("/counts", GetTeacherCountsAPI)
	api.Get("/stats", GetTeacherStatsAPI)
	api.Get("/search", SearchTeachersAPI)
	api.Get("/department-overview", GetDepartmentOverviewAPI)
	api.Post("/", CreateTeacherAPI)
	api.Get("/:id", GetTeacherAPI)
	api.Put("/:id", UpdateTeacherAPI)
	api.Delete("/:id", DeleteTeacherAPI)
	api.Get("/:id/subjects", GetTeacherSubjectsAPI)
	api.Post("/:id/subjects", AssignTeacherSubjectsAPI)
}

func TeachersPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("teachers/index", fiber.Map{
		"Title":       "Teachers - Swadiq Schools",
		"CurrentPage": "teachers",
		"teachers":    []*models.User{}, // Empty array
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}

func TeacherViewPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	teacherID := c.Params("id")
	
	return c.Render("teachers/view", fiber.Map{
		"Title":       "Teacher Details - Swadiq Schools",
		"CurrentPage": "teachers",
		"teacherID":   teacherID,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}


