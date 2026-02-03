package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// GenerateDailyAllowances checks for teachers present today and creates allowance expenses
func GenerateDailyAllowances(db *sql.DB) error {
	log.Println("Starting daily allowance generation...")

	// 1. Get or Create "Allowance" Category
	var categoryID string
	err := db.QueryRow("SELECT id FROM categories WHERE name = 'Allowance'").Scan(&categoryID)
	if err == sql.ErrNoRows {
		log.Println("Allowance category not found, creating it...")
		err = db.QueryRow("INSERT INTO categories (name, is_active) VALUES ('Allowance', true) RETURNING id").Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("failed to create allowance category: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to fetch allowance category: %v", err)
	}

	// 2. Find eligible teachers (Present today AND have active allowance > 0)
	// We check for expenses created today for this category and teacher to avoid duplicates
	today := time.Now().Format("2006-01-02")

	query := `
		SELECT 
			u.id, 
			u.first_name || ' ' || u.last_name as name,
			ta.amount
		FROM teacher_attendances att
		JOIN teacher_allowances ta ON att.teacher_id = ta.user_id
		JOIN users u ON u.id = att.teacher_id
		WHERE att.date = $1 
		AND att.status = 'present'
		AND ta.is_active = true
		AND ta.amount > 0
		AND NOT EXISTS (
			SELECT 1 FROM expenses e 
			WHERE e.category_id = $2 
			AND e.date = $1 
			AND e.title LIKE 'Daily Allowance: ' || u.first_name || ' ' || u.last_name || '%'
		)
	`

	rows, err := db.Query(query, today, categoryID)
	if err != nil {
		return fmt.Errorf("failed to query eligible teachers: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var teacherID, teacherName string
		var amount int64
		if err := rows.Scan(&teacherID, &teacherName, &amount); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// 3. Create Expense Record
		title := fmt.Sprintf("Daily Allowance: %s", teacherName)
		notes := fmt.Sprintf("Auto-generated allowance for being on duty on %s", today)

		_, err := db.Exec(`
			INSERT INTO expenses (
				category_id, title, amount, currency, date, status, notes
			) VALUES (
				$1, $2, $3, 'UGX', $4, 'UNPAID', $5
			)
		`, categoryID, title, amount, today, notes)

		if err != nil {
			log.Printf("Failed to create allowance expense for %s: %v", teacherName, err)
		} else {
			count++
			log.Printf("Created allowance expense for %s: %d UGX", teacherName, amount)
		}
	}

	log.Printf("Daily allowance generation completed. Created %d records.", count)
	return nil
}
