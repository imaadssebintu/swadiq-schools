package subjects

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupSubjectsRoutes(app *fiber.App) {
	subjects := app.Group("/subjects")
	subjects.Use(auth.AuthMiddleware)

	// Routes
	subjects.Get("/", SubjectsPage)

	// API routes
	api := app.Group("/api/subjects")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetSubjectsAPI)
	api.Post("/", CreateSubjectAPI)
}

func SubjectsPage(c *fiber.Ctx) error {
	subjects, err := database.GetAllSubjects(config.GetDB())
	if err != nil {
		// Log the error for debugging
		println("Error getting subjects:", err.Error())
		// Initialize empty slice if there's an error
		subjects = []*models.Subject{}
	}

	// Ensure subjects is never nil
	if subjects == nil {
		subjects = []*models.Subject{}
	}

	user := c.Locals("user").(*models.User)
	return c.Render("subjects/index", fiber.Map{
		"Title":       "Subjects Management - Swadiq Schools",
		"CurrentPage": "subjects",
		"subjects":    subjects,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
