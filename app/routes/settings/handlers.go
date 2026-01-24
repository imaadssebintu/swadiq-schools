package settings

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/academic"

	"github.com/gofiber/fiber/v2"
)

func SettingsPageHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		// Get all academic years
		academicYears, err := academic.GetAcademicYearsForTemplate(config.GetDB())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to load academic years: "+err.Error())
		}

		// Get all terms
		terms, err := academic.GetTermsForTemplate(config.GetDB())
		if err != nil {
			terms = []*models.Term{}
		}

		// Get all assessment types
		assessmentTypes, err := academic.GetAllAssessmentTypes(config.GetDB())
		if err != nil {
			assessmentTypes = []*models.AssessmentType{}
		}

		return c.Render("settings/index", fiber.Map{
			"Title":           "App Settings",
			"CurrentPage":     "settings",
			"FirstName":       user.FirstName,
			"LastName":        user.LastName,
			"Email":           user.Email,
			"user":            user,
			"AcademicYears":   academicYears,
			"Terms":           terms,
			"AssessmentTypes": assessmentTypes,
		})
	}
}

// Assessment Type Handlers
func GetAllAssessmentTypes(c *fiber.Ctx) error {
	types, err := academic.GetAllAssessmentTypes(config.GetDB())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load assessment types"})
	}
	return c.JSON(types)
}

func CreateAssessmentType(c *fiber.Ctx) error {
	var t models.AssessmentType
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := academic.CreateAssessmentType(config.GetDB(), &t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create assessment type"})
	}

	return c.Status(fiber.StatusCreated).JSON(t)
}

func UpdateAssessmentType(c *fiber.Ctx) error {
	id := c.Params("id")
	var t models.AssessmentType
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	t.ID = id
	if err := academic.UpdateAssessmentType(config.GetDB(), &t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update assessment type"})
	}

	return c.JSON(t)
}

func DeleteAssessmentType(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := academic.DeleteAssessmentType(config.GetDB(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete assessment type"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
