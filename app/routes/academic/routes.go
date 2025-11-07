package academic

import (
	"database/sql"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers the academic year and term routes
func RegisterRoutes(app *fiber.App, db *sql.DB) {
	// Academic Year routes
	app.Get("/api/academic-years", GetAllAcademicYearsHandler(db))
	app.Get("/api/academic-years/:id", GetAcademicYearHandler(db))
	app.Post("/api/academic-years", CreateAcademicYearHandler(db))
	app.Put("/api/academic-years/:id", UpdateAcademicYearHandler(db))
	app.Delete("/api/academic-years/:id", DeleteAcademicYearHandler(db))
	app.Put("/api/academic-years/:id/set-current", SetCurrentAcademicYearHandler(db))

	// Term routes
	app.Get("/api/terms", GetAllTermsHandler(db))
	app.Get("/api/terms/:id", GetTermHandler(db))
	app.Post("/api/terms", CreateTermHandler(db))
	app.Put("/api/terms/:id", UpdateTermHandler(db))
	app.Delete("/api/terms/:id", DeleteTermHandler(db))
	app.Put("/api/terms/:id/set-current", SetCurrentTermHandler(db))

	// Terms by Academic Year
	app.Get("/api/academic-years/:academicYearId/terms", GetTermsByAcademicYearHandler(db))

	// Auto-set current based on date
	app.Post("/api/academic/auto-set-current", AutoSetCurrentByDateHandler(db))

	// Serve the academic settings page
	app.Get("/settings/academic", auth.AuthMiddleware, AcademicSettingsPageHandler(db))
}

// AcademicSettingsPageHandler serves the academic settings page
func AcademicSettingsPageHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from context (set by auth middleware)
		user := c.Locals("user").(*models.User)

		// Get all academic years
		academicYears, err := database.GetAllAcademicYears(db)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to load academic years: "+err.Error())
		}

		// Get all terms
		terms, err := database.GetAllTerms(db)
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

// GetAllAcademicYearsHandler returns all academic years
func GetAllAcademicYearsHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		academicYears, err := database.GetAllAcademicYears(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve academic years"})
		}

		return c.JSON(academicYears)
	}
}

// GetAcademicYearHandler returns a specific academic year by ID
func GetAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		academicYearID := c.Params("id")

		academicYear, err := database.GetAcademicYearByID(db, academicYearID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Academic year not found"})
		}

		return c.JSON(academicYear)
	}
}

// CreateAcademicYearHandler creates a new academic year
func CreateAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var academicYear models.AcademicYear
		if err := c.BodyParser(&academicYear); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
				"body":  string(c.Body()),
			})
		}

		// Validate dates
		if academicYear.EndDate.Time.Before(academicYear.StartDate.Time) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
		}

		if err := database.CreateAcademicYear(db, &academicYear); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create academic year: " + err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(academicYear)
	}
}

// UpdateAcademicYearHandler updates an existing academic year
func UpdateAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		academicYearID := c.Params("id")

		var academicYear models.AcademicYear
		if err := c.BodyParser(&academicYear); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
				"body":  string(c.Body()),
			})
		}

		// Set the ID from the URL
		academicYear.ID = academicYearID

		// Validate dates
		if academicYear.EndDate.Time.Before(academicYear.StartDate.Time) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
		}

		if err := database.UpdateAcademicYear(db, &academicYear); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update academic year: " + err.Error()})
		}

		return c.JSON(academicYear)
	}
}

// DeleteAcademicYearHandler deletes an academic year
func DeleteAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		academicYearID := c.Params("id")

		if err := database.DeleteAcademicYear(db, academicYearID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete academic year"})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetAllTermsHandler returns all terms
func GetAllTermsHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		terms, err := database.GetAllTerms(db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve terms"})
		}

		return c.JSON(terms)
	}
}

// GetTermHandler returns a specific term by ID
func GetTermHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		termID := c.Params("id")

		term, err := database.GetTermByID(db, termID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Term not found"})
		}

		return c.JSON(term)
	}
}

// CreateTermHandler creates a new term
func CreateTermHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var term models.Term
		if err := c.BodyParser(&term); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
				"body":  string(c.Body()),
			})
		}

		// Validate dates
		if term.EndDate.Before(term.StartDate) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
		}

		// Check if the term dates are within the academic year dates
		academicYear, err := database.GetAcademicYearByID(db, term.AcademicYearID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Academic year not found"})
		}

		if term.StartDate.Before(academicYear.StartDate.Time) || term.EndDate.After(academicYear.EndDate.Time) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Term dates must be within the academic year dates"})
		}

		if err := database.CreateTerm(db, &term); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create term: " + err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(term)
	}
}

// UpdateTermHandler updates an existing term
func UpdateTermHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		termID := c.Params("id")

		var term models.Term
		if err := c.BodyParser(&term); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
				"body":  string(c.Body()),
			})
		}

		// Set the ID from the URL
		term.ID = termID

		// Validate dates
		if term.EndDate.Before(term.StartDate) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
		}

		// Check if the term dates are within the academic year dates
		academicYear, err := database.GetAcademicYearByID(db, term.AcademicYearID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Academic year not found"})
		}

		if term.StartDate.Before(academicYear.StartDate.Time) || term.EndDate.After(academicYear.EndDate.Time) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Term dates must be within the academic year dates"})
		}

		if err := database.UpdateTerm(db, &term); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update term: " + err.Error()})
		}

		return c.JSON(term)
	}
}

// DeleteTermHandler deletes a term
func DeleteTermHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		termID := c.Params("id")

		if err := database.DeleteTerm(db, termID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete term"})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetTermsByAcademicYearHandler returns all terms for a specific academic year
func GetTermsByAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		academicYearID := c.Params("academicYearId")

		terms, err := database.GetTermsByAcademicYearID(db, academicYearID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve terms"})
		}

		return c.JSON(terms)
	}
}

// SetCurrentAcademicYearHandler sets an academic year as current
func SetCurrentAcademicYearHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return SetCurrentAcademicYear(c, db)
	}
}

// SetCurrentTermHandler sets a term as current
func SetCurrentTermHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return SetCurrentTerm(c, db)
	}
}

// AutoSetCurrentByDateHandler automatically sets current academic year and term based on current date
func AutoSetCurrentByDateHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return AutoSetCurrentByDate(c, db)
	}
}