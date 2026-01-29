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

	// 3. Add term_id to attendance
	err = addAttendanceTermIDColumn(db)
	if err != nil {
		return err
	}

	// 4. Create conducted_lessons table
	err = createConductedLessonsTable(db)
	if err != nil {
		return err
	}

	// 5. Update teacher_payments schema
	err = updatePaymentTableSchema(db)
	if err != nil {
		return err
	}

	// 6. Refine assessment types
	err = refineAssessmentTypes(db)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func updatePaymentTableSchema(db *sql.DB) error {
	// Add status column if not exists
	_, err := db.Exec(`ALTER TABLE teacher_payments ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'completed'`)
	if err != nil {
		return err
	}

	// Make paid_at nullable
	_, err = db.Exec(`ALTER TABLE teacher_payments ALTER COLUMN paid_at DROP NOT NULL`)
	return err
}

func addAttendanceTermIDColumn(db *sql.DB) error {
	query := `ALTER TABLE attendance ADD COLUMN IF NOT EXISTS term_id UUID REFERENCES terms(id)`
	_, err := db.Exec(query)
	return err
}

func createConductedLessonsTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS conducted_lessons (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		timetable_entry_id UUID NOT NULL REFERENCES timetable_entries(id),
		term_id UUID REFERENCES terms(id),
		date DATE NOT NULL,
		teacher_id UUID NOT NULL REFERENCES users(id),
		topic TEXT,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(timetable_entry_id, date)
	)`
	_, err := db.Exec(query)
	return err
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

func refineAssessmentTypes(db *sql.DB) error {
	log.Println("Refining assessment types...")

	// 1. Add columns to assessment_types
	queries := []string{
		"ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES assessment_types(id)",
		"ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS category VARCHAR(50)",
		"ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS weight NUMERIC(5,2) DEFAULT 1.0",
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Warning: Migration step failed: %v", err)
		}
	}

	// 2. Add column to exams
	_, err := db.Exec("ALTER TABLE exams ADD COLUMN IF NOT EXISTS assessment_type_id UUID REFERENCES assessment_types(id)")
	if err != nil {
		log.Printf("Warning: Migration step failed: %v", err)
	}

	// 3. Seed Standard Assessment Types
	seedQueries := []string{
		`INSERT INTO assessment_types (name, code, category, color, is_active)
		 VALUES ('Beginning of Term', 'BOT', 'Exam', 'indigo', true)
		 ON CONFLICT (code) DO UPDATE SET category = 'Exam'`,
		`INSERT INTO assessment_types (name, code, category, color, is_active)
		 VALUES ('Mid Term', 'MT', 'Exam', 'purple', true)
		 ON CONFLICT (code) DO UPDATE SET category = 'Exam'`,
		`INSERT INTO assessment_types (name, code, category, color, is_active)
		 VALUES ('End of Term', 'EOT', 'Exam', 'pink', true)
		 ON CONFLICT (code) DO UPDATE SET category = 'Exam'`,
		`INSERT INTO assessment_types (name, code, category, color, is_active)
		 VALUES ('Class Test', 'CTEST', 'Test', 'amber', true)
		 ON CONFLICT (code) DO UPDATE SET category = EXCLUDED.category`,
		`INSERT INTO assessment_types (name, code, category, color, is_active)
		 VALUES ('Course Project', 'CPROJ', 'Project', 'emerald', true)
		 ON CONFLICT (code) DO UPDATE SET category = EXCLUDED.category`,
	}

	for _, q := range seedQueries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Warning: Seeding assessment types failed: %v", err)
		}
	}

	// 4. Seed Sub-types (Sample)
	subTypeLogic := `
	DO $$ 
	DECLARE 
		ctest_id UUID;
		cproj_id UUID;
	BEGIN
		SELECT id INTO ctest_id FROM assessment_types WHERE code = 'CTEST';
		SELECT id INTO cproj_id FROM assessment_types WHERE code = 'CPROJ';

		IF ctest_id IS NOT NULL THEN
			INSERT INTO assessment_types (name, code, parent_id, category, color, is_active)
			VALUES 
			('Test 1', 'T1', ctest_id, 'Test', 'amber', true),
			('Test 2', 'T2', ctest_id, 'Test', 'amber', true)
			ON CONFLICT (code) DO NOTHING;
		END IF;

		IF cproj_id IS NOT NULL THEN
			INSERT INTO assessment_types (name, code, parent_id, category, color, is_active)
			VALUES 
			('Project 1', 'P1', cproj_id, 'Project', 'emerald', true),
			('Lab Report', 'LR1', cproj_id, 'Project', 'emerald', true)
			ON CONFLICT (code) DO NOTHING;
		END IF;
	END $$;`

	if _, err := db.Exec(subTypeLogic); err != nil {
		log.Printf("Warning: Seeding assessment sub-types failed: %v", err)
	}

	return nil
}
