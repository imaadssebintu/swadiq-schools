package database

import (
	"database/sql"
	"log"
)

// RunMigrations checks and applies necessary schema updates
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// 1. Add allowance column to teacher_salaries if not exists
	err := addAllowanceColumn(db)
	if err != nil {
		return err
	}

	// 2. Add has_allowance column to teacher_salaries if not exists
	err = addHasAllowanceColumn(db)
	if err != nil {
		return err
	}

	// 3. Add allowance_period column to teacher_salaries if not exists
	err = addAllowancePeriodColumn(db)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func addAllowanceColumn(db *sql.DB) error {
	query := `
		DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'teacher_salaries' 
				AND column_name = 'allowance'
			) THEN 
				ALTER TABLE teacher_salaries ADD COLUMN allowance BIGINT NOT NULL DEFAULT 0;
				RAISE NOTICE 'Added allowance column to teacher_salaries';
			END IF; 
		END $$;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to run migration for allowance column: %v", err)
		return err
	}
	return nil
}

func addHasAllowanceColumn(db *sql.DB) error {
	query := `
		DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'teacher_salaries' 
				AND column_name = 'has_allowance'
			) THEN 
				ALTER TABLE teacher_salaries ADD COLUMN has_allowance BOOLEAN NOT NULL DEFAULT false;
				RAISE NOTICE 'Added has_allowance column to teacher_salaries';
			END IF; 
		END $$;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to run migration for has_allowance column: %v", err)
		return err
	}
	return nil
}

func addAllowancePeriodColumn(db *sql.DB) error {
	query := `
		DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'teacher_salaries' 
				AND column_name = 'allowance_period'
			) THEN 
				ALTER TABLE teacher_salaries ADD COLUMN allowance_period VARCHAR(20) NOT NULL DEFAULT 'month';
				RAISE NOTICE 'Added allowance_period column to teacher_salaries';
			END IF; 
		END $$;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to run migration for allowance_period column: %v", err)
		return err
	}
	return nil
}
