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
	baseSalary, err := GetTeacherBaseSalary(db, teacherID)
	if err != nil {
		return nil, err
	}

	allowance, err := GetTeacherAllowance(db, teacherID)
	// err handled implicitly below

	dutyDays, err := GetTeacherDutyDays(db, teacherID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate weeks roughly
	days := endDate.Sub(startDate).Hours() / 24
	weeks := math.Max(days/7.0, 0)

	// Combine into legacy models.TeacherSalary for CalculateTeacherPeriodPay if compatible,
	// OR better: handle independently.
	salary := &models.TeacherSalary{
		Amount: baseSalary.Amount,
		Period: baseSalary.Period,
	}
	if allowance != nil && allowance.IsActive {
		salary.HasAllowance = true
		salary.Allowance = allowance.Amount
		salary.AllowancePeriod = allowance.Period
	}

	base, allow, total := CalculateTeacherPeriodPay(salary, dutyDays, weeks)

	return map[string]interface{}{
		"teacher_id":    teacherID,
		"period_start":  startDate,
		"period_end":    endDate,
		"duty_days":     dutyDays,
		"base_pay":      base,
		"allowance_pay": allow,
		"total_pay":     total,
	}, nil
}

// GetTotalPaid retrieves the total amount paid to a teacher within a period
func GetTotalPaid(db *sql.DB, teacherID string, startDate, endDate time.Time) (int64, int64, int64, error) {
	query := `SELECT 
		COALESCE(SUM(CASE WHEN type = 'base_salary' THEN amount ELSE 0 END), 0) + 
		COALESCE(SUM(CASE WHEN type = 'combined' THEN amount ELSE 0 END), 0) as base_paid,
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
	baseSalary, err := GetTeacherBaseSalary(db, teacherID)
	if err != nil {
		return nil, err
	}
	allowance, _ := GetTeacherAllowance(db, teacherID)

	dutyDays, err := GetTeacherDutyDays(db, teacherID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	days := endDate.Sub(startDate).Hours() / 24
	weeks := math.Max(days/7.0, 0)
	if weeks < 1 && days > 0 {
		weeks = 1
	}

	// Legacy wrapper for calculation
	salary := &models.TeacherSalary{
		Amount: baseSalary.Amount,
		Period: baseSalary.Period,
	}
	if allowance != nil && allowance.IsActive {
		salary.HasAllowance = true
		salary.Allowance = allowance.Amount
		salary.AllowancePeriod = allowance.Period
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

// GetTeacherLedger generates a period-based financial history for a teacher
// It looks back N months and calculates what was owed vs what was paid for each month.
func GetTeacherLedger(db *sql.DB, teacherID string, monthsToLookBack int) ([]map[string]interface{}, error) {
	var ledger []map[string]interface{}

	// Get current salary config via merged wrapper
	salary, err := GetTeacherSalary(db, teacherID)
	if err != nil {
		// No salary configured yet - return empty ledger
		return ledger, nil
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	// Iterate back N months (including current)
	for i := 0; i < monthsToLookBack; i++ {
		// Calculate the start of the month we are looking at
		targetTime := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).AddDate(0, -i, 0)
		startOfMonth := targetTime
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		// Cutoff Check: Don't show requirements from before the salary was actually set.
		// Use salary.CreatedAt month as the starting point.
		configStart := time.Date(salary.CreatedAt.Year(), salary.CreatedAt.Month(), 1, 0, 0, 0, 0, salary.CreatedAt.Location())
		if startOfMonth.Before(configStart) {
			// Check if there's any payment in this month anyway; if so, show it, else skip ghost requirement.
			_, _, pTotal, _ := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)
			if pTotal == 0 {
				continue
			}
		}

		dutyDays, err := GetTeacherDutyDays(db, teacherID, startOfMonth, endOfMonth)
		if err != nil {
			continue // Skip or error? Skip safe.
		}

		// Calculate what should have been paid
		// Calculate weeks in this specific month
		daysInMonth := endOfMonth.Day() // e.g. 31
		weeks := float64(daysInMonth) / 7.0

		// Use Accrued Salary Logic
		// Note: CalculateAccruedSalary takes weeks as float.
		baseAccrued, allowAccrued, totalAccrued := CalculateAccruedSalary(salary, dutyDays, weeks)

		// Get Actual Payments in this window
		basePaid, allowPaid, totalPaid, err := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)
		if err != nil {
			continue
		}

		// Determine Status
		balance := totalAccrued - totalPaid
		status := "Paid"
		if balance > 0 {
			if totalPaid > 0 {
				status = "Partial"
			} else {
				status = "Unpaid"
			}
		} else if balance < 0 {
			status = "Overpaid"
		} else {
			// Exact match, check if 0
			if totalAccrued == 0 {
				status = "No Activity"
			}
		}

		entry := map[string]interface{}{
			"period_name": startOfMonth.Format("January 2006"),
			"start_date":  startOfMonth.Format("2006-01-02"),
			"end_date":    endOfMonth.Format("2006-01-02"),
			"due_date":    endOfMonth.Format("2006-01-02"), // Payday = Last day
			"duty_days":   dutyDays,
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
			"balance": balance,
			"status":  status,
		}

		ledger = append(ledger, entry)
	}

	return ledger, nil
}

// GetTeacherBaseSalaryLedger returns only base salary ledger entries
func GetTeacherBaseSalaryLedger(db *sql.DB, teacherID string, monthsToLookBack int) ([]map[string]interface{}, error) {
	var ledger []map[string]interface{}

	baseSalary, err := GetTeacherBaseSalary(db, teacherID)
	if err != nil {
		// No salary configured yet - return empty ledger
		return ledger, nil
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	for i := 0; i < monthsToLookBack; i++ {
		targetTime := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).AddDate(0, -i, 0)
		startOfMonth := targetTime
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		// Cutoff Check
		configStart := time.Date(baseSalary.CreatedAt.Year(), baseSalary.CreatedAt.Month(), 1, 0, 0, 0, 0, baseSalary.CreatedAt.Location())
		if startOfMonth.Before(configStart) {
			_, _, pTotal, _ := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)
			if pTotal == 0 {
				continue
			}
		}

		dutyDays, err := GetTeacherDutyDays(db, teacherID, startOfMonth, endOfMonth)
		if err != nil {
			continue
		}

		days := endOfMonth.Sub(startOfMonth).Hours() / 24
		weeks := days / 7.0

		// Calculate base salary accrual
		var baseAccrued int64
		if baseSalary.Period == "day" {
			baseAccrued = baseSalary.Amount * int64(dutyDays)
		} else if baseSalary.Period == "week" {
			baseAccrued = int64(float64(baseSalary.Amount) * weeks)
		} else {
			baseAccrued = baseSalary.Amount
		}

		basePaid, _, _, _ := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)

		entry := map[string]interface{}{
			"period_name": startOfMonth.Format("January 2006"),
			"start_date":  startOfMonth.Format("2006-01-02"),
			"end_date":    endOfMonth.Format("2006-01-02"),
			"accrued":     baseAccrued,
			"paid":        basePaid,
		}

		ledger = append(ledger, entry)
	}

	return ledger, nil
}

// GetTeacherAllowanceLedger returns only allowance ledger entries
func GetTeacherAllowanceLedger(db *sql.DB, teacherID string, monthsToLookBack int) ([]map[string]interface{}, error) {
	var ledger []map[string]interface{}

	allowance, err := GetTeacherAllowance(db, teacherID)
	if err != nil || allowance == nil || !allowance.IsActive {
		return ledger, nil // Return empty if no active allowance
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	for i := 0; i < monthsToLookBack; i++ {
		targetTime := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation).AddDate(0, -i, 0)
		startOfMonth := targetTime
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		// Cutoff Check
		configStart := time.Date(allowance.CreatedAt.Year(), allowance.CreatedAt.Month(), 1, 0, 0, 0, 0, allowance.CreatedAt.Location())
		if startOfMonth.Before(configStart) {
			_, allowPaid, _, _ := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)
			if allowPaid == 0 {
				continue
			}
		}

		dutyDays, err := GetTeacherDutyDays(db, teacherID, startOfMonth, endOfMonth)
		if err != nil {
			continue
		}

		days := endOfMonth.Sub(startOfMonth).Hours() / 24
		weeks := days / 7.0

		// Calculate allowance accrual
		var allowAccrued int64
		if allowance.Period == "day" {
			allowAccrued = allowance.Amount * int64(dutyDays)
		} else if allowance.Period == "week" {
			allowAccrued = int64(float64(allowance.Amount) * weeks)
		} else {
			allowAccrued = allowance.Amount
		}

		_, allowPaid, _, _ := GetTotalPaid(db, teacherID, startOfMonth, endOfMonth)

		entry := map[string]interface{}{
			"period_name": startOfMonth.Format("January 2006"),
			"start_date":  startOfMonth.Format("2006-01-02"),
			"end_date":    endOfMonth.Format("2006-01-02"),
			"accrued":     allowAccrued,
			"paid":        allowPaid,
		}

		ledger = append(ledger, entry)
	}

	return ledger, nil
}
