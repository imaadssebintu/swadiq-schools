package database

import (
	"database/sql"
	"math"
	"swadiq-schools/app/models"
	"time"
)

// GetTeacherDutyDays counts the number of days a teacher was present within a date range
func GetTeacherDutyDays(db *sql.DB, teacherID string, startDate, endDate time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM teacher_attendance 
	          WHERE teacher_id = $1 
	          AND date >= $2 AND date <= $3 
	          AND status = 'present'`

	var count int
	err := db.QueryRow(query, teacherID, startDate, endDate).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CalculateTeacherPeriodPay computes the payout breakdown for a teacher
// This logic encapsulates the business rules:
// - Daily Salary: Fixed amount * Duty Days (if applicable) or just Fixed Amount? Usually Daily Salary means Base * Days.
// - Allowance:
//   - Daily Allowance: Amount * Duty Days.
//   - Weekly Allowance: Amount * Weeks (Approx).
func CalculateTeacherPeriodPay(salary *models.TeacherSalary, dutyDays int, weeks float64) (float64, float64, float64) {
	if salary == nil {
		return 0, 0, 0
	}

	basePay := float64(salary.Amount)
	// If base salary is 'day', then base pay is Rate * DutyDays
	if salary.Period == "day" {
		basePay = float64(salary.Amount) * float64(dutyDays)
	}

	allowancePay := 0.0
	if salary.HasAllowance {
		if salary.AllowancePeriod == "day" {
			allowancePay = float64(salary.Allowance) * float64(dutyDays)
		} else if salary.AllowancePeriod == "week" {
			allowancePay = float64(salary.Allowance) * weeks
		} else if salary.AllowancePeriod == "month" {
			allowancePay = float64(salary.Allowance) // Fixed monthly
		}
	}

	totalPay := basePay + allowancePay
	return basePay, allowancePay, totalPay
}

// GetProposedPayout calculates the payout for a specific period based on attendance
func GetProposedPayout(db *sql.DB, teacherID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	salary, err := GetTeacherSalary(db, teacherID)
	if err != nil {
		return nil, err
	}

	dutyDays, err := GetTeacherDutyDays(db, teacherID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate weeks roughly
	days := endDate.Sub(startDate).Hours() / 24
	weeks := days / 7.0
	if weeks < 1 {
		weeks = 1 // Minimum 1 week if < 7 days? Or 0? Let's use strict ratio.
		weeks = math.Max(days/7.0, 0)
	}

	base, allowance, total := CalculateTeacherPeriodPay(salary, dutyDays, weeks)

	return map[string]interface{}{
		"teacher_id":    teacherID,
		"period_start":  startDate,
		"period_end":    endDate,
		"duty_days":     dutyDays,
		"salary_config": salary,
		"base_pay":      base,
		"allowance_pay": allowance,
		"total_pay":     total,
	}, nil
}

// GetTotalPaid retrieves the total amount paid to a teacher within a period
func GetTotalPaid(db *sql.DB, teacherID string, startDate, endDate time.Time) (int64, int64, int64, error) {
	query := `SELECT 
		COALESCE(SUM(CASE WHEN type = 'base_salary' THEN amount ELSE 0 END), 0) as base_paid,
		COALESCE(SUM(CASE WHEN type = 'allowance' THEN amount ELSE 0 END), 0) as allowance_paid,
		COALESCE(SUM(amount), 0) as total_paid
		FROM teacher_payments 
		WHERE teacher_id = $1 AND paid_at >= $2 AND paid_at <= $3`

	var basePaid, allowPaid, totalPaid int64
	err := db.QueryRow(query, teacherID, startDate, endDate).Scan(&basePaid, &allowPaid, &totalPaid)
	if err != nil {
		return 0, 0, 0, err
	}
	return basePaid, allowPaid, totalPaid, nil
}

// CalculateAccruedSalary calculates what a teacher has earned based on attendance and salary config
func CalculateAccruedSalary(salary *models.TeacherSalary, dutyDays int, weeks float64) (int64, int64, int64) {
	if salary == nil {
		return 0, 0, 0
	}

	var baseAccrued, allowAccrued float64

	// Base Salary Logic
	if salary.Period == "day" {
		baseAccrued = float64(salary.Amount) * float64(dutyDays)
	} else if salary.Period == "week" {
		baseAccrued = float64(salary.Amount) * weeks
	} else {
		// Monthly - assume fixed for the period (simplified, or pro-rated if implemented)
		// For simplicity, showing full amount if period covers a month, or just fixed rate
		baseAccrued = float64(salary.Amount)
	}

	// Allowance Logic
	if salary.HasAllowance {
		if salary.AllowancePeriod == "day" {
			allowAccrued = float64(salary.Allowance) * float64(dutyDays)
		} else if salary.AllowancePeriod == "week" {
			allowAccrued = float64(salary.Allowance) * weeks
		} else {
			allowAccrued = float64(salary.Allowance)
		}
	}

	return int64(baseAccrued), int64(allowAccrued), int64(baseAccrued + allowAccrued)
}

// GetTeacherPayrollStatus returns the full financial status (Accrued, Paid, Unpaid)
func GetTeacherPayrollStatus(db *sql.DB, teacherID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	salary, err := GetTeacherSalary(db, teacherID)
	if err != nil {
		return nil, err
	}

	dutyDays, err := GetTeacherDutyDays(db, teacherID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate weeks (simplified)
	days := endDate.Sub(startDate).Hours() / 24
	weeks := math.Max(days/7.0, 0)
	if weeks < 1 && days > 0 {
		weeks = 1
	}

	baseAccrued, allowAccrued, totalAccrued := CalculateAccruedSalary(salary, dutyDays, weeks)
	basePaid, allowPaid, totalPaid, err := GetTotalPaid(db, teacherID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"accrued": map[string]int64{
			"base":      baseAccrued,
			"allowance": allowAccrued,
			"total":     totalAccrued,
		},
		"paid": map[string]int64{
			"base":      basePaid,
			"allowance": allowPaid,
			"total":     totalPaid,
		},
		"unpaid": map[string]int64{
			"base":      baseAccrued - basePaid,
			"allowance": allowAccrued - allowPaid,
			"total":     totalAccrued - totalPaid,
		},
		"duty_days": dutyDays,
	}, nil
}
