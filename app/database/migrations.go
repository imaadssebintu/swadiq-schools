package database

import (
	"database/sql"
	"log"
)

// RunMigrations checks and applies necessary schema updates
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// 1. Permanently delete the legacy tables you requested
	err := dropLegacyTables(db)
	if err != nil {
		return err
	}

	// 2. Ensure the new simplified tables exist
	err = createSimplifiedPayrollTables(db)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func WipeAllPayrollData(db *sql.DB) error {
	log.Println("Wiping all payroll data (Fresh Start)...")
	queries := []string{
		"TRUNCATE TABLE teacher_payments CASCADE",
		"TRUNCATE TABLE teacher_base_salaries CASCADE",
		"TRUNCATE TABLE teacher_allowances CASCADE",
	}
	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Printf("Warning: Failed to wipe table: %v", err)
		}
	}
	return nil
}

func dropLegacyTables(db *sql.DB) error {
	queries := []string{
		"DROP TABLE IF EXISTS teacher_salaries CASCADE",
		"DROP TABLE IF EXISTS legacy_teacher_salaries CASCADE",
		"DROP TABLE IF EXISTS legacy_teacher_salaries_backup CASCADE",
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Printf("Note: Failed to drop legacy table: %v", err)
		}
	}
	return nil
}

func createSimplifiedPayrollTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS teacher_base_salaries (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			amount BIGINT NOT NULL,
			period VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS teacher_allowances (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			amount BIGINT DEFAULT 0,
			period VARCHAR(20) DEFAULT 'month',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS teacher_payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			teacher_id UUID NOT NULL REFERENCES users(id),
			amount BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL,
			period_start DATE NOT NULL,
			period_end DATE NOT NULL,
			paid_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			reference VARCHAR(100),
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Printf("Failed to create payroll table: %v", err)
			return err
		}
	}
	return nil
}
