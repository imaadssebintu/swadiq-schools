package services

import (
	"database/sql"
	"log"
	"time"
)

// StartScheduler starts the background task scheduler
func StartScheduler(db *sql.DB) {
	go func() {
		log.Println("Scheduler started...")
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			// Trigger at 8:14 PM (20:14)
			if now.Hour() == 20 && now.Minute() == 14 {
				log.Println("Triggering scheduled tasks [20:14]...")

				// Generate Allowances
				if err := GenerateDailyAllowances(db); err != nil {
					log.Printf("Error generating daily allowances: %v", err)
				}
			}
		}
	}()
}
