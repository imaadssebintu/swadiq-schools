package results

import (
	"database/sql"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupResultsRoutes sets up all results-related routes
func SetupResultsRoutes(app *fiber.App, db *sql.DB) {
	// API routes
	api := app.Group("/api/results")
	api.Use(auth.AuthMiddleware)
	api.Get("/", func(c *fiber.Ctx) error { return GetResultsByExam(c, db) })
	api.Post("/batch", func(c *fiber.Ctx) error { return BatchSaveResults(c, db) })
	api.Put("/:id", func(c *fiber.Ctx) error { return UpdateSingleResult(c, db) })
	api.Get("/student/:id", func(c *fiber.Ctx) error { return GetStudentResults(c, db) })
	api.Delete("/:id", func(c *fiber.Ctx) error { return DeleteSingleResult(c, db) })

	// Exam-specific API route
	examAPI := app.Group("/api/assessments")
	examAPI.Use(auth.AuthMiddleware)
	examAPI.Get("/:id/students-with-results", func(c *fiber.Ctx) error { return GetStudentsWithResults(c, db) })

	// Page route for results entry
	app.Get("/assessments/:id/results", auth.AuthMiddleware, func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		examID := c.Params("id")

		c.Locals("Title", "Enter Results")
		return c.Render("exams/results", fiber.Map{
			"Title":       "Enter Results",
			"CurrentPage": "exams",
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
			"user":        user,
			"examID":      examID,
		})
	})

	// Grades API routes
	gradesAPI := app.Group("/api/settings/grades")
	gradesAPI.Use(auth.AuthMiddleware)
	gradesAPI.Get("/", func(c *fiber.Ctx) error { return GetGradesAPI(c, db) })
	gradesAPI.Post("/", func(c *fiber.Ctx) error { return CreateGradeAPI(c, db) })
	gradesAPI.Put("/:id", func(c *fiber.Ctx) error { return UpdateGradeAPI(c, db) })
	gradesAPI.Delete("/:id", func(c *fiber.Ctx) error { return DeleteGradeAPI(c, db) })

	// Grades management page
	app.Get("/settings/grades", auth.AuthMiddleware, func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		c.Locals("Title", "Manage Grades")
		return c.Render("results/grades", fiber.Map{
			"Title":       "Manage Grades",
			"CurrentPage": "settings",
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
			"user":        user,
		})
	})
}
