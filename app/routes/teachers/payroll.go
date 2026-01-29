package teachers

import (
	"database/sql"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GeneratePayrollAPI creates pending payment records for all teachers based on their salary period
func GeneratePayrollAPI(c *fiber.Ctx) error {
	db := config.GetDB()

	// Get period type from query (day, week, month)
	periodType := c.Query("period", "month")
	if periodType != "day" && periodType != "week" && periodType != "month" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid period type. Must be 'day', 'week', or 'month'",
		})
	}

	// Calculate period dates
	now := time.Now()
	var periodStart, periodEnd time.Time

	switch periodType {
	case "day":
		periodStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		periodEnd = periodStart.Add(24*time.Hour - time.Nanosecond)
	case "week":
		// Start of current week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		periodStart = now.AddDate(0, 0, -(weekday - 1))
		periodStart = time.Date(periodStart.Year(), periodStart.Month(), periodStart.Day(), 0, 0, 0, 0, periodStart.Location())
		periodEnd = periodStart.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case "month":
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}

	// Get all active teachers
	teachers, err := database.GetAllTeachers(db)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch teachers: " + err.Error(),
		})
	}

	generated := 0
	skipped := 0
	errors := []string{}

	for _, teacher := range teachers {
		// Get teacher's salary configuration
		baseSalary, err := database.GetTeacherBaseSalary(db, teacher.ID)
		if err != nil {
			skipped++
			continue // Skip teachers without salary config
		}

		// Check if teacher's salary period matches the generation period
		if string(baseSalary.Period) != periodType {
			skipped++
			continue
		}

		// Check if payment already exists for this period
		exists, err := checkPaymentExists(db, teacher.ID, periodStart, periodEnd)
		if err != nil {
			errors = append(errors, "Teacher "+teacher.FirstName+" "+teacher.LastName+": "+err.Error())
			continue
		}
		if exists {
			skipped++
			continue
		}

		// Calculate accrued amount
		dutyDays, err := database.GetTeacherDutyDays(db, teacher.ID, periodStart, periodEnd)
		if err != nil {
			errors = append(errors, "Teacher "+teacher.FirstName+" "+teacher.LastName+": "+err.Error())
			continue
		}

		days := periodEnd.Sub(periodStart).Hours() / 24
		weeks := days / 7.0

		allowance, _ := database.GetTeacherAllowance(db, teacher.ID)

		// Build legacy salary struct for calculation
		salary := &models.TeacherSalary{
			Amount: baseSalary.Amount,
			Period: baseSalary.Period,
		}
		if allowance != nil && allowance.IsActive {
			salary.HasAllowance = true
			salary.Allowance = allowance.Amount
			salary.AllowancePeriod = allowance.Period
		}

		baseAccrued, allowAccrued, totalAccrued := database.CalculateAccruedSalary(salary, dutyDays, weeks)

		// Create pending payment records
		if baseAccrued > 0 {
			err = createPendingPayment(db, teacher.ID, baseAccrued, "base_salary", periodStart, periodEnd)
			if err != nil {
				errors = append(errors, "Teacher "+teacher.FirstName+" "+teacher.LastName+" (base): "+err.Error())
				continue
			}
		}

		if allowAccrued > 0 {
			err = createPendingPayment(db, teacher.ID, allowAccrued, "allowance", periodStart, periodEnd)
			if err != nil {
				errors = append(errors, "Teacher "+teacher.FirstName+" "+teacher.LastName+" (allowance): "+err.Error())
				continue
			}
		}

		if totalAccrued > 0 {
			generated++
		}
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"generated": generated,
		"skipped":   skipped,
		"errors":    errors,
		"period": fiber.Map{
			"type":  periodType,
			"start": periodStart.Format("2006-01-02"),
			"end":   periodEnd.Format("2006-01-02"),
		},
	})
}

func checkPaymentExists(db *sql.DB, teacherID string, periodStart, periodEnd time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM teacher_payments 
	          WHERE teacher_id = $1 
	          AND period_start = $2 
	          AND period_end = $3`

	var count int
	err := db.QueryRow(query, teacherID, periodStart, periodEnd).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func createPendingPayment(db *sql.DB, teacherID string, amount int64, paymentType string, periodStart, periodEnd time.Time) error {
	query := `INSERT INTO teacher_payments 
	          (teacher_id, amount, type, period_start, period_end, status, paid_at) 
	          VALUES ($1, $2, $3, $4, $5, 'pending', NULL)`

	_, err := db.Exec(query, teacherID, amount, paymentType, periodStart, periodEnd)
	return err
}
