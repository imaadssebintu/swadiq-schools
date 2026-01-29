package results

import (
	"database/sql"
	"fmt"
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

// ApiGetClassResultsMatrix returns all data for the grid entry view
func ApiGetClassResultsMatrix(c *fiber.Ctx, db *sql.DB) error {
	classID := c.Query("class_id")
	termID := c.Query("term_id")
	assessmentTypeID := c.Query("assessment_type_id")

	if classID == "" || termID == "" || assessmentTypeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "class_id, term_id, and assessment_type_id are required",
		})
	}

	matrix, err := GetClassResultsMatrix(db, classID, termID, assessmentTypeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch results matrix: " + err.Error(),
		})
	}

	return c.JSON(matrix)
}

// ApiBatchSaveGridResults handles high-frequency saving from the grid view
func ApiBatchSaveGridResults(c *fiber.Ctx, db *sql.DB) error {
	var request struct {
		ClassID          string `json:"class_id"`
		TermID           string `json:"term_id"`
		AcademicYearID   string `json:"academic_year_id"`
		AssessmentTypeID string `json:"assessment_type_id"`
		Results          []struct {
			StudentID string  `json:"student_id"`
			PaperID   string  `json:"paper_id"`
			Marks     float64 `json:"marks"`
		} `json:"results"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Get or Create Exams for each paper involved
	examMap := make(map[string]string)

	rows, err := db.Query(`
		SELECT id, paper_id FROM exams 
		WHERE class_id = $1 AND term_id = $2 AND assessment_type_id = $3 AND deleted_at IS NULL`,
		request.ClassID, request.TermID, request.AssessmentTypeID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, pID string
			if err := rows.Scan(&id, &pID); err == nil {
				examMap[pID] = id
			}
		}
	}

	var finalResults []*models.Result

	for _, r := range request.Results {
		examID, exists := examMap[r.PaperID]
		if !exists {
			// Create new exam record for this paper
			var paperName, subjectName string
			db.QueryRow(`
				SELECT p.name, s.name FROM papers p 
				JOIN subjects s ON p.subject_id = s.id 
				WHERE p.id = $1`, r.PaperID).Scan(&paperName, &subjectName)

			examName := fmt.Sprintf("%s - %s", subjectName, paperName)

			// Insert exam directly to avoid complex imports for now
			err := db.QueryRow(`
				INSERT INTO exams (name, class_id, academic_year_id, term_id, paper_id, assessment_type_id, type, start_time, end_time, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW() + INTERVAL '2 hours', true, NOW(), NOW())
				RETURNING id`,
				examName, request.ClassID, request.AcademicYearID, request.TermID, r.PaperID, request.AssessmentTypeID, "exam").Scan(&examID)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create exam record: " + err.Error()})
			}
			examMap[r.PaperID] = examID
		}

		finalResults = append(finalResults, &models.Result{
			ExamID:    examID,
			StudentID: r.StudentID,
			PaperID:   r.PaperID,
			TermID:    &request.TermID,
			Marks:     r.Marks,
		})
	}

	// 2. Batch save results
	if err := BatchCreateOrUpdateResults(db, finalResults); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save results: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Grid results saved successfully",
		"count":   len(finalResults),
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
