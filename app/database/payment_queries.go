package database

import (
	"database/sql"
	"fmt"
	"swadiq-schools/app/models"
)

// CreateTeacherPayment records a payment and creates a corresponding expense entry in a transaction
func CreateTeacherPayment(db *sql.DB, payment *models.TeacherPayment, teacherName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Payment Record
	queryPayment := `INSERT INTO teacher_payments (teacher_id, amount, type, period_start, period_end, reference, notes, paid_at) 
	                 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
					 RETURNING id, paid_at`
	err = tx.QueryRow(queryPayment,
		payment.TeacherID,
		payment.Amount,
		string(payment.Type),
		payment.PeriodStart,
		payment.PeriodEnd,
		payment.Reference,
		payment.Notes,
	).Scan(&payment.ID, &payment.PaidAt)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %v", err)
	}

	// 2. Handle Expense Integration
	var categoryID string
	err = tx.QueryRow("SELECT id FROM categories WHERE name = 'Salaries' AND deleted_at IS NULL").Scan(&categoryID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow("INSERT INTO categories (name, is_active) VALUES ('Salaries', true) RETURNING id").Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("failed to create category: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to find category: %v", err)
	}

	title := fmt.Sprintf("Salary Payout: %s", teacherName)
	if payment.Type == models.PaymentTypeAllowance {
		title = fmt.Sprintf("Allowance Payout: %s", teacherName)
	} else if payment.Type == models.PaymentTypeCombined {
		title = fmt.Sprintf("Full Salary Payout: %s", teacherName)
	}

	notes := fmt.Sprintf("System generated expense for teacher payroll disbursement. Period: %s to %s",
		payment.PeriodStart.Format("2006-01-02"), payment.PeriodEnd.Format("2006-01-02"))

	queryExpense := `INSERT INTO expenses (category_id, title, amount, currency, date, period_start, period_end, due_date, notes) 
	                 VALUES ($1, $2, $3, 'UGX', NOW(), $4, $5, $6, $7)`
	_, err = tx.Exec(queryExpense, categoryID, title, float64(payment.Amount),
		payment.PeriodStart, payment.PeriodEnd, payment.PeriodEnd, notes)
	if err != nil {
		return fmt.Errorf("failed to create expense: %v", err)
	}

	return tx.Commit()
}

// GetTeacherPayments retrieves all payments for a specific teacher
func GetTeacherPayments(db *sql.DB, teacherID string) ([]*models.TeacherPayment, error) {
	query := `SELECT id, teacher_id, amount, type, period_start, period_end, paid_at, reference, notes 
	          FROM teacher_payments 
			  WHERE teacher_id = $1 
			  ORDER BY paid_at DESC`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.TeacherPayment
	for rows.Next() {
		p := &models.TeacherPayment{}
		var pType string
		err := rows.Scan(
			&p.ID, &p.TeacherID, &p.Amount, &pType,
			&p.PeriodStart, &p.PeriodEnd, &p.PaidAt,
			&p.Reference, &p.Notes,
		)
		if err != nil {
			continue
		}
		p.Type = models.PaymentType(pType)
		payments = append(payments, p)
	}

	return payments, nil
}
