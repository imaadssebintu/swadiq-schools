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
	weekday := time.Now().Weekday().String()
	query := `
		WITH scheduled_counts AS (
			SELECT teacher_id, COUNT(*) as scheduled_count
			FROM timetable_entries
			WHERE day_of_week = $2
			GROUP BY teacher_id
		),
		conducted_counts AS (
			SELECT teacher_id, COUNT(*) as conducted_count
			FROM conducted_lessons
			WHERE date = $1
			GROUP BY teacher_id
		)
		SELECT
			u.id, 
			u.first_name || ' ' || u.last_name as name,
			ta.amount
		FROM teacher_allowances ta
		JOIN users u ON u.id = ta.user_id
		LEFT JOIN teacher_attendances att ON u.id = att.teacher_id AND att.date = $1 AND att.status = 'present'
		LEFT JOIN scheduled_counts sc ON u.id = sc.teacher_id
		LEFT JOIN conducted_counts cc ON u.id = cc.teacher_id
		WHERE ta.is_active = true
		AND ta.period = 'day'
		AND ta.amount > 0
		AND u.is_active = true
		AND (
			att.id IS NOT NULL 
			OR (sc.scheduled_count IS NOT NULL AND cc.conducted_count IS NOT NULL AND cc.conducted_count >= sc.scheduled_count)
			OR (sc.scheduled_count IS NULL AND cc.conducted_count IS NOT NULL AND cc.conducted_count > 0)
		)
		AND NOT EXISTS (
			SELECT 1 FROM teacher_allowance_accruals taa 
			WHERE taa.teacher_id = u.id 
			AND taa.date = $1
		)
		GROUP BY u.id, name, ta.amount
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

	rows, err := db.Query(query, today, weekday)
	if err != nil {
		return fmt.Errorf("failed to query eligible teachers: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("Diagnostic: No teachers found for date %s (weekday: %s).", today, weekday)

		var attCount int
		db.QueryRow("SELECT COUNT(*) FROM teacher_attendances WHERE date = $1 AND status = 'present'", today).Scan(&attCount)
		log.Printf("Diagnostic: Total teachers marked 'present' today: %d", attCount)

		var clCount int
		db.QueryRow("SELECT COUNT(*) FROM conducted_lessons WHERE date = $1", today).Scan(&clCount)
		log.Printf("Diagnostic: Total conducted lessons found for today: %d", clCount)

		if clCount > 0 {
			// List teachers who conducted lessons and their counts vs scheduled
			clRows, _ := db.Query(`
				SELECT 
					u.first_name || ' ' || u.last_name, 
					cl.teacher_id, 
					COUNT(cl.id) as conducted,
					COALESCE((SELECT COUNT(*) FROM timetable_entries WHERE teacher_id = cl.teacher_id AND day = $2), 0) as scheduled
				FROM conducted_lessons cl
				JOIN users u ON cl.teacher_id = u.id
				WHERE cl.date = $1
				GROUP BY u.first_name, u.last_name, cl.teacher_id
			`, today, weekday)
			log.Println("Diagnostic: Lessons conducted breakdown:")
			for clRows.Next() {
				var name, tid string
				var cond, sched int
				clRows.Scan(&name, &tid, &cond, &sched)
				log.Printf("  - Teacher %s (%s): Conducted %d / Scheduled %d", name, tid, cond, sched)
			}
			clRows.Close()
		}
	} else {
		// Reset rows for the loop
		rows.Close()
		rows, _ = db.Query(query, today, weekday)
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
		notes := fmt.Sprintf("Auto-generated allowance (lessons/attendance) for %s", today)

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
