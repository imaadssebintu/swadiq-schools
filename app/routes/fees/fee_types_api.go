package fees

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FeeTypeResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Code             string    `json:"code"`
	Description      *string   `json:"description"`
	Amount           int       `json:"amount"`
	PaymentFrequency string    `json:"payment_frequency"`
	IsActive         bool      `json:"is_active"`
	IsRequired       bool      `json:"is_required"`
	Scope            string    `json:"scope"`
	TargetClassID    *string   `json:"target_class_id"`
	TargetStudentID  *string   `json:"target_student_id"`
	Classes          []string  `json:"classes"`
	Students         []string  `json:"students"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateFeeTypeRequest struct {
	Name             string `json:"name"`
	Code             string `json:"code"`
	Description      string `json:"description"`
	Amount           string `json:"amount"`
	PaymentFrequency string `json:"payment_frequency"`
	Scope            string `json:"scope"`
	TargetClassID    string `json:"target_class_id"`
	TargetStudentID  string `json:"target_student_id"`
}

// autoApplyFees automatically creates fee records for students based on fee type scope
func autoApplyFees(db *sql.DB, feeTypeID, scope, targetClassID, targetStudentID string, amount float64, feeTitle string) (int, error) {
	var studentIDs []string

	// Get current academic year and term
	var academicYearID, termID sql.NullString
	db.QueryRow(`SELECT id FROM academic_years WHERE is_current = true LIMIT 1`).Scan(&academicYearID)
	db.QueryRow(`SELECT id FROM terms WHERE is_current = true LIMIT 1`).Scan(&termID)

	// Calculate due date (30 days from now)
	dueDate := time.Now().AddDate(0, 0, 30)

	// Get student IDs based on scope
	switch scope {
	case "all_students":
		rows, err := db.Query(`SELECT id FROM students WHERE deleted_at IS NULL AND is_active = true`)
		if err != nil {
			return 0, err
		}
		defer rows.Close()
		for rows.Next() {
			var studentID string
			rows.Scan(&studentID)
			studentIDs = append(studentIDs, studentID)
		}

	case "class":
		if targetClassID != "" {
			classIDs := strings.Split(targetClassID, ",")
			for _, classID := range classIDs {
				if strings.TrimSpace(classID) != "" {
					rows, err := db.Query(`SELECT student_id FROM class_students WHERE class_id = $1`, strings.TrimSpace(classID))
					if err != nil {
						continue
					}
					for rows.Next() {
						var studentID string
						rows.Scan(&studentID)
						studentIDs = append(studentIDs, studentID)
					}
					rows.Close()
				}
			}
		}

	case "student":
		if targetStudentID != "" {
			studentIDList := strings.Split(targetStudentID, ",")
			for _, studentID := range studentIDList {
				if strings.TrimSpace(studentID) != "" {
					studentIDs = append(studentIDs, strings.TrimSpace(studentID))
				}
			}
		}

	default:
		return 0, nil // Manual scope, no auto-application
	}

	// Create fee records for all selected students
	for _, studentID := range studentIDs {
		// Check if fee already exists for this student and fee type
		var existingCount int
		db.QueryRow(`SELECT COUNT(*) FROM fees WHERE student_id = $1 AND fee_type_id = $2 AND deleted_at IS NULL`,
			studentID, feeTypeID).Scan(&existingCount)

		if existingCount == 0 {
			_, err := db.Exec(`
				INSERT INTO fees (student_id, fee_type_id, academic_year_id, term_id, title, amount, balance, due_date, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $6, $7, NOW(), NOW())
			`, studentID, feeTypeID, academicYearID, termID, feeTitle, amount, dueDate)

			if err != nil {
				log.Printf("Failed to create fee for student %s: %v", studentID, err)
				continue
			}
		}
	}

	return len(studentIDs), nil
}

// GetFeeTypesAPI returns all fee types with optional filtering by class or student
func GetFeeTypesAPI(c *fiber.Ctx, db *sql.DB) error {
	classID := c.Query("class_id")
	studentID := c.Query("student_id")

	query := `SELECT 
		ft.id, ft.name, ft.code, ft.description, COALESCE(ft.amount, 0) as amount,
		COALESCE(ft.payment_frequency, 'per_term') as payment_frequency,
		ft.is_active, COALESCE(ft.scope, 'manual') as scope,
		ft.created_at, ft.updated_at,
		COALESCE(array_agg(DISTINCT c.name) FILTER (WHERE c.name IS NOT NULL), '{}') as classes,
		COALESCE(array_agg(DISTINCT s.first_name || ' ' || s.last_name) FILTER (WHERE s.first_name IS NOT NULL), '{}') as students
		FROM fee_types ft
		LEFT JOIN fee_type_assignments fta ON ft.id = fta.fee_type_id
		LEFT JOIN classes c ON fta.class_id = c.id
		LEFT JOIN students s ON fta.student_id = s.id
		WHERE ft.deleted_at IS NULL`

	var args []interface{}
	if classID != "" || studentID != "" {
		query += ` AND (ft.scope IN ('all_students', 'all_classes', 'manual')`
		argIdx := 1
		if classID != "" {
			query += fmt.Sprintf(` OR (ft.scope = 'class' AND fta.class_id = $%d)`, argIdx)
			args = append(args, classID)
			argIdx++
		}
		if studentID != "" {
			query += fmt.Sprintf(` OR (ft.scope = 'student' AND fta.student_id = $%d)`, argIdx)
			args = append(args, studentID)
			argIdx++
		}
		query += `)`
	}

	query += ` GROUP BY ft.id, ft.name, ft.code, ft.description, ft.amount, ft.payment_frequency, ft.is_active, ft.scope, ft.created_at, ft.updated_at
		ORDER BY ft.name`

	rows, err := db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch fee types: " + err.Error(),
		})
	}
	defer rows.Close()

	var feeTypes []FeeTypeResponse
	for rows.Next() {
		var feeType FeeTypeResponse
		var classesArray, studentsArray string
		err := rows.Scan(
			&feeType.ID, &feeType.Name, &feeType.Code, &feeType.Description, &feeType.Amount,
			&feeType.PaymentFrequency, &feeType.IsActive, &feeType.Scope,
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

// CreateFeeTypeAPI creates a new fee type
func CreateFeeTypeAPI(c *fiber.Ctx, db *sql.DB) error {
	// Log raw body for debugging
	log.Printf("Raw request body: %s", string(c.Body()))

	var req CreateFeeTypeRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Body parsing error: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Log the request for debugging
	log.Printf("CreateFeeTypeAPI request: %+v", req)

	// Convert amount string to int
	amount, err := strconv.Atoi(req.Amount)
	if err != nil || amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Valid amount is required")
	}

	if req.Name == "" || req.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name and code are required")
	}

	if req.PaymentFrequency == "" {
		req.PaymentFrequency = "per_term"
	}

	query := `INSERT INTO fee_types (name, code, description, amount, payment_frequency, scope, is_required, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) 
			  RETURNING id, created_at, updated_at`

	var feeType FeeTypeResponse
	scope := req.Scope
	if scope == "" {
		scope = "manual"
	}

	// Determine if this is a required fee (must pay)
	isRequired := scope != "manual"

	// Log the query parameters for debugging
	log.Printf("Inserting fee type with params: name=%s, code=%s, description=%s, payment_frequency=%s, scope=%s, is_required=%t, target_class_id=%s, target_student_id=%s",
		req.Name, req.Code, req.Description, req.PaymentFrequency, scope, isRequired, req.TargetClassID, req.TargetStudentID)

	err = db.QueryRow(query, req.Name, req.Code, req.Description, amount, req.PaymentFrequency, scope, isRequired).Scan(
		&feeType.ID, &feeType.CreatedAt, &feeType.UpdatedAt,
	)
	if err != nil {
		log.Printf("Failed to create fee type: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create fee type: "+err.Error())
	}

	feeType.Name = req.Name
	feeType.Code = req.Code
	feeType.Amount = amount
	feeType.PaymentFrequency = req.PaymentFrequency
	feeType.IsRequired = isRequired
	if req.Description != "" {
		feeType.Description = &req.Description
	}
	feeType.IsActive = true

	// Save assignments based on scope
	if scope == "class" && req.TargetClassID != "" {
		classIDs := strings.Split(req.TargetClassID, ",")
		for _, classID := range classIDs {
			if strings.TrimSpace(classID) != "" {
				db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, class_id) VALUES ($1, $2)`, feeType.ID, strings.TrimSpace(classID))
			}
		}
	} else if scope == "all_classes" {
		// Get all active classes and create assignments
		rows, err := db.Query(`SELECT id FROM classes WHERE deleted_at IS NULL`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var classID string
				if rows.Scan(&classID) == nil {
					db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, class_id) VALUES ($1, $2)`, feeType.ID, classID)
				}
			}
		}
	} else if scope == "student" && req.TargetStudentID != "" {
		studentIDs := strings.Split(req.TargetStudentID, ",")
		for _, studentID := range studentIDs {
			if strings.TrimSpace(studentID) != "" {
				db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, student_id) VALUES ($1, $2)`, feeType.ID, strings.TrimSpace(studentID))
			}
		}
	}

	// Auto-apply fees if this is a required fee type
	if isRequired {
		var studentsAffected int
		studentsAffected, err = autoApplyFees(db, feeType.ID, scope, req.TargetClassID, req.TargetStudentID, float64(amount), req.Name)
		if err != nil {
			log.Printf("Failed to auto-apply fees: %v", err)
			// Don't fail the creation, just log the error
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"data":    feeType,
			"message": "Fee type created and automatically applied to " + strconv.Itoa(studentsAffected) + " students",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    feeType,
		"message": "Fee type created successfully",
	})
}

// GetFeeTypeAPI returns a single fee type by ID
func GetFeeTypeAPI(c *fiber.Ctx, db *sql.DB) error {
	feeTypeID := c.Params("id")

	query := `SELECT id, name, code, description, COALESCE(amount, 0) as amount,
			  COALESCE(payment_frequency, 'per_term') as payment_frequency,
			  is_active, COALESCE(scope, 'manual') as scope, is_required,
			  created_at, updated_at
			  FROM fee_types WHERE id = $1 AND deleted_at IS NULL`

	var feeType FeeTypeResponse
	err := db.QueryRow(query, feeTypeID).Scan(
		&feeType.ID, &feeType.Name, &feeType.Code, &feeType.Description, &feeType.Amount,
		&feeType.PaymentFrequency, &feeType.IsActive, &feeType.Scope, &feeType.IsRequired,
		&feeType.CreatedAt, &feeType.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee type not found"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to fetch fee type"})
	}

	// Get current assignments
	assignQuery := `SELECT 
					COALESCE(array_agg(DISTINCT c.id) FILTER (WHERE c.id IS NOT NULL), '{}') as class_ids,
					COALESCE(array_agg(DISTINCT s.id) FILTER (WHERE s.id IS NOT NULL), '{}') as student_ids
					FROM fee_type_assignments fta
					LEFT JOIN classes c ON fta.class_id = c.id
					LEFT JOIN students s ON fta.student_id = s.id
					WHERE fta.fee_type_id = $1`

	var classIDs, studentIDs string
	db.QueryRow(assignQuery, feeTypeID).Scan(&classIDs, &studentIDs)

	// Parse class IDs
	if classIDs != "{}" && classIDs != "" {
		classIDs = strings.Trim(classIDs, "{}")
		if classIDs != "" {
			feeType.TargetClassID = &classIDs
		}
	}

	// Parse student IDs
	if studentIDs != "{}" && studentIDs != "" {
		studentIDs = strings.Trim(studentIDs, "{}")
		if studentIDs != "" {
			feeType.TargetStudentID = &studentIDs
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    feeType,
	})
}

// UpdateFeeTypeAPI updates an existing fee type
func UpdateFeeTypeAPI(c *fiber.Ctx, db *sql.DB) error {
	feeTypeID := c.Params("id")

	type UpdateFeeTypeRequest struct {
		Name             string `json:"name"`
		Code             string `json:"code"`
		Description      string `json:"description"`
		Amount           string `json:"amount"`
		PaymentFrequency string `json:"payment_frequency"`
		IsRequired       string `json:"is_required"`
		IsActive         string `json:"is_active"`
		Scope            string `json:"scope"`
		TargetClassID    string `json:"target_class_id"`
		TargetStudentID  string `json:"target_student_id"`
	}

	var req UpdateFeeTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.Name == "" || req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Name and code are required"})
	}

	amount, err := strconv.Atoi(req.Amount)
	if err != nil || amount < 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Valid amount is required"})
	}

	isRequired := req.IsRequired == "true"
	isActive := req.IsActive == "true"
	scope := req.Scope
	if scope == "" {
		scope = "manual"
	}

	query := `UPDATE fee_types SET name = $1, code = $2, description = $3, amount = $4,
			  payment_frequency = $5, is_required = $6, is_active = $7, scope = $8, updated_at = NOW()
			  WHERE id = $9 AND deleted_at IS NULL`

	result, err := db.Exec(query, req.Name, req.Code, req.Description, amount,
		req.PaymentFrequency, isRequired, isActive, scope, feeTypeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to update fee type"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee type not found"})
	}

	// Update assignments
	db.Exec(`DELETE FROM fee_type_assignments WHERE fee_type_id = $1`, feeTypeID)

	if scope == "class" && req.TargetClassID != "" {
		classIDs := strings.Split(req.TargetClassID, ",")
		for _, classID := range classIDs {
			if strings.TrimSpace(classID) != "" {
				db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, class_id) VALUES ($1, $2)`, feeTypeID, strings.TrimSpace(classID))
			}
		}
	} else if scope == "all_classes" {
		// Get all active classes and create assignments
		rows, err := db.Query(`SELECT id FROM classes WHERE deleted_at IS NULL`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var classID string
				if rows.Scan(&classID) == nil {
					db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, class_id) VALUES ($1, $2)`, feeTypeID, classID)
				}
			}
		}
	} else if scope == "student" && req.TargetStudentID != "" {
		studentIDs := strings.Split(req.TargetStudentID, ",")
		for _, studentID := range studentIDs {
			if strings.TrimSpace(studentID) != "" {
				db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, student_id) VALUES ($1, $2)`, feeTypeID, strings.TrimSpace(studentID))
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee type updated successfully",
	})
}

// DeleteFeeTypeAPI deletes a fee type
func DeleteFeeTypeAPI(c *fiber.Ctx, db *sql.DB) error {
	feeTypeID := c.Params("id")

	// Check if there are existing fees using this fee type
	var feeCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM fees WHERE fee_type_id = $1 AND deleted_at IS NULL`, feeTypeID).Scan(&feeCount)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to check fee dependencies"})
	}

	if feeCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Cannot delete fee type. There are " + strconv.Itoa(feeCount) + " existing fees using this fee type. Please remove or reassign these fees first.",
		})
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Delete assignments first
	_, err = tx.Exec(`DELETE FROM fee_type_assignments WHERE fee_type_id = $1`, feeTypeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to delete fee type assignments"})
	}

	// Soft delete fee type
	result, err := tx.Exec(`UPDATE fee_types SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, feeTypeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to delete fee type"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Fee type not found"})
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to commit deletion"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee type deleted successfully",
	})
}

// GetFeeTypeAssignmentsAPI returns assignments for a fee type
func GetFeeTypeAssignmentsAPI(c *fiber.Ctx, db *sql.DB) error {
	feeTypeID := c.Params("id")

	query := `SELECT 
			COALESCE(c.name, '') as class_name,
			COALESCE(s.first_name || ' ' || s.last_name, '') as student_name
		  FROM fee_type_assignments fta
		  LEFT JOIN classes c ON fta.class_id = c.id
		  LEFT JOIN students s ON fta.student_id = s.id
		  WHERE fta.fee_type_id = $1`

	rows, err := db.Query(query, feeTypeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch assignments"})
	}
	defer rows.Close()

	var classes []string
	var students []string

	for rows.Next() {
		var className, studentName string
		rows.Scan(&className, &studentName)
		if className != "" {
			classes = append(classes, className)
		}
		if studentName != "" {
			students = append(students, studentName)
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"classes":  classes,
		"students": students,
	})
}
