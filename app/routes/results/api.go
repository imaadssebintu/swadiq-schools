package results

import (
	"database/sql"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetResultsByExam returns all results for a specific exam
func GetResultsByExam(c *fiber.Ctx, db *sql.DB) error {
	examID := c.Query("exam_id")
	if examID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exam_id is required",
		})
	}

	results, err := GetResultsByExamID(db, examID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch results",
		})
	}

	return c.JSON(results)
}

// GetStudentsWithResults returns all students in a class with their results for an exam
func GetStudentsWithResults(c *fiber.Ctx, db *sql.DB) error {
	examID := c.Params("id")
	if examID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exam_id is required",
		})
	}

	// Get exam details to find the class
	exam, err := getExamByID(db, examID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Exam not found",
		})
	}

	studentsWithResults, err := GetStudentsWithResultsByExam(db, examID, exam.ClassID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch students with results",
		})
	}

	return c.JSON(fiber.Map{
		"exam":     exam,
		"students": studentsWithResults,
	})
}

// BatchSaveResults handles batch create/update of results
func BatchSaveResults(c *fiber.Ctx, db *sql.DB) error {
	var request struct {
		ExamID  string  `json:"exam_id"`
		PaperID string  `json:"paper_id"`
		TermID  *string `json:"term_id,omitempty"`
		Results []struct {
			StudentID string  `json:"student_id"`
			Marks     float64 `json:"marks"`
			GradeID   *string `json:"grade_id,omitempty"`
		} `json:"results"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if request.ExamID == "" || request.PaperID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exam_id and paper_id are required",
		})
	}

	// Debug: Log term_id
	if request.TermID != nil {
		println("Term ID received:", *request.TermID)
	} else {
		println("Term ID is nil")
	}

	// Convert request to models.Result slice
	var results []*models.Result
	for _, r := range request.Results {
		// Validate marks
		if r.Marks < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Marks cannot be negative",
			})
		}

		if r.Marks > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Marks cannot be above 100",
			})
		}

		result := &models.Result{
			ExamID:    request.ExamID,
			StudentID: r.StudentID,
			PaperID:   request.PaperID,
			TermID:    request.TermID,
			Marks:     r.Marks,
			GradeID:   r.GradeID,
		}
		results = append(results, result)
	}

	// Batch save
	if err := BatchCreateOrUpdateResults(db, results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save results",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Results saved successfully",
		"count":   len(results),
	})
}

// UpdateSingleResult updates a single result
func UpdateSingleResult(c *fiber.Ctx, db *sql.DB) error {
	resultID := c.Params("id")

	var request struct {
		Marks   float64 `json:"marks"`
		GradeID *string `json:"grade_id,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate marks
	if request.Marks < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Marks cannot be negative",
		})
	}

	if request.Marks > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Marks cannot be above 100",
		})
	}

	result := &models.Result{
		ID:      resultID,
		Marks:   request.Marks,
		GradeID: request.GradeID,
	}

	if err := UpdateResult(db, result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update result",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Result updated successfully",
		"result":  result,
	})
}

// DeleteSingleResult deletes a result
func DeleteSingleResult(c *fiber.Ctx, db *sql.DB) error {
	resultID := c.Params("id")

	if err := DeleteResult(db, resultID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete result",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Result deleted successfully",
	})
}

// Helper function to get exam by ID
func getExamByID(db *sql.DB, examID string) (*models.Exam, error) {
	query := `
		SELECT 
			e.id, e.name, e.class_id, e.academic_year_id, e.term_id, 
			e.paper_id, e.type, e.start_time, e.end_time, e.is_active,
			e.created_at, e.updated_at,
			c.id, c.name, c.code,
			p.id, p.name, p.code,
			s.id, s.name, s.code
		FROM exams e
		LEFT JOIN classes c ON e.class_id = c.id
		LEFT JOIN papers p ON e.paper_id = p.id
		LEFT JOIN subjects s ON p.subject_id = s.id
		WHERE e.id = $1 AND e.deleted_at IS NULL
	`

	var exam models.Exam
	var class models.Class
	var paper models.Paper
	var subject models.Subject
	var academicYearID, termID sql.NullString
	var examType sql.NullString

	err := db.QueryRow(query, examID).Scan(
		&exam.ID, &exam.Name, &exam.ClassID, &academicYearID, &termID,
		&exam.PaperID, &examType, &exam.StartTime, &exam.EndTime, &exam.IsActive,
		&exam.CreatedAt, &exam.UpdatedAt,
		&class.ID, &class.Name, &class.Code,
		&paper.ID, &paper.Name, &paper.Code,
		&subject.ID, &subject.Name, &subject.Code,
	)

	if err != nil {
		return nil, err
	}

	// Always set these fields, even if null
	if academicYearID.Valid {
		exam.AcademicYearID = &academicYearID.String
	} else {
		exam.AcademicYearID = nil
	}

	if termID.Valid {
		exam.TermID = &termID.String
	} else {
		exam.TermID = nil
	}

	if examType.Valid {
		exam.Type = examType.String
	}

	exam.Class = &class
	exam.Paper = &paper
	paper.Subject = &subject
	exam.Paper = &paper

	return &exam, nil
}

// GetStudentResults returns all results for a specific student
func GetStudentResults(c *fiber.Ctx, db *sql.DB) error {
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "student_id is required",
		})
	}

	results, err := GetStudentAssessmentHistory(db, studentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch student assessment history",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"results": results,
	})
}

// GetGradesAPI returns all grades
func GetGradesAPI(c *fiber.Ctx, db *sql.DB) error {
	grades, err := GetAllGrades(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch grades",
		})
	}

	return c.JSON(grades)
}

// CreateGradeAPI handles the creation of a new grade
func CreateGradeAPI(c *fiber.Ctx, db *sql.DB) error {
	var g models.Grade
	if err := c.BodyParser(&g); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := CreateGrade(db, &g); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create grade",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(g)
}

// UpdateGradeAPI handles updating an existing grade
func UpdateGradeAPI(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")
	var g models.Grade
	if err := c.BodyParser(&g); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	g.ID = id
	if err := UpdateGrade(db, &g); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update grade",
		})
	}

	return c.JSON(g)
}

// DeleteGradeAPI handles the deletion of a grade
func DeleteGradeAPI(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")
	if err := DeleteGrade(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete grade",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
