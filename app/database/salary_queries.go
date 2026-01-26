package database

import (
	"database/sql"

	"swadiq-schools/app/models"
)

// GetTeacherBaseSalary returns the base salary configuration for a given teacher
func GetTeacherBaseSalary(db *sql.DB, userID string) (*models.TeacherBaseSalary, error) {
	query := `
		SELECT id, user_id, amount, period, created_at, updated_at
		FROM teacher_base_salaries
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var salary models.TeacherBaseSalary
	err := db.QueryRow(query, userID).Scan(
		&salary.ID,
		&salary.UserID,
		&salary.Amount,
		&salary.Period,
		&salary.CreatedAt,
		&salary.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &salary, nil
}

// GetTeacherAllowance returns the allowance configuration for a given teacher
func GetTeacherAllowance(db *sql.DB, userID string) (*models.TeacherAllowance, error) {
	query := `
		SELECT id, user_id, amount, period, is_active, created_at, updated_at
		FROM teacher_allowances
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var allowance models.TeacherAllowance
	err := db.QueryRow(query, userID).Scan(
		&allowance.ID,
		&allowance.UserID,
		&allowance.Amount,
		&allowance.Period,
		&allowance.IsActive,
		&allowance.CreatedAt,
		&allowance.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil if no allowance found
	}
	if err != nil {
		return nil, err
	}
	return &allowance, nil
}

// UpsertTeacherBaseSalary creates a new base salary record
func UpsertTeacherBaseSalary(db *sql.DB, salary *models.TeacherBaseSalary) error {
	query := `
		INSERT INTO teacher_base_salaries (user_id, amount, period, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		salary.UserID,
		salary.Amount,
		salary.Period,
	).Scan(&salary.ID, &salary.CreatedAt, &salary.UpdatedAt)
}

// UpsertTeacherAllowance creates a new allowance record
func UpsertTeacherAllowance(db *sql.DB, allowance *models.TeacherAllowance) error {
	query := `
		INSERT INTO teacher_allowances (user_id, amount, period, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		allowance.UserID,
		allowance.Amount,
		allowance.Period,
		allowance.IsActive,
	).Scan(&allowance.ID, &allowance.CreatedAt, &allowance.UpdatedAt)
}

// GetTeacherSalary (Legacy Compatibility) - Returns a merged view
func GetTeacherSalary(db *sql.DB, userID string) (*models.TeacherSalary, error) {
	base, err := GetTeacherBaseSalary(db, userID)
	if err != nil {
		return nil, err
	}

	allow, _ := GetTeacherAllowance(db, userID)

	salary := &models.TeacherSalary{
		ID:        base.ID,
		UserID:    userID,
		Amount:    base.Amount,
		Period:    base.Period,
		CreatedAt: base.CreatedAt,
	}

	if allow != nil && allow.IsActive {
		salary.HasAllowance = true
		salary.Allowance = allow.Amount
		salary.AllowancePeriod = allow.Period
	}

	return salary, nil
}
