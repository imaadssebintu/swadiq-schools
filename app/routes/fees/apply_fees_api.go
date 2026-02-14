package fees

import (
	"database/sql"
	"log"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

type ApplyFeeRequest struct {
	FeeTypeID      string            `json:"fee_type_id"`
	Amount         float64           `json:"amount"`
	DueDate        models.CustomTime `json:"due_date"`
	AcademicYearID string            `json:"academic_year_id"`
	TermID         string            `json:"term_id"`
}

// ApplyFeesAPI applies fees based on fee type scope
func ApplyFeesAPI(c *fiber.Ctx, db *sql.DB) error {
	var req ApplyFeeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Get fee type details
	var feeType struct {
		Name            string
		Scope           string
		TargetClassID   *string
		TargetStudentID *string
	}

	log.Printf("Applying fee type with ID: %s", req.FeeTypeID)

	err := db.QueryRow(`
		SELECT name, COALESCE(scope, 'manual')
		FROM fee_types WHERE id = $1 AND is_active = true
	`, req.FeeTypeID).Scan(&feeType.Name, &feeType.Scope)

	if err != nil {
		log.Printf("Fee type lookup failed: %v", err)
		return fiber.NewError(fiber.StatusNotFound, "Fee type not found")
	}

	log.Printf("Fee Type Scope: %s", feeType.Scope)

	var studentIDs []string

	// Get student IDs based on scope
	switch feeType.Scope {
	case "all_students":
		rows, err := db.Query("SELECT id FROM students WHERE deleted_at IS NULL")
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch students")
		}
		defer rows.Close()

		for rows.Next() {
			var studentID string
			rows.Scan(&studentID)
			studentIDs = append(studentIDs, studentID)
		}

	case "class":
		// Get class IDs from assignments
		rows, err := db.Query("SELECT class_id FROM fee_type_assignments WHERE fee_type_id = $1", req.FeeTypeID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch fee assignments")
		}
		defer rows.Close()

		var classIDs []string
		for rows.Next() {
			var classID string
			rows.Scan(&classID)
			classIDs = append(classIDs, classID)
		}
		rows.Close() // Close before reusing

		log.Printf("Found %d assigned classes for fee type", len(classIDs))

		if len(classIDs) == 0 {
			// If no assignments found, look for ALL classes? No, that's unsafe.
			// But maybe the user intends to select a class manually?
			// For now, return specific error
			return fiber.NewError(fiber.StatusBadRequest, "No classes assigned to this fee type. Edit the fee type to assign classes.")
		}

		// Get students for these classes
		for _, classID := range classIDs {
			// Query students table directly as it has class_id
			classRow, err := db.Query("SELECT id FROM students WHERE class_id = $1 AND deleted_at IS NULL AND is_active = true", classID)
			if err != nil {
				log.Printf("Error fetching students for class %s: %v", classID, err)
				continue
			}
			defer classRow.Close()
			for classRow.Next() {
				var studentID string
				classRow.Scan(&studentID)
				studentIDs = append(studentIDs, studentID)
			}
			classRow.Close()
		}
		log.Printf("Found %d students in assigned classes", len(studentIDs))

	case "all_classes":
		// Query all active students assigned to any class
		rows, err := db.Query("SELECT id FROM students WHERE class_id IS NOT NULL AND deleted_at IS NULL AND is_active = true")
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch all class students")
		}
		defer rows.Close()

		for rows.Next() {
			var studentID string
			rows.Scan(&studentID)
			studentIDs = append(studentIDs, studentID)
		}

	case "student":
		// Get student IDs from assignments
		rows, err := db.Query("SELECT student_id FROM fee_type_assignments WHERE fee_type_id = $1", req.FeeTypeID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch fee assignments")
		}
		defer rows.Close()

		for rows.Next() {
			var studentID string
			rows.Scan(&studentID)
			studentIDs = append(studentIDs, studentID)
		}

	default:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid fee scope")
	}

	// Apply fees to all selected students
	for _, studentID := range studentIDs {
		_, err := db.Exec(`
			INSERT INTO fees (student_id, fee_type_id, academic_year_id, term_id, title, amount, balance, due_date, created_at, updated_at)
			VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $6, $7, NOW(), NOW())
		`, studentID, req.FeeTypeID, req.AcademicYearID, req.TermID, feeType.Name, req.Amount, req.DueDate)

		if err != nil {
			log.Printf("Failed to insert fee for student %s: %v", studentID, err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to apply fee to student: "+studentID)
		}
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"message":        "Fees applied successfully",
		"students_count": len(studentIDs),
	})
}
