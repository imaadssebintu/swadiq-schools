package fees

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ActivateFeesRequest struct {
	FeeTypeID string `json:"fee_type_id"`
}

// ActivateFeesAPI activates fees for students to pay based on existing assignments
func ActivateFeesAPI(c *fiber.Ctx, db *sql.DB) error {
	var req ActivateFeesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.FeeTypeID == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Fee type ID is required",
		})
	}

	// Check if fee type exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM fee_types WHERE id = $1)", req.FeeTypeID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking fee type existence: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Database error: " + err.Error(),
		})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Fee type not found",
		})
	}

	// Activate the fee type
	result, err := db.Exec("UPDATE fee_types SET is_active = $1 WHERE id = $2", true, req.FeeTypeID)
	if err != nil {
		log.Printf("Error activating fee type: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Database error: " + err.Error(),
		})
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Rows affected: %d", rowsAffected)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee type activated successfully",
	})
}

// GetActiveFeeTypesAPI returns all active fee types with their assignments
func GetActiveFeeTypesAPI(c *fiber.Ctx, db *sql.DB) error {
	query := `SELECT 
		ft.id, ft.name, ft.code, ft.description, COALESCE(ft.amount, 0) as amount,
		COALESCE(ft.payment_frequency, 'per_term') as payment_frequency,
		ft.is_required, COALESCE(ft.scope, 'manual') as scope,
		ft.created_at, ft.updated_at,
		COALESCE(array_agg(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') as classes,
		COALESCE(array_agg(DISTINCT s.first_name || ' ' || s.last_name) FILTER (WHERE s.first_name IS NOT NULL), '{}') as students
		FROM fee_types ft
		LEFT JOIN fee_type_assignments fta ON ft.id = fta.fee_type_id
		LEFT JOIN classes c ON fta.class_id = c.id
		LEFT JOIN students s ON fta.student_id = s.id
		WHERE ft.is_active = true AND ft.deleted_at IS NULL
		GROUP BY ft.id, ft.name, ft.code, ft.description, ft.amount, ft.payment_frequency, ft.is_required, ft.scope, ft.created_at, ft.updated_at
		ORDER BY ft.name`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch active fee types",
		})
	}
	defer rows.Close()

	var feeTypes []FeeTypeResponse
	for rows.Next() {
		var feeType FeeTypeResponse
		var classesArray, studentsArray string
		err := rows.Scan(
			&feeType.ID, &feeType.Name, &feeType.Code, &feeType.Description, &feeType.Amount,
			&feeType.PaymentFrequency, &feeType.IsRequired, &feeType.Scope,
			&feeType.CreatedAt, &feeType.UpdatedAt, &classesArray, &studentsArray,
		)
		if err != nil {
			continue
		}
		
		// Parse arrays
		if classesArray != "{}" {
			classesArray = strings.Trim(classesArray, "{}")
			if classesArray != "" {
				feeType.Classes = strings.Split(classesArray, ",")
			}
		}
		if studentsArray != "{}" {
			studentsArray = strings.Trim(studentsArray, "{}")
			if studentsArray != "" {
				feeType.Students = strings.Split(studentsArray, ",")
			}
		}
		
		feeType.IsActive = true
		feeTypes = append(feeTypes, feeType)
	}

	if feeTypes == nil {
		feeTypes = []FeeTypeResponse{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    feeTypes,
	})
}