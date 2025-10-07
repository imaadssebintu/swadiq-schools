package papers

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetPapersAPI(c *fiber.Ctx) error {
	papers, err := database.GetAllPapers(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch papers"})
	}

	return c.JSON(fiber.Map{
		"papers": papers,
		"count":  len(papers),
	})
}

func GetPaperAPI(c *fiber.Ctx) error {
	paperID := c.Params("id")

	paper, err := database.GetPaperByID(config.GetDB(), paperID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Paper not found"})
	}

	return c.JSON(paper)
}

func CreatePaperAPI(c *fiber.Ctx) error {
	var paper models.Paper
	if err := c.BodyParser(&paper); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if paper.SubjectID == "" || paper.Name == "" || paper.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Subject ID, name, and code are required"})
	}

	if err := database.CreatePaper(config.GetDB(), &paper); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create paper",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Paper created successfully",
		"paper":   paper,
	})
}

func UpdatePaperAPI(c *fiber.Ctx) error {
	paperID := c.Params("id")

	var paper models.Paper
	if err := c.BodyParser(&paper); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	paper.ID = paperID

	if err := database.UpdatePaper(config.GetDB(), &paper); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update paper"})
	}

	return c.JSON(fiber.Map{
		"message": "Paper updated successfully",
		"paper":   paper,
	})
}

func DeletePaperAPI(c *fiber.Ctx) error {
	paperID := c.Params("id")

	if err := database.DeletePaper(config.GetDB(), paperID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete paper"})
	}

	return c.JSON(fiber.Map{"message": "Paper deleted successfully"})
}
