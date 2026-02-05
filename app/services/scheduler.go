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

			// Trigger at 8:05 PM (20:05)
			if now.Hour() == 20 && now.Minute() == 5 {
				log.Println("Triggering scheduled tasks [20:05]...")

				// Generate Allowances
				if err := GenerateDailyAllowances(db); err != nil {
					log.Printf("Error generating daily allowances: %v", err)
				}
			}
		}
	}()
}
