package academic

import (
	"database/sql"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers the academic year and term routes
func RegisterRoutes(app *fiber.App, db *sql.DB) {
	// Academic Year routes
	app.Get("/api/academic-years", func(c *fiber.Ctx) error { return GetAllAcademicYears(c, db) })
	app.Get("/api/academic-years/:id", func(c *fiber.Ctx) error { return GetAcademicYear(c, db) })
	app.Post("/api/academic-years", func(c *fiber.Ctx) error { return CreateAcademicYear(c, db) })
	app.Put("/api/academic-years/:id", func(c *fiber.Ctx) error { return UpdateAcademicYear(c, db) })
	app.Delete("/api/academic-years/:id", func(c *fiber.Ctx) error { return DeleteAcademicYear(c, db) })
	app.Put("/api/academic-years/:id/set-current", func(c *fiber.Ctx) error { return SetCurrentAcademicYear(c, db) })

	// Term routes
	app.Get("/api/terms", func(c *fiber.Ctx) error { return GetAllTerms(c, db) })
	app.Get("/api/terms/:id", func(c *fiber.Ctx) error { return GetTerm(c, db) })
	app.Post("/api/terms", func(c *fiber.Ctx) error { return CreateTerm(c, db) })
	app.Put("/api/terms/:id", func(c *fiber.Ctx) error { return UpdateTerm(c, db) })
	app.Delete("/api/terms/:id", func(c *fiber.Ctx) error { return DeleteTerm(c, db) })
	app.Put("/api/terms/:id/set-current", func(c *fiber.Ctx) error { return SetCurrentTerm(c, db) })

	// Terms by Academic Year
	app.Get("/api/academic-years/:academicYearId/terms", func(c *fiber.Ctx) error { return GetTermsByAcademicYear(c, db) })

	// Auto-set current based on date
	app.Post("/api/academic/auto-set-current", func(c *fiber.Ctx) error { return AutoSetCurrentByDate(c, db) })

	// Serve the academic settings page
	app.Get("/settings/academic", auth.AuthMiddleware, AcademicSettingsPageHandler(db))
}

// AcademicSettingsPageHandler serves the academic settings page
func AcademicSettingsPageHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from context (set by auth middleware)
		user := c.Locals("user").(*models.User)

		// Get all academic years
		academicYears, err := getAllAcademicYears(db)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to load academic years: "+err.Error())
		}

		// Get all terms
		terms, err := getAllTerms(db)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to load terms: "+err.Error())
		}

		// Create data for template
		data := fiber.Map{
			"AcademicYears": academicYears,
			"Terms":         terms,
			"CurrentPage":   "academic",
			"Title":         "Academic Settings",
			"FirstName":     user.FirstName,
			"LastName":      user.LastName,
			"Email":         user.Email,
			"user":          user,
		}

		// Render template
		return c.Render("academic/settings", data)
	}
}
