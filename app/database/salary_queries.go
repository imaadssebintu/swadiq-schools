package database

import (
	"database/sql"

	"swadiq-schools/app/models"
)

// GetTeacherSalary returns the salary configuration for a given teacher
func GetTeacherSalary(db *sql.DB, userID string) (*models.TeacherSalary, error) {
	query := `
		SELECT id, user_id, amount, allowance, has_allowance, period, allowance_period, effective_date, allowance_trigger, created_at, updated_at
		FROM teacher_salaries
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var salary models.TeacherSalary
	err := db.QueryRow(query, userID).Scan(
		&salary.ID,
		&salary.UserID,
		&salary.Amount,
		&salary.Allowance,
		&salary.HasAllowance,
		&salary.Period,
		&salary.AllowancePeriod,
		&salary.EffectiveDate,
		&salary.AllowanceTrigger,
		&salary.CreatedAt,
		&salary.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &salary, nil
}

// UpsertTeacherSalary creates or updates a teacher's salary
// For simplicity and history tracking, we might want to just insert new records (effective date logic),
// but for this basic requirement, we'll check if one exists or insert new.
// Actually, a simple Insert is better for history, fetching the latest one.
// However, to keep it simple as "Set Salary", we will insert a new record which becomes the current one.
func UpsertTeacherSalary(db *sql.DB, salary *models.TeacherSalary) error {
	query := `
		INSERT INTO teacher_salaries (user_id, amount, allowance, has_allowance, period, allowance_period, effective_date, allowance_trigger, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		salary.UserID,
		salary.Amount,
		salary.Allowance,
		salary.HasAllowance,
		salary.Period,
		salary.AllowancePeriod,
		salary.EffectiveDate,
		salary.AllowanceTrigger,
	).Scan(&salary.ID, &salary.CreatedAt, &salary.UpdatedAt)
}
