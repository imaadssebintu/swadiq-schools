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

			// Trigger at 7:45 PM (19:45)
			if now.Hour() == 19 && now.Minute() == 45 {
				log.Println("Triggering scheduled tasks [19:45]...")

				// Generate Allowances
				if err := GenerateDailyAllowances(db); err != nil {
					log.Printf("Error generating daily allowances: %v", err)
				}
			}
		}
	}()
}
