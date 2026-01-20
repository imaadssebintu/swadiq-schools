package academic

import (
	"database/sql"
	"strings"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetAllAcademicYears returns all academic years
func GetAllAcademicYears(c *fiber.Ctx, db *sql.DB) error {
	academicYears, err := getAllAcademicYears(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve academic years"})
	}

	// Ensure we always return an array, never null
	if academicYears == nil {
		academicYears = []*models.AcademicYear{}
	}

	return c.JSON(academicYears)
}

// GetAcademicYear returns a specific academic year by ID
func GetAcademicYear(c *fiber.Ctx, db *sql.DB) error {
	academicYearID := c.Params("id")

	academicYear, err := getAcademicYearByID(db, academicYearID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Academic year not found"})
	}

	return c.JSON(academicYear)
}

// CreateAcademicYear creates a new academic year
func CreateAcademicYear(c *fiber.Ctx, db *sql.DB) error {
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

	if err := createAcademicYear(db, &academicYear); err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "An academic year with this name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create academic year: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(academicYear)
}

// UpdateAcademicYear updates an existing academic year
func UpdateAcademicYear(c *fiber.Ctx, db *sql.DB) error {
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

	if err := updateAcademicYear(db, &academicYear); err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "An academic year with this name already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update academic year: " + err.Error()})
	}

	return c.JSON(academicYear)
}

// DeleteAcademicYear deletes an academic year
func DeleteAcademicYear(c *fiber.Ctx, db *sql.DB) error {
	academicYearID := c.Params("id")

	if err := deleteAcademicYear(db, academicYearID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete academic year"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetAllTerms returns all terms
func GetAllTerms(c *fiber.Ctx, db *sql.DB) error {
	terms, err := getAllTerms(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve terms"})
	}

	// Ensure we always return an array, never null
	if terms == nil {
		terms = []*models.Term{}
	}

	return c.JSON(terms)
}

// GetTerm returns a specific term by ID
func GetTerm(c *fiber.Ctx, db *sql.DB) error {
	termID := c.Params("id")

	term, err := getTermByID(db, termID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Term not found"})
	}

	return c.JSON(term)
}

// CreateTerm creates a new term
func CreateTerm(c *fiber.Ctx, db *sql.DB) error {
	var term models.Term
	if err := c.BodyParser(&term); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
			"body":  string(c.Body()),
		})
	}

	// Validate dates
	if term.EndDate.Time.Before(term.StartDate.Time) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
	}

	// Check if the term dates are within the academic year dates
	academicYear, err := getAcademicYearByID(db, term.AcademicYearID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Academic year not found"})
	}

	if term.StartDate.Time.Before(academicYear.StartDate.Time) || term.EndDate.Time.After(academicYear.EndDate.Time) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Term dates must be within the academic year dates"})
	}

	if err := createTerm(db, &term); err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "A term with this name already exists in this academic year"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create term: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(term)
}

// UpdateTerm updates an existing term
func UpdateTerm(c *fiber.Ctx, db *sql.DB) error {
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
	if term.EndDate.Time.Before(term.StartDate.Time) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "End date must be after start date"})
	}

	// Check if the term dates are within the academic year dates
	academicYear, err := getAcademicYearByID(db, term.AcademicYearID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Academic year not found"})
	}

	if term.StartDate.Time.Before(academicYear.StartDate.Time) || term.EndDate.Time.After(academicYear.EndDate.Time) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Term dates must be within the academic year dates"})
	}

	if err := updateTerm(db, &term); err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "A term with this name already exists in this academic year"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update term: " + err.Error()})
	}

	return c.JSON(term)
}

// DeleteTerm deletes a term
func DeleteTerm(c *fiber.Ctx, db *sql.DB) error {
	termID := c.Params("id")

	if err := deleteTerm(db, termID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete term"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetTermsByAcademicYear returns all terms for a specific academic year
func GetTermsByAcademicYear(c *fiber.Ctx, db *sql.DB) error {
	academicYearID := c.Params("academicYearId")

	terms, err := getTermsByAcademicYearID(db, academicYearID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve terms"})
	}

	return c.JSON(terms)
}

// SetCurrentAcademicYear sets an academic year as current
func SetCurrentAcademicYear(c *fiber.Ctx, db *sql.DB) error {
	academicYearID := c.Params("id")

	if err := setCurrentAcademicYear(db, academicYearID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set current academic year"})
	}

	return c.JSON(fiber.Map{"message": "Academic year set as current"})
}

// SetCurrentTerm sets a term as current
func SetCurrentTerm(c *fiber.Ctx, db *sql.DB) error {
	termID := c.Params("id")

	if err := setCurrentTerm(db, termID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set current term"})
	}

	return c.JSON(fiber.Map{"message": "Term set as current"})
}

// AutoSetCurrentByDate automatically sets current academic year and term based on current date
func AutoSetCurrentByDate(c *fiber.Ctx, db *sql.DB) error {
	if err := autoSetCurrentAcademicYear(db); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to auto-set current academic year"})
	}

	if err := autoSetCurrentTerm(db); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to auto-set current term"})
	}

	return c.JSON(fiber.Map{"message": "Current academic year and term set automatically"})
}
