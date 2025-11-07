package fees

import (
	"database/sql"
	"fmt"
	"strings"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FeeResponse represents the response structure for fees
type FeeResponse struct {
	ID           string     `json:"id"`
	StudentID    string     `json:"student_id"`
	FeeTypeID    string     `json:"fee_type_id"`
	Title        string     `json:"title"`
	Amount       float64    `json:"amount"`
	Paid         bool       `json:"paid"`
	DueDate      time.Time  `json:"due_date"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StudentName  string     `json:"student_name,omitempty"`
	StudentCode  string     `json:"student_code,omitempty"`
	FeeTypeName  string     `json:"fee_type_name,omitempty"`
	FeeTypeCode  string     `json:"fee_type_code,omitempty"`
}

// FeeStatsResponse represents the response structure for fee statistics
type FeeStatsResponse struct {
	TotalFees        int     `json:"total_fees"`
	PaidFees         int     `json:"paid_fees"`
	UnpaidFees       int     `json:"unpaid_fees"`
	TotalPaid        float64 `json:"total_paid"`
	TotalUnpaid      float64 `json:"total_unpaid"`
	StudentsWithFees int     `json:"students_with_fees"`
}

// GetFeesAPI returns all fees with optional filtering
func GetFeesAPI(c *fiber.Ctx, db *sql.DB) error {
	// Get query parameters for filtering
	studentID := c.Query("student_id")
	status := c.Query("status") // "paid", "unpaid", "all"

	// Build base query
	baseQuery := `SELECT f.id, f.student_id, '', f.title, f.amount, f.paid, 
				  f.due_date, f.paid_at, f.created_at, f.updated_at,
				  s.first_name as student_first_name, s.last_name as student_last_name,
				  s.student_id as student_code,
				  '' as fee_type_name, '' as fee_type_code
				  FROM fees f
				  JOIN students s ON f.student_id = s.id
				  WHERE s.is_active = true AND f.deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add student filter if provided
	if studentID != "" {
		conditions = append(conditions, fmt.Sprintf("f.student_id = $%d", argIndex))
		args = append(args, studentID)
		argIndex++
	}

	// Add status filter if provided
	if status == "paid" {
		conditions = append(conditions, fmt.Sprintf("f.paid = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if status == "unpaid" {
		conditions = append(conditions, fmt.Sprintf("f.paid = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	// Add conditions to query
	if len(conditions) > 0 {
		baseQuery += " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			baseQuery += " AND " + conditions[i]
		}
	}

	// Add ordering
	baseQuery += " ORDER BY f.created_at DESC"

	// Execute query
	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch fees")
	}
	defer rows.Close()

	var fees []FeeResponse
	for rows.Next() {
		var fee FeeResponse
		var studentFirstName, studentLastName, studentCode *string
		var feeTypeName, feeTypeCode *string
		var paidAt *time.Time

		err := rows.Scan(
			&fee.ID, &fee.StudentID, &fee.FeeTypeID, &fee.Title, &fee.Amount, &fee.Paid,
			&fee.DueDate, &paidAt, &fee.CreatedAt, &fee.UpdatedAt,
			&studentFirstName, &studentLastName, &studentCode,
			&feeTypeName, &feeTypeCode,
		)
		if err != nil {
			continue
		}

		// Set student info
		if studentFirstName != nil && studentLastName != nil {
			fee.StudentName = *studentFirstName + " " + *studentLastName
		}
		if studentCode != nil {
			fee.StudentCode = *studentCode
		}

		// Set fee type info
		if feeTypeName != nil {
			fee.FeeTypeName = *feeTypeName
		}
		if feeTypeCode != nil {
			fee.FeeTypeCode = *feeTypeCode
		}

		// Set paid_at if exists
		if paidAt != nil {
			fee.PaidAt = paidAt
		}

		fees = append(fees, fee)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fees,
	})
}

// GetFeeByIDAPI returns a specific fee by ID
func GetFeeByIDAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	query := `SELECT f.id, f.student_id, f.title, f.amount, f.paid, 
			  f.due_date, f.paid_at, f.created_at, f.updated_at,
			  s.first_name as student_first_name, s.last_name as student_last_name,
			  s.student_id as student_code
			  FROM fees f
			  JOIN students s ON f.student_id = s.id
			  WHERE f.id = $1 AND s.is_active = true`

	var fee FeeResponse
	var studentFirstName, studentLastName, studentCode *string
	var paidAt *time.Time

	err := db.QueryRow(query, feeID).Scan(
		&fee.ID, &fee.StudentID, &fee.Title, &fee.Amount, &fee.Paid,
		&fee.DueDate, &paidAt, &fee.CreatedAt, &fee.UpdatedAt,
		&studentFirstName, &studentLastName, &studentCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Fee not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch fee")
	}

	// Set student info
	if studentFirstName != nil && studentLastName != nil {
		fee.StudentName = *studentFirstName + " " + *studentLastName
	}
	if studentCode != nil {
		fee.StudentCode = *studentCode
	}
	if paidAt != nil {
		fee.PaidAt = paidAt
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fee,
	})
}

// CreateFeeAPI creates a new fee
func CreateFeeAPI(c *fiber.Ctx, db *sql.DB) error {
	var fee models.Fee
	if err := c.BodyParser(&fee); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if fee.StudentID == "" || fee.Title == "" || fee.Amount <= 0 || fee.DueDate.IsZero() {
		return fiber.NewError(fiber.StatusBadRequest, "Missing required fields")
	}

	// Set default values
	fee.Paid = false
	if fee.Currency == "" {
		fee.Currency = "USD"
	}

	// Insert fee into database
	query := `INSERT INTO fees (student_id, title, amount, currency, due_date, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, fee.StudentID, fee.Title, fee.Amount, fee.Currency, fee.DueDate).Scan(
		&fee.ID, &fee.CreatedAt, &fee.UpdatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create fee")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    fee,
		"message": "Fee created successfully",
	})
}

// UpdateFeeAPI updates an existing fee
func UpdateFeeAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	var fee models.Fee
	if err := c.BodyParser(&fee); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Update fee in database
	query := `UPDATE fees SET title = $1, amount = $2, due_date = $3, updated_at = NOW()
			  WHERE id = $4`

	result, err := db.Exec(query, fee.Title, fee.Amount, fee.DueDate, feeID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update fee")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Fee not found")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee updated successfully",
	})
}

// DeleteFeeAPI deletes a fee
func DeleteFeeAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	// Soft delete fee (set deleted_at)
	query := `UPDATE fees SET deleted_at = NOW() WHERE id = $1`

	result, err := db.Exec(query, feeID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete fee")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Fee not found")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee deleted successfully",
	})
}

// MarkFeeAsPaidAPI marks a fee as paid
func MarkFeeAsPaidAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	// Update fee as paid
	query := `UPDATE fees SET paid = true, paid_at = NOW(), updated_at = NOW() WHERE id = $1`

	result, err := db.Exec(query, feeID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to mark fee as paid")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Fee not found")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee marked as paid successfully",
	})
}

// GetFeeStatsAPI returns fee statistics
func GetFeeStatsAPI(c *fiber.Ctx, db *sql.DB) error {
	query := `
		SELECT 
			COUNT(*) as total_fees,
			COUNT(CASE WHEN paid = true THEN 1 END) as paid_fees,
			COUNT(CASE WHEN paid = false THEN 1 END) as unpaid_fees,
			COALESCE(SUM(CASE WHEN paid = true THEN amount END), 0) as total_paid,
			COALESCE(SUM(CASE WHEN paid = false THEN amount END), 0) as total_unpaid,
			COUNT(DISTINCT student_id) as students_with_fees
		FROM fees 
		WHERE deleted_at IS NULL
	`

	stats := FeeStatsResponse{
		TotalFees:        0,
		PaidFees:         0,
		UnpaidFees:       0,
		TotalPaid:        0,
		TotalUnpaid:      0,
		StudentsWithFees: 0,
	}

	db.QueryRow(query).Scan(
		&stats.TotalFees, &stats.PaidFees, &stats.UnpaidFees,
		&stats.TotalPaid, &stats.TotalUnpaid, &stats.StudentsWithFees,
	)
	// Ignore errors and return zero stats - this ensures the frontend always gets valid data

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetStudentsForClassesAPI returns students from specific classes with pagination and search
func GetStudentsForClassesAPI(c *fiber.Ctx, db *sql.DB) error {
	classIDsParam := c.Query("class_ids")
	if classIDsParam == "" {
		return c.Status(400).JSON(fiber.Map{"error": "class_ids is required"})
	}
	
	classIDs := strings.Split(classIDsParam, ",")
	search := c.Query("search")
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)
	
	// Build placeholders for class IDs
	placeholders := make([]string, len(classIDs))
	args := make([]interface{}, 0)
	for i, classID := range classIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, strings.TrimSpace(classID))
	}
	
	// Simple query that works with basic student table structure
	query := fmt.Sprintf(`SELECT id, first_name, last_name, student_id, class_id 
						  FROM students 
						  WHERE is_active = true AND class_id IN (%s)`, 
						  strings.Join(placeholders, ","))
	
	// Add search conditions if provided
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query += fmt.Sprintf(` AND (LOWER(student_id) LIKE $%d 
							   OR LOWER(first_name) LIKE $%d 
							   OR LOWER(last_name) LIKE $%d 
							   OR LOWER(first_name || ' ' || last_name) LIKE $%d)`, 
							   len(args)+1, len(args)+2, len(args)+3, len(args)+4)
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}
	
	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY first_name, last_name LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch students")
	}
	defer rows.Close()
	
	students := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, firstName, lastName, studentID, classID string
		if err := rows.Scan(&id, &firstName, &lastName, &studentID, &classID); err != nil {
			continue
		}
		students = append(students, map[string]interface{}{
			"id":         id,
			"first_name": firstName,
			"last_name":  lastName,
			"student_id": studentID,
			"class_id":   classID,
		})
	}
	
	return c.JSON(fiber.Map{
		"success":     true,
		"students":    students,
		"total_count": len(students),
		"has_more":    len(students) == limit,
		"next_offset": offset + len(students),
	})
}
