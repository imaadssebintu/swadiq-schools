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

	// 6. Create grades table
	err = createGradesTable(db)
	if err != nil {
		return err
	}

	// 7. Add grade_value to grades if not exists
	err = addGradeValueColumn(db)
	if err != nil {
		return err
	}

	// 8. Create paper_weights table
	err = createPaperWeightsTable(db)
	if err != nil {
		return err
	}

	// 9. Create teacher_allowance_accruals table
	err = createTeacherAllowanceAccrualsTable(db)
	if err != nil {
		return err
	}

	// 10. Create teacher_attendances table
	err = createTeacherAttendancesTable(db)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func createGradesTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS grades (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(50) UNIQUE NOT NULL,
		min_marks DECIMAL(5,2) NOT NULL,
		max_marks DECIMAL(5,2) NOT NULL,
		grade_value DECIMAL(5,2) DEFAULT 0,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	)`
	_, err := db.Exec(query)
	return err
}

func addGradeValueColumn(db *sql.DB) error {
	query := `ALTER TABLE grades ADD COLUMN IF NOT EXISTS grade_value DECIMAL(5,2) DEFAULT 0`
	_, err := db.Exec(query)
	return err
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

func createPaperWeightsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS paper_weights (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
		subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
		paper_id UUID NOT NULL REFERENCES papers(id) ON DELETE CASCADE,
		term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
		weight INTEGER NOT NULL CHECK (weight >= 0 AND weight <= 100),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(class_id, subject_id, paper_id, term_id)
	)`)
	return err
}

func createTeacherAllowanceAccrualsTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS teacher_allowance_accruals (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		teacher_id UUID NOT NULL REFERENCES users(id),
		amount BIGINT NOT NULL,
		date DATE NOT NULL,
		status VARCHAR(20) DEFAULT 'unpaid',
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(teacher_id, date)
	)`
	_, err := db.Exec(query)
	return err
}

func createTeacherAttendancesTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS teacher_attendances (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		teacher_id UUID NOT NULL REFERENCES users(id),
		date DATE NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'present',
		remarks TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(teacher_id, date)
	)`
	_, err := db.Exec(query)
	return err
}
