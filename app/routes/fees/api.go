package fees

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FeeResponse represents the response structure for fees
type FeeResponse struct {
	ID               string     `json:"id"`
	StudentID        string     `json:"student_id"`
	FeeTypeID        string     `json:"fee_type_id"`
	Title            string     `json:"title"`
	Amount           float64    `json:"amount"`
	Balance          float64    `json:"balance"`
	Paid             bool       `json:"paid"`
	DueDate          time.Time  `json:"due_date"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	StudentName      string     `json:"student_name,omitempty"`
	StudentCode      string     `json:"student_code,omitempty"`
	FeeTypeName      string     `json:"fee_type_name,omitempty"`
	FeeTypeCode      string     `json:"fee_type_code,omitempty"`
	AcademicYearName string     `json:"academic_year_name,omitempty"`
	TermName         string     `json:"term_name,omitempty"`
}

// FeeStatsResponse represents the response structure for fee statistics
type FeeStatsResponse struct {
	TotalFees        int     `json:"total_fees"`
	PaidFees         int     `json:"paid_fees"`
	UnpaidFees       int     `json:"unpaid_fees"`
	TotalAmount      float64 `json:"total_amount"`
	TotalPaid        float64 `json:"total_paid"`
	TotalBalance     float64 `json:"total_balance"`
	StudentsWithFees int     `json:"students_with_fees"`
}

// GetFeeStatsAPI returns fee statistics (unfiltered summary)
func GetFeeStatsAPI(c *fiber.Ctx, db *sql.DB) error {
	log.Println("Fetching fee statistics...")

	query := `SELECT 
				COUNT(*)::int as total_fees,
				COUNT(CASE WHEN paid = true THEN 1 END)::int as paid_fees,
				COUNT(CASE WHEN paid = false OR paid IS NULL THEN 1 END)::int as unpaid_fees,
				COALESCE(SUM(amount), 0)::float as total_amount,
				COALESCE(SUM(amount - COALESCE(balance, amount)), 0)::float as total_paid,
				COALESCE(SUM(COALESCE(balance, 0)), 0)::float as total_balance,
				COUNT(DISTINCT student_id)::int as students_with_fees
			  FROM fees 
			  WHERE deleted_at IS NULL`

	var stats FeeStatsResponse
	err := db.QueryRow(query).Scan(
		&stats.TotalFees, &stats.PaidFees, &stats.UnpaidFees,
		&stats.TotalAmount, &stats.TotalPaid, &stats.TotalBalance,
		&stats.StudentsWithFees,
	)
	if err != nil {
		log.Printf("Error fetching fee statistics: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to fetch statistics: " + err.Error()})
	}

	log.Printf("Fee stats fetched successfully: %+v", stats)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetFeesAPI returns all fees with optional filtering
func GetFeesAPI(c *fiber.Ctx, db *sql.DB) error {
	// Check if fees table exists
	var tableExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'fees')").Scan(&tableExists)
	if err != nil || !tableExists {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    []FeeResponse{},
		})
	}

	// Get query parameters for filtering
	studentSearch := c.Query("student")
	studentID := c.Query("student_id")
	yearID := c.Query("academic_year_id")
	termID := c.Query("term_id")
	status := c.Query("status") // "paid", "unpaid", "all"

	// Build base query with LEFT JOINs to handle missing data
	baseQuery := `SELECT f.id, f.student_id, COALESCE(f.fee_type_id::text, '') as fee_type_id, f.title, f.amount, f.balance, f.paid, 
				  f.due_date, f.paid_at, f.created_at, f.updated_at,
				  COALESCE(s.first_name, '') as student_first_name, COALESCE(s.last_name, '') as student_last_name,
				  COALESCE(s.student_id, '') as student_code,
				  COALESCE(ft.name, '') as fee_type_name, COALESCE(ft.code, '') as fee_type_code,
				  COALESCE(ay.name, '') as academic_year_name,
				  COALESCE(t.name, '') as term_name
				  FROM fees f
				  LEFT JOIN students s ON f.student_id = s.id
				  LEFT JOIN fee_types ft ON f.fee_type_id = ft.id
				  LEFT JOIN academic_years ay ON f.academic_year_id = ay.id
				  LEFT JOIN terms t ON f.term_id = t.id
				  WHERE f.deleted_at IS NULL`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add student search (by name or code) if provided
	if studentSearch != "" {
		conditions = append(conditions, fmt.Sprintf("(s.first_name ILIKE $%d OR s.last_name ILIKE $%d OR s.student_id ILIKE $%d)", argIndex, argIndex+1, argIndex+2))
		searchPattern := "%" + studentSearch + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	// Add exact student ID filter if provided
	if studentID != "" {
		conditions = append(conditions, fmt.Sprintf("f.student_id = $%d", argIndex))
		args = append(args, studentID)
		argIndex++
	}

	// Add academic year filter if provided
	if yearID != "" {
		conditions = append(conditions, fmt.Sprintf("f.academic_year_id = $%d", argIndex))
		args = append(args, yearID)
		argIndex++
	}

	// Add term filter if provided
	if termID != "" {
		conditions = append(conditions, fmt.Sprintf("f.term_id = $%d", argIndex))
		args = append(args, termID)
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

	// Add pagination
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    []FeeResponse{},
		})
	}
	defer rows.Close()

	var fees []FeeResponse
	for rows.Next() {
		var fee FeeResponse
		var studentFirstName, studentLastName, studentCode string
		var feeTypeName, feeTypeCode string
		var academicYearName, termName string
		var paidAt *time.Time

		err := rows.Scan(
			&fee.ID, &fee.StudentID, &fee.FeeTypeID, &fee.Title, &fee.Amount, &fee.Balance, &fee.Paid,
			&fee.DueDate, &paidAt, &fee.CreatedAt, &fee.UpdatedAt,
			&studentFirstName, &studentLastName, &studentCode,
			&feeTypeName, &feeTypeCode,
			&academicYearName, &termName,
		)
		if err != nil {
			continue
		}

		// Set student info
		fee.StudentName = strings.TrimSpace(studentFirstName + " " + studentLastName)
		fee.StudentCode = studentCode

		// Set fee type info
		fee.FeeTypeName = feeTypeName
		fee.FeeTypeCode = feeTypeCode

		// Set academic year and term info
		fee.AcademicYearName = academicYearName
		fee.TermName = termName

		// Set paid_at if exists
		if paidAt != nil {
			fee.PaidAt = paidAt
		}

		fees = append(fees, fee)
	}

	if fees == nil {
		fees = []FeeResponse{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fees,
	})
}

// GetFeeByIDAPI returns a specific fee by ID
func GetFeeByIDAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	query := `SELECT f.id, f.student_id, COALESCE(f.fee_type_id::text, '') as fee_type_id, f.title, f.amount, f.balance, f.paid, 
			  f.due_date, f.paid_at, f.created_at, f.updated_at,
			  COALESCE(s.first_name, '') as student_first_name, COALESCE(s.last_name, '') as student_last_name,
			  COALESCE(s.student_id, '') as student_code,
			  COALESCE(ft.name, '') as fee_type_name, COALESCE(ft.code, '') as fee_type_code
			  FROM fees f
			  LEFT JOIN students s ON f.student_id = s.id
			  LEFT JOIN fee_types ft ON f.fee_type_id = ft.id
			  WHERE f.id = $1 AND f.deleted_at IS NULL`

	var fee FeeResponse
	var studentFirstName, studentLastName, studentCode string
	var feeTypeName, feeTypeCode string
	var paidAt *time.Time

	err := db.QueryRow(query, feeID).Scan(
		&fee.ID, &fee.StudentID, &fee.FeeTypeID, &fee.Title, &fee.Amount, &fee.Balance, &fee.Paid,
		&fee.DueDate, &paidAt, &fee.CreatedAt, &fee.UpdatedAt,
		&studentFirstName, &studentLastName, &studentCode,
		&feeTypeName, &feeTypeCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee not found"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to fetch fee"})
	}

	// Set student info
	fee.StudentName = strings.TrimSpace(studentFirstName + " " + studentLastName)
	fee.StudentCode = studentCode
	fee.FeeTypeName = feeTypeName
	fee.FeeTypeCode = feeTypeCode
	if paidAt != nil {
		fee.PaidAt = paidAt
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fee,
	})
}

// CreateFeeRequest represents the request structure for creating fees
type CreateFeeRequest struct {
	StudentID string    `json:"student_id" validate:"required,uuid"`
	FeeTypeID string    `json:"fee_type_id" validate:"required,uuid"`
	TermID    *string   `json:"term_id,omitempty" validate:"omitempty,uuid"`
	Title     string    `json:"title" validate:"required"`
	Amount    float64   `json:"amount" validate:"required,gt=0"`
	DueDate   time.Time `json:"due_date" validate:"required"`
	Currency  string    `json:"currency,omitempty"`
}

// CreateFeeAPI creates a new fee
func CreateFeeAPI(c *fiber.Ctx, db *sql.DB) error {
	var req CreateFeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	// Validate required fields
	if req.StudentID == "" || req.FeeTypeID == "" || req.Title == "" || req.Amount <= 0 || req.DueDate.IsZero() {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Missing required fields"})
	}

	// Set default currency
	if req.Currency == "" {
		req.Currency = "UGX"
	}

	// Insert fee into database
	query := `INSERT INTO fees (student_id, fee_type_id, term_id, title, amount, balance, currency, paid, due_date, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $5, $6, false, $7, NOW(), NOW()) RETURNING id, created_at, updated_at`

	var fee models.Fee
	err := db.QueryRow(query, req.StudentID, req.FeeTypeID, req.TermID, req.Title, req.Amount, req.Currency, req.DueDate).Scan(
		&fee.ID, &fee.CreatedAt, &fee.UpdatedAt,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to create fee"})
	}

	// Set response data
	fee.StudentID = req.StudentID
	fee.FeeTypeID = req.FeeTypeID
	fee.TermID = req.TermID
	fee.Title = req.Title
	fee.Amount = req.Amount
	fee.Balance = req.Amount
	fee.Currency = req.Currency
	fee.Paid = false
	fee.DueDate = req.DueDate

	return c.Status(201).JSON(fiber.Map{
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
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	// Update fee in database
	query := `UPDATE fees SET title = $1, amount = $2, due_date = $3, updated_at = NOW()
			  WHERE id = $4 AND deleted_at IS NULL`

	result, err := db.Exec(query, fee.Title, fee.Amount, fee.DueDate, feeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to update fee"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee not found"})
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
	query := `UPDATE fees SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := db.Exec(query, feeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to delete fee"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee deleted successfully",
	})
}

// RecordPaymentRequest represents the request structure for recording payments
type RecordPaymentRequest struct {
	StudentID      string          `json:"student_id" validate:"required,uuid"`
	TotalAmount    float64         `json:"total_amount" validate:"required,gt=0"`
	PaymentMethod  string          `json:"payment_method" validate:"required"`
	TransactionID  *string         `json:"transaction_id,omitempty"`
	FeeAllocations []FeeAllocation `json:"fee_allocations" validate:"required,dive"`
}

type FeeAllocation struct {
	FeeID  string  `json:"fee_id" validate:"required,uuid"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// RecordPaymentAPI records a payment and allocates it to specific fees
func RecordPaymentAPI(c *fiber.Ctx, db *sql.DB) error {
	var req RecordPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Get user ID from context (set by AuthMiddleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "User not authenticated"})
	}

	// Insert payment record
	paymentQuery := `INSERT INTO payments (student_id, total_amount, payment_date, payment_method, paid_by, transaction_id, status, paid_at, created_at, updated_at)
					 VALUES ($1, $2, NOW(), $3, $4, $5, 'completed', NOW(), NOW(), NOW()) RETURNING id`

	var paymentID string
	err = tx.QueryRow(paymentQuery, req.StudentID, req.TotalAmount, req.PaymentMethod, userID, req.TransactionID).Scan(&paymentID)
	if err != nil {
		log.Printf("Error creating payment record: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to create payment record"})
	}

	// Process each fee allocation
	for _, allocation := range req.FeeAllocations {
		// Get current fee details
		var currentBalance float64
		var feeTypeID string
		err = tx.QueryRow("SELECT balance, fee_type_id FROM fees WHERE id = $1", allocation.FeeID).Scan(&currentBalance, &feeTypeID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Fee not found: " + allocation.FeeID})
		}

		// Calculate new balance
		newBalance := currentBalance - allocation.Amount
		if newBalance < 0 {
			newBalance = 0
		}

		// Insert payment allocation
		allocationQuery := `INSERT INTO payment_allocations (payment_id, fee_id, fee_type_id, amount, balance, is_fully_paid, created_at, updated_at)
							   VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

		_, err = tx.Exec(allocationQuery, paymentID, allocation.FeeID, feeTypeID, allocation.Amount, newBalance, newBalance == 0)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to create payment allocation"})
		}

		// Update fee balance and paid status
		updateFeeQuery := `UPDATE fees SET balance = $1, paid = $2, updated_at = NOW()`
		if newBalance == 0 {
			updateFeeQuery += `, paid_at = NOW()`
		}
		updateFeeQuery += ` WHERE id = $3`

		_, err = tx.Exec(updateFeeQuery, newBalance, newBalance == 0, allocation.FeeID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to update fee"})
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"payment_id": paymentID,
			"message":    "Payment recorded successfully",
		},
	})
}

// GetStudentFeesAPI returns all fees for a specific student with payment details
func GetStudentFeesAPI(c *fiber.Ctx, db *sql.DB) error {
	studentID := c.Params("student_id")
	termID := c.Query("term_id")

	query := `SELECT f.id, f.student_id, f.fee_type_id, f.term_id, f.title, f.amount, f.balance, 
				 f.currency, f.paid, f.due_date, f.paid_at, f.created_at, f.updated_at,
				 ft.name as fee_type_name, ft.code as fee_type_code,
				 t.name as term_name,
				 COALESCE(SUM(pa.amount), 0) as total_paid
			  FROM fees f
			  LEFT JOIN fee_types ft ON f.fee_type_id = ft.id
			  LEFT JOIN terms t ON f.term_id = t.id
			  LEFT JOIN payment_allocations pa ON f.id = pa.fee_id
			  WHERE f.student_id = $1 AND f.deleted_at IS NULL`

	args := []interface{}{studentID}
	if termID != "" {
		query += " AND f.term_id = $2"
		args = append(args, termID)
	}

	query += ` GROUP BY f.id, ft.name, ft.code, t.name ORDER BY f.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to fetch fees"})
	}
	defer rows.Close()

	type StudentFeeResponse struct {
		ID          string     `json:"id"`
		StudentID   string     `json:"student_id"`
		FeeTypeID   string     `json:"fee_type_id"`
		TermID      *string    `json:"term_id"`
		Title       string     `json:"title"`
		Amount      float64    `json:"amount"`
		Balance     float64    `json:"balance"`
		TotalPaid   float64    `json:"total_paid"`
		Currency    string     `json:"currency"`
		Paid        bool       `json:"paid"`
		DueDate     time.Time  `json:"due_date"`
		PaidAt      *time.Time `json:"paid_at"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		FeeTypeName string     `json:"fee_type_name"`
		FeeTypeCode string     `json:"fee_type_code"`
		TermName    *string    `json:"term_name"`
	}

	var fees []StudentFeeResponse
	for rows.Next() {
		var fee StudentFeeResponse
		err := rows.Scan(
			&fee.ID, &fee.StudentID, &fee.FeeTypeID, &fee.TermID, &fee.Title, &fee.Amount, &fee.Balance,
			&fee.Currency, &fee.Paid, &fee.DueDate, &fee.PaidAt, &fee.CreatedAt, &fee.UpdatedAt,
			&fee.FeeTypeName, &fee.FeeTypeCode, &fee.TermName, &fee.TotalPaid,
		)
		if err != nil {
			continue
		}
		fees = append(fees, fee)
	}

	if fees == nil {
		fees = []StudentFeeResponse{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fees,
	})
}

// MarkFeeAsPaidAPI marks a fee as paid
func MarkFeeAsPaidAPI(c *fiber.Ctx, db *sql.DB) error {
	feeID := c.Params("id")

	// Update fee as paid
	query := `UPDATE fees SET paid = true, paid_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := db.Exec(query, feeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to mark fee as paid"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "Fee not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Fee marked as paid successfully",
	})
}

// GetStudentsForClassesAPI returns students from specific classes with pagination and search
func GetStudentsForClassesAPI(c *fiber.Ctx, db *sql.DB) error {
	classIDsParam := c.Query("class_ids")
	if classIDsParam == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "class_ids is required"})
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
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "Failed to fetch students"})
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
