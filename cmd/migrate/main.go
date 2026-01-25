package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"swadiq-schools/app/config"
)

func main() {
	log.Println("Starting manual migration for payments...")

	config.InitDB()
	db := config.GetDB()
	if db == nil {
		log.Fatal("Failed to get database instance")
	}
	defer db.Close()

	// Read and execute the SQL file directly
	executeSQLFile(db, "schema_update_allowance_trigger.sql")

	// Also ensure teacher_attendance schema is applied if not already
	executeSQLFile(db, "schema_update_teacher_attendance.sql")

	log.Println("Manual migration completed successfully!")
}

func executeSQLFile(db *sql.DB, filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Skipping %s: %v", filePath, err)
		return
	}

	log.Printf("Executing %s...", filePath)
	if _, err := db.Exec(string(content)); err != nil {
		log.Printf("Error executing %s: %v", filePath, err)
	} else {
		log.Printf("Successfully executed %s", filePath)
	}
}
