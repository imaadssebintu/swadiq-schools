package fees

import (
	"database/sql"
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
	Scope            string    `json:"scope"`
	TargetClassID    *string   `json:"target_class_id"`
	TargetStudentID  *string   `json:"target_student_id"`
	Classes          []string  `json:"classes"`
	Students         []string  `json:"students"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateFeeTypeRequest struct {
	Name             string  `json:"name"`
	Code             string  `json:"code"`
	Description      string  `json:"description"`
	Amount           string  `json:"amount"`
	PaymentFrequency string  `json:"payment_frequency"`
	Scope            string  `json:"scope"`
	TargetClassID    string  `json:"target_class_id"`
	TargetStudentID  string  `json:"target_student_id"`
}

// GetFeeTypesAPI returns all fee types
func GetFeeTypesAPI(c *fiber.Ctx, db *sql.DB) error {
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
		WHERE ft.deleted_at IS NULL
		GROUP BY ft.id, ft.name, ft.code, ft.description, ft.amount, ft.payment_frequency, ft.is_active, ft.scope, ft.created_at, ft.updated_at
		ORDER BY ft.name`

	rows, err := db.Query(query)
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

	query := `INSERT INTO fee_types (name, code, description, amount, payment_frequency, scope, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) 
			  RETURNING id, created_at, updated_at`

	var feeType FeeTypeResponse
	scope := req.Scope
	if scope == "" {
		scope = "manual"
	}
	
	// Log the query parameters for debugging
	log.Printf("Inserting fee type with params: name=%s, code=%s, description=%s, payment_frequency=%s, scope=%s, target_class_id=%s, target_student_id=%s",
		req.Name, req.Code, req.Description, req.PaymentFrequency, scope, req.TargetClassID, req.TargetStudentID)
	
	err = db.QueryRow(query, req.Name, req.Code, req.Description, amount, req.PaymentFrequency, scope).Scan(
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
	} else if scope == "student" && req.TargetStudentID != "" {
		studentIDs := strings.Split(req.TargetStudentID, ",")
		for _, studentID := range studentIDs {
			if strings.TrimSpace(studentID) != "" {
				db.Exec(`INSERT INTO fee_type_assignments (fee_type_id, student_id) VALUES ($1, $2)`, feeType.ID, strings.TrimSpace(studentID))
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    feeType,
		"message": "Fee type created successfully",
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
		"success": true,
		"classes": classes,
		"students": students,
	})
}
