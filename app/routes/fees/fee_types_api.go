package fees

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FeeTypeResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Code             string    `json:"code"`
	Description      *string   `json:"description"`
	PaymentFrequency string    `json:"payment_frequency"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateFeeTypeRequest struct {
	Name             string `json:"name"`
	Code             string `json:"code"`
	Description      string `json:"description"`
	PaymentFrequency string `json:"payment_frequency"`
}

// GetFeeTypesAPI returns all fee types
func GetFeeTypesAPI(c *fiber.Ctx, db *sql.DB) error {
	query := `SELECT id, name, code, description, payment_frequency, is_active, created_at, updated_at 
			  FROM fee_types 
			  WHERE deleted_at IS NULL 
			  ORDER BY name`

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
		err := rows.Scan(
			&feeType.ID, &feeType.Name, &feeType.Code, &feeType.Description,
			&feeType.PaymentFrequency, &feeType.IsActive, &feeType.CreatedAt, &feeType.UpdatedAt,
		)
		if err != nil {
			continue
		}
		feeTypes = append(feeTypes, feeType)
	}

	// Ensure feeTypes is not nil
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
	var req CreateFeeTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" || req.Code == "" || req.PaymentFrequency == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, code, and payment frequency are required")
	}

	query := `INSERT INTO fee_types (name, code, description, payment_frequency, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW()) 
			  RETURNING id, created_at, updated_at`

	var feeType FeeTypeResponse
	err := db.QueryRow(query, req.Name, req.Code, req.Description, req.PaymentFrequency).Scan(
		&feeType.ID, &feeType.CreatedAt, &feeType.UpdatedAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create fee type")
	}

	feeType.Name = req.Name
	feeType.Code = req.Code
	feeType.PaymentFrequency = req.PaymentFrequency
	if req.Description != "" {
		feeType.Description = &req.Description
	}
	feeType.IsActive = true

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    feeType,
		"message": "Fee type created successfully",
	})
}
