package papers

import (
	"log"
	"swadiq-schools/app/config"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetPaperWeightsAPI fetches weights for a specific class, subject, and term
func GetPaperWeightsAPI(c *fiber.Ctx) error {
	classID := c.Query("class_id")
	subjectID := c.Query("subject_id")
	termID := c.Query("term_id")

	if classID == "" || subjectID == "" || termID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "class_id, subject_id, and term_id are required"})
	}

	db := config.GetDB()
	rows, err := db.Query("SELECT id, paper_id, weight FROM paper_weights WHERE class_id = $1 AND subject_id = $2 AND term_id = $3", classID, subjectID, termID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch paper weights: " + err.Error()})
	}
	defer rows.Close()

	var weights []models.PaperWeight
	for rows.Next() {
		var w models.PaperWeight
		// We only populate fields we need for the UI
		if err := rows.Scan(&w.ID, &w.PaperID, &w.Weight); err != nil {
			log.Printf("Error scanning weight: %v", err)
			continue
		}
		w.ClassID = classID
		w.SubjectID = subjectID
		w.TermID = termID
		weights = append(weights, w)
	}

	return c.JSON(weights)
}

// SavePaperWeightsAPI saves or updates paper weights
func SavePaperWeightsAPI(c *fiber.Ctx) error {
	type WeightInput struct {
		PaperID string `json:"paper_id"`
		Weight  int    `json:"weight"`
	}

	type SaveRequest struct {
		ClassID   string        `json:"class_id"`
		SubjectID string        `json:"subject_id"`
		TermID    string        `json:"term_id"`
		Weights   []WeightInput `json:"weights"`
	}

	var req SaveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate total weight sums to 100
	totalWeight := 0
	for _, w := range req.Weights {
		totalWeight += w.Weight
	}

	if totalWeight != 100 {
		return c.Status(400).JSON(fiber.Map{"error": "Total weight must sum up to 100%"})
	}

	db := config.GetDB()
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to start transaction"})
	}

	// Delete existing weights
	_, err = tx.Exec("DELETE FROM paper_weights WHERE class_id = $1 AND subject_id = $2 AND term_id = $3", req.ClassID, req.SubjectID, req.TermID)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear existing weights: " + err.Error()})
	}

	// Insert new weights
	for _, w := range req.Weights {
		_, err := tx.Exec(`
			INSERT INTO paper_weights (class_id, subject_id, paper_id, term_id, weight)
			VALUES ($1, $2, $3, $4, $5)
		`, req.ClassID, req.SubjectID, w.PaperID, req.TermID, w.Weight)

		if err != nil {
			tx.Rollback()
			log.Printf("Error saving weight: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save weights"})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Paper weights saved successfully"})
}
