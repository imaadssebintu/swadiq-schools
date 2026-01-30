package exams

import (
	"database/sql"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupExamRoutes sets up all exam-related routes
func SetupExamRoutes(app *fiber.App, db *sql.DB) {
	// API routes
	api := app.Group("/api/assessments")
	api.Get("/", func(c *fiber.Ctx) error { return GetAllExams(c, db) })
	api.Get("/:id", func(c *fiber.Ctx) error { return GetExam(c, db) })
	api.Post("/", func(c *fiber.Ctx) error { return CreateExam(c, db) })
	api.Put("/:id", func(c *fiber.Ctx) error { return UpdateExam(c, db) })
	api.Delete("/:id", func(c *fiber.Ctx) error { return DeleteExam(c, db) })

	// Page routes
	app.Get("/assessments", auth.AuthMiddleware, func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		return c.Render("exams/hub", fiber.Map{
			"Title":       "Assessments Hub",
			"CurrentPage": "exams",
			"user":        user,
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
		})
	})

	app.Get("/assessments/:class_id", auth.AuthMiddleware, func(c *fiber.Ctx) error {
		classID := c.Params("class_id")
		user := c.Locals("user").(*models.User)
		return c.Render("exams/index", fiber.Map{
			"Title":       "Manage Assessments",
			"CurrentPage": "exams",
			"classID":     classID,
			"user":        user,
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
		})
	})
}
