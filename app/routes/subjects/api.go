package subjects

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func SearchSubjectsAPI(c *fiber.Ctx) error {
	query := c.Query("q", "")
	
	var subjects []*models.Subject
	var err error
	
	if query == "" {
		subjects, err = database.GetAllSubjects(config.GetDB())
	} else {
		subjects, err = database.SearchSubjects(config.GetDB(), query)
	}
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to search subjects"})
	}

	return c.JSON(fiber.Map{
		"subjects": subjects,
		"count":    len(subjects),
	})
}

func GetSubjectsAPI(c *fiber.Ctx) error {
	departmentID := c.Query("department_id")
	
	var subjects []*models.Subject
	var err error
	
	if departmentID != "" {
		subjects, err = database.GetSubjectsByDepartment(config.GetDB(), departmentID)
	} else {
		subjects, err = database.GetAllSubjects(config.GetDB())
	}
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch subjects"})
	}

	return c.JSON(fiber.Map{
		"subjects": subjects,
		"count":    len(subjects),
	})
}

func GetSubjectAPI(c *fiber.Ctx) error {
	subjectID := c.Params("id")
	
	subject, err := database.GetSubjectByID(config.GetDB(), subjectID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Subject not found"})
	}

	return c.JSON(subject)
}

func CreateSubjectAPI(c *fiber.Ctx) error {
	var subject models.Subject
	if err := c.BodyParser(&subject); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if subject.Name == "" || subject.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and code are required"})
	}

	if err := database.CreateSubject(config.GetDB(), &subject); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create subject",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Subject created successfully",
		"subject": subject,
	})
}

func UpdateSubjectAPI(c *fiber.Ctx) error {
	subjectID := c.Params("id")
	
	var subject models.Subject
	if err := c.BodyParser(&subject); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	subject.ID = subjectID

	if err := database.UpdateSubject(config.GetDB(), &subject); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update subject"})
	}

	return c.JSON(fiber.Map{
		"message": "Subject updated successfully",
		"subject": subject,
	})
}

func DeleteSubjectAPI(c *fiber.Ctx) error {
	subjectID := c.Params("id")

	if err := database.DeleteSubject(config.GetDB(), subjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete subject"})
	}

	return c.JSON(fiber.Map{"message": "Subject deleted successfully"})
}

func GetSubjectsWithPapersAPI(c *fiber.Ctx) error {
	db := config.GetDB()

	query := `
		SELECT 
			s.id as subject_id, 
			s.name as subject_name, 
			s.code as subject_code, 
			s.department_id,
			p.id as paper_id,
			p.name as paper_name,
			p.code as paper_code
		FROM 
			subjects s
		LEFT JOIN 
			papers p ON s.id = p.subject_id AND p.deleted_at IS NULL
		WHERE 
			s.deleted_at IS NULL
		ORDER BY 
			s.name, p.name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch subjects with papers"})
	}
	defer rows.Close()

	subjectsMap := make(map[string]fiber.Map)
	var subjectsOrder []string

	for rows.Next() {
		var subjectID, subjectName, subjectCode, departmentID string
		var paperID, paperName, paperCode *string // Use pointers for nullable fields from LEFT JOIN

		if err := rows.Scan(&subjectID, &subjectName, &subjectCode, &departmentID, &paperID, &paperName, &paperCode); err != nil {
			continue
		}

		if _, ok := subjectsMap[subjectID]; !ok {
			subjectsMap[subjectID] = fiber.Map{
				"id":            subjectID,
				"name":          subjectName,
				"code":          subjectCode,
				"department_id": departmentID,
				"papers":        []fiber.Map{},
			}
			subjectsOrder = append(subjectsOrder, subjectID)
		}

		if paperID != nil {
			papers := subjectsMap[subjectID]["papers"].([]fiber.Map)
			subjectsMap[subjectID]["papers"] = append(papers, fiber.Map{
				"id":   *paperID,
				"name": *paperName,
				"code": *paperCode,
			})
		}
	}

	// Create the final subjects slice in order
	subjects := make([]fiber.Map, len(subjectsOrder))
	for i, subjectID := range subjectsOrder {
		subjects[i] = subjectsMap[subjectID]
	}

	return c.JSON(fiber.Map{
		"subjects": subjects,
		"count":    len(subjects),
	})
}
