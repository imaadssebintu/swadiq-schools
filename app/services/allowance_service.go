package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// GenerateDailyAllowances checks for teachers who conducted lessons today and creates allowance records
func GenerateDailyAllowances(db *sql.DB) error {
	log.Println("Starting daily allowance generation...")

	// 1. Find eligible teachers (Conducted a lesson today AND have active allowance > 0)
	// We check for records already created today in teacher_allowance_accruals to avoid duplicates
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT DISTINCT
			u.id, 
			u.first_name || ' ' || u.last_name as name,
			ta.amount
		FROM conducted_lessons cl
		JOIN teacher_allowances ta ON cl.teacher_id = ta.user_id
		JOIN users u ON u.id = cl.teacher_id
		WHERE cl.date = $1 
		AND ta.is_active = true
		AND ta.period = 'day'
		AND ta.amount > 0
		AND NOT EXISTS (
			SELECT 1 FROM teacher_allowance_accruals taa 
			WHERE taa.teacher_id = u.id 
			AND taa.date = $1
		)
	`

	// Diagnostic: Log total count of active daily allowances
	var activeCount int
	db.QueryRow("SELECT COUNT(*) FROM teacher_allowances WHERE is_active = true AND period = 'day'").Scan(&activeCount)
	log.Printf("Diagnostic: Found %d teachers with active daily allowances", activeCount)

	// Diagnostic: Log ALL active allowances
	diagRows, _ := db.Query("SELECT user_id, amount, period FROM teacher_allowances WHERE is_active = true")
	log.Println("Diagnostic: All active allowances:")
	for diagRows.Next() {
		var uid, prd string
		var amt int64
		diagRows.Scan(&uid, &amt, &prd)
		log.Printf("  - Teacher: %s, Amount: %d, Period: %s", uid, amt, prd)
	}
	diagRows.Close()

	rows, err := db.Query(query, today)
	if err != nil {
		return fmt.Errorf("failed to query eligible teachers: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("Diagnostic: No teachers found for date %s with period='day'.", today)
		var clCount int
		db.QueryRow("SELECT COUNT(*) FROM conducted_lessons WHERE date = $1", today).Scan(&clCount)
		log.Printf("Diagnostic: Total conducted lessons found for today (%s): %d", today, clCount)

		if clCount > 0 {
			// List teachers who conducted lessons
			clRows, _ := db.Query("SELECT DISTINCT teacher_id FROM conducted_lessons WHERE date = $1", today)
			log.Println("Diagnostic: Teachers who conducted lessons today:")
			for clRows.Next() {
				var tid string
				clRows.Scan(&tid)
				log.Printf("  - Teacher ID: %s", tid)
			}
			clRows.Close()
		}
	} else {
		// Reset rows for the loop
		rows.Close()
		rows, _ = db.Query(query, today)
		defer rows.Close()
	}

	count := 0
	for rows.Next() {
		var teacherID, teacherName string
		var amount int64
		if err := rows.Scan(&teacherID, &teacherName, &amount); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// 2. Create Accrual Record
		notes := fmt.Sprintf("Auto-generated allowance for lessons conducted on %s", today)

		_, err := db.Exec(`
			INSERT INTO teacher_allowance_accruals (
				teacher_id, amount, date, status, notes
			) VALUES (
				$1, $2, $3, 'unpaid', $4
			)
		`, teacherID, amount, today, notes)

		if err != nil {
			log.Printf("Failed to create allowance accrual for %s: %v", teacherName, err)
		} else {
			count++
			log.Printf("Created allowance accrual for %s: %d UGX", teacherName, amount)
		}
	}

	log.Printf("Daily allowance generation completed. Created %d records.", count)
	return nil
}
