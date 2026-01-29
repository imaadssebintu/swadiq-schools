package settings

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/routes/academic"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupSettingsRoutes(app *fiber.App) {
	settings := app.Group("/settings")
	settings.Use(auth.AuthMiddleware)

	settings.Get("/", SettingsPageHandler())
	settings.Get("/assessment-types", AssessmentTypePageHandler())

	// API Routes for Settings
	db := config.GetDB()
	api := app.Group("/api/settings")
	api.Use(auth.AuthMiddleware)

	// Academic Year routes
	api.Get("/academic-years", func(c *fiber.Ctx) error { return academic.GetAllAcademicYears(c, db) })
	api.Get("/academic-years/:id", func(c *fiber.Ctx) error { return academic.GetAcademicYear(c, db) })
	api.Post("/academic-years", func(c *fiber.Ctx) error { return academic.CreateAcademicYear(c, db) })
	api.Put("/academic-years/:id", func(c *fiber.Ctx) error { return academic.UpdateAcademicYear(c, db) })
	api.Delete("/academic-years/:id", func(c *fiber.Ctx) error { return academic.DeleteAcademicYear(c, db) })
	api.Put("/academic-years/:id/set-current", func(c *fiber.Ctx) error { return academic.SetCurrentAcademicYear(c, db) })

	// Term routes
	api.Get("/terms", func(c *fiber.Ctx) error { return academic.GetAllTerms(c, db) })
	api.Get("/terms/:id", func(c *fiber.Ctx) error { return academic.GetTerm(c, db) })
	api.Post("/terms", func(c *fiber.Ctx) error { return academic.CreateTerm(c, db) })
	api.Put("/terms/:id", func(c *fiber.Ctx) error { return academic.UpdateTerm(c, db) })
	api.Delete("/terms/:id", func(c *fiber.Ctx) error { return academic.DeleteTerm(c, db) })
	api.Put("/terms/:id/set-current", func(c *fiber.Ctx) error { return academic.SetCurrentTerm(c, db) })

	// Terms by Academic Year
	api.Get("/academic-years/:academicYearId/terms", func(c *fiber.Ctx) error { return academic.GetTermsByAcademicYear(c, db) })

	// Auto-set current based on date
	api.Post("/auto-set-current", func(c *fiber.Ctx) error { return academic.AutoSetCurrentByDate(c, db) })

	// Assessment Type routes
	api.Get("/assessment-types", GetAllAssessmentTypes)
	api.Post("/assessment-types", CreateAssessmentType)
	api.Put("/assessment-types/:id", UpdateAssessmentType)
	api.Delete("/assessment-types/:id", DeleteAssessmentType)
}
