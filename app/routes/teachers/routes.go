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

	// API routes
	api := app.Group("/api/teachers")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetTeachersAPI)
	api.Get("/selection", GetTeachersForSelectionAPI) // Fast endpoint for selection
	api.Get("/counts", GetTeacherCountsAPI)
	api.Get("/stats", GetTeacherStatsAPI)
	api.Get("/search", SearchTeachersAPI)
	api.Post("/", CreateTeacherAPI)
	api.Get("/:id", GetTeacherAPI)
	api.Put("/:id", UpdateTeacherAPI)
	api.Delete("/:id", DeleteTeacherAPI)

	subjectsAPI := app.Group("/api/subjects")
	subjectsAPI.Use(auth.AuthMiddleware)
	subjectsAPI.Get("/", GetSubjectsAPI)
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


