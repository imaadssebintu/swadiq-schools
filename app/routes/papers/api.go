package papers

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetPapersBySubjectAPI(c *fiber.Ctx) error {
	subjectID := c.Params("subjectId")

	if subjectID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Subject ID is required"})
	}

	papers, err := database.GetPapersBySubject(config.GetDB(), subjectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to fetch papers",
			"details":    err.Error(),
			"subject_id": subjectID,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"papers":  papers,
		"count":   len(papers),
	})
}

func GetPapersAPI(c *fiber.Ctx) error {
	classID := c.Query("class_id")
	subjectID := c.Query("subject_id")

	var papers []*models.Paper
	var err error

	if classID != "" && subjectID != "" {
		// Get papers for a specific class and subject
		papers, err = database.GetPapersByClassAndSubject(config.GetDB(), classID, subjectID)
	} else if subjectID != "" {
		// Get papers for a specific subject
		papers, err = database.GetPapersBySubject(config.GetDB(), subjectID)
	} else if classID != "" {
		// Get papers for a specific class
		papers, err = database.GetPapersByClass(config.GetDB(), classID)
	} else {
		// Get all papers
		papers, err = database.GetAllPapers(config.GetDB())
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch papers"})
	}

	return c.JSON(papers)
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

	if paper.SubjectID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Subject ID is required"})
	}

	if paper.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Paper name is required"})
	}

	// Always generate paper code based on subject code, ignoring any provided code
	// Get the subject to get its code
	subject, err := database.GetSubjectByID(config.GetDB(), paper.SubjectID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid subject ID"})
	}

	// Get existing papers for this subject to determine the next paper number
	existingPapers, err := database.GetPapersBySubject(config.GetDB(), paper.SubjectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch existing papers for subject"})
	}

	// Use first 3 characters of subject code
	subjectPrefix := subject.Code
	if len(subject.Code) > 3 {
		subjectPrefix = subject.Code[:3]
	}

	// Determine the next paper number
	nextPaperNumber := 1
	for _, existingPaper := range existingPapers {
		// Extract the number from the paper code (e.g., "ENG-2" -> 2)
		if len(existingPaper.Code) > len(subjectPrefix)+1 {
			// Check if the code starts with the subject prefix followed by a dash
			if existingPaper.Code[:len(subjectPrefix)] == subjectPrefix && existingPaper.Code[len(subjectPrefix)] == '-' {
				// Try to parse the number after the dash
				var num int
				_, err := fmt.Sscanf(existingPaper.Code[len(subjectPrefix)+1:], "%d", &num)
				if err == nil && num >= nextPaperNumber {
					nextPaperNumber = num + 1
				}
			}
		}
	}

	paper.Code = fmt.Sprintf("%s-%d", subjectPrefix, nextPaperNumber)

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

	var updatedPaper models.Paper
	if err := c.BodyParser(&updatedPaper); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get the existing paper to preserve the auto-generated code
	existingPaper, err := database.GetPaperByID(config.GetDB(), paperID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Paper not found"})
	}

	// Update only the fields that can be changed, preserving the auto-generated code
	paperToUpdate := &models.Paper{
		ID:           paperID,
		SubjectID:    updatedPaper.SubjectID,
		Name:         updatedPaper.Name,
		Code:         existingPaper.Code, // Preserve the auto-generated code
		IsCompulsory: updatedPaper.IsCompulsory,
		IsActive:     updatedPaper.IsActive,
	}

	if err := database.UpdatePaper(config.GetDB(), paperToUpdate); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update paper"})
	}

	return c.JSON(fiber.Map{
		"message": "Paper updated successfully",
		"paper":   paperToUpdate,
	})
}

func DeletePaperAPI(c *fiber.Ctx) error {
	paperID := c.Params("id")

	if err := database.DeletePaper(config.GetDB(), paperID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete paper"})
	}

	return c.JSON(fiber.Map{"message": "Paper deleted successfully"})
}

// GetPapersTableAPI returns papers data for table display
func GetPapersTableAPI(c *fiber.Ctx) error {
	papers, err := database.GetAllPapers(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch papers"})
	}

	return c.JSON(fiber.Map{
		"papers": papers,
		"count":  len(papers),
	})
}

// GetPapersStatsAPI returns statistics for papers page
func GetPapersStatsAPI(c *fiber.Ctx) error {
	stats, err := database.GetPapersStats(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch stats"})
	}

	return c.JSON(stats)
}

// AssignTeacherToClassPaperAPI assigns a teacher to a class paper
func AssignTeacherToClassPaperAPI(c *fiber.Ctx) error {
	var req struct {
		ClassID   string  `json:"class_id"`
		PaperID   string  `json:"paper_id"`
		TeacherID *string `json:"teacher_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	db := config.GetDB()

	// Check if class paper already exists
	var existingID string
	checkQuery := `SELECT id FROM class_papers WHERE class_id = $1 AND paper_id = $2 AND deleted_at IS NULL`
	err := db.QueryRow(checkQuery, req.ClassID, req.PaperID).Scan(&existingID)

	if err != nil {
		// Create new class paper
		query := `INSERT INTO class_papers (class_id, paper_id, teacher_id, created_at, updated_at)
				  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

		var classPaperID string
		err = db.QueryRow(query, req.ClassID, req.PaperID, req.TeacherID).Scan(&classPaperID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create class paper"})
		}

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"message": "Teacher assigned to class paper successfully",
			"id":      classPaperID,
		})
	}

	// Update existing class paper
	updateQuery := `UPDATE class_papers SET teacher_id = $1, updated_at = NOW() WHERE id = $2`
	_, err = db.Exec(updateQuery, req.TeacherID, existingID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update class paper"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Teacher assignment updated successfully",
		"id":      existingID,
	})
}
