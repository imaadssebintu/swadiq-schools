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

			// Trigger at 7:20 PM (19:20)
			if now.Hour() == 19 && now.Minute() == 20 {
				log.Println("Triggering scheduled tasks [19:20]...")

				// Generate Allowances
				if err := GenerateDailyAllowances(db); err != nil {
					log.Printf("Error generating daily allowances: %v", err)
				}
			}
		}
	}()
}
