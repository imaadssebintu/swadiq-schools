package academic

import (
	"database/sql"
	"swadiq-schools/app/models"
)

// Data Access Functions

func GetAcademicYearsForTemplate(db *sql.DB) ([]*models.AcademicYear, error) {
	query := `SELECT ay.id, ay.name, ay.start_date, ay.end_date, ay.is_current, ay.is_active, ay.created_at, ay.updated_at,
			  COALESCE(t.term_count, 0) as term_count
			  FROM academic_years ay
			  LEFT JOIN (
				  SELECT academic_year_id, COUNT(*) as term_count 
				  FROM terms 
				  WHERE deleted_at IS NULL 
				  GROUP BY academic_year_id
			  ) t ON ay.id = t.academic_year_id
			  WHERE ay.deleted_at IS NULL 
			  ORDER BY ay.start_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.AcademicYear{}, nil
	}
	defer rows.Close()

	var years []*models.AcademicYear
	for rows.Next() {
		year := &models.AcademicYear{}
		var termCount int
		err := rows.Scan(&year.ID, &year.Name, &year.StartDate.Time, &year.EndDate.Time,
			&year.IsCurrent, &year.IsActive, &year.CreatedAt, &year.UpdatedAt, &termCount)
		if err != nil {
			continue
		}

		// Load terms for this academic year
		terms, _ := getTermsByAcademicYearID(db, year.ID)
		if terms == nil {
			terms = []*models.Term{}
		}
		year.Terms = terms

		years = append(years, year)
	}
	if years == nil {
		years = []*models.AcademicYear{}
	}
	return years, nil
}

func getAcademicYearByID(db *sql.DB, id string) (*models.AcademicYear, error) {
	query := `SELECT id, name, start_date, end_date, is_current, is_active, created_at, updated_at
			  FROM academic_years WHERE id = $1 AND deleted_at IS NULL`

	year := &models.AcademicYear{}
	err := db.QueryRow(query, id).Scan(&year.ID, &year.Name, &year.StartDate.Time, &year.EndDate.Time,
		&year.IsCurrent, &year.IsActive, &year.CreatedAt, &year.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return year, nil
}

func createAcademicYear(db *sql.DB, year *models.AcademicYear) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if year.IsActive {
		deactivateQuery := `UPDATE academic_years SET is_active = false`
		_, err = tx.Exec(deactivateQuery)
		if err != nil {
			return err
		}
	}

	insertQuery := `INSERT INTO academic_years (name, start_date, end_date, is_current, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = tx.QueryRow(insertQuery, year.Name, year.StartDate.Time, year.EndDate.Time,
		year.IsCurrent, year.IsActive).Scan(&year.ID, &year.CreatedAt, &year.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func updateAcademicYear(db *sql.DB, year *models.AcademicYear) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if year.IsActive {
		deactivateQuery := `UPDATE academic_years SET is_active = false WHERE id != $1`
		_, err = tx.Exec(deactivateQuery, year.ID)
		if err != nil {
			return err
		}
	}

	updateQuery := `UPDATE academic_years
			  SET name = $1, start_date = $2, end_date = $3, is_current = $4, is_active = $5, updated_at = NOW()
			  WHERE id = $6 AND deleted_at IS NULL`

	_, err = tx.Exec(updateQuery, year.Name, year.StartDate.Time, year.EndDate.Time,
		year.IsCurrent, year.IsActive, year.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func deleteAcademicYear(db *sql.DB, id string) error {
	query := `UPDATE academic_years SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

func GetTermsForTemplate(db *sql.DB) ([]*models.Term, error) {
	query := `SELECT t.id, t.academic_year_id, t.name, t.start_date, t.end_date, t.is_current, t.is_active, t.created_at, t.updated_at,
			  ay.name as academic_year_name
			  FROM terms t
			  LEFT JOIN academic_years ay ON t.academic_year_id = ay.id
			  WHERE t.deleted_at IS NULL ORDER BY t.start_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Term{}, nil
	}
	defer rows.Close()

	var terms []*models.Term
	for rows.Next() {
		term := &models.Term{}
		var academicYearName *string
		err := rows.Scan(&term.ID, &term.AcademicYearID, &term.Name, &term.StartDate.Time, &term.EndDate.Time,
			&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt, &academicYearName)
		if err != nil {
			continue
		}
		if academicYearName != nil {
			term.AcademicYear = &models.AcademicYear{ID: term.AcademicYearID, Name: *academicYearName}
		}
		terms = append(terms, term)
	}
	if terms == nil {
		terms = []*models.Term{}
	}
	return terms, nil
}

func getTermByID(db *sql.DB, id string) (*models.Term, error) {
	query := `SELECT t.id, t.academic_year_id, t.name, t.start_date, t.end_date, t.is_current, t.is_active, t.created_at, t.updated_at,
			  ay.name as academic_year_name
			  FROM terms t
			  LEFT JOIN academic_years ay ON t.academic_year_id = ay.id
			  WHERE t.id = $1 AND t.deleted_at IS NULL`

	term := &models.Term{}
	var academicYearName *string
	err := db.QueryRow(query, id).Scan(&term.ID, &term.AcademicYearID, &term.Name, &term.StartDate.Time, &term.EndDate.Time,
		&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt, &academicYearName)
	if err != nil {
		return nil, err
	}
	if academicYearName != nil {
		term.AcademicYear = &models.AcademicYear{ID: term.AcademicYearID, Name: *academicYearName}
	}
	return term, nil
}

func createTerm(db *sql.DB, term *models.Term) error {
	query := `INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive).Scan(&term.ID, &term.CreatedAt, &term.UpdatedAt)
	return err
}

func updateTerm(db *sql.DB, term *models.Term) error {
	query := `UPDATE terms
			  SET academic_year_id = $1, name = $2, start_date = $3, end_date = $4, is_current = $5, is_active = $6, updated_at = NOW()
			  WHERE id = $7 AND deleted_at IS NULL`

	_, err := db.Exec(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive, term.ID)
	return err
}

func deleteTerm(db *sql.DB, id string) error {
	query := `UPDATE terms SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

func getTermsByAcademicYearID(db *sql.DB, yearID string) ([]*models.Term, error) {
	query := `SELECT id, academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at
			  FROM terms WHERE academic_year_id = $1 AND deleted_at IS NULL ORDER BY start_date`

	rows, err := db.Query(query, yearID)
	if err != nil {
		return []*models.Term{}, err
	}
	defer rows.Close()

	var terms []*models.Term
	for rows.Next() {
		term := &models.Term{}
		err := rows.Scan(&term.ID, &term.AcademicYearID, &term.Name, &term.StartDate.Time, &term.EndDate.Time,
			&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt)
		if err != nil {
			continue
		}
		terms = append(terms, term)
	}
	return terms, nil
}

func setCurrentAcademicYear(db *sql.DB, yearID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deactivateQuery := `UPDATE academic_years SET is_current = false`
	_, err = tx.Exec(deactivateQuery)
	if err != nil {
		return err
	}

	setCurrentQuery := `UPDATE academic_years SET is_current = true WHERE id = $1`
	_, err = tx.Exec(setCurrentQuery, yearID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func setCurrentTerm(db *sql.DB, termID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deactivateQuery := `UPDATE terms SET is_current = false`
	_, err = tx.Exec(deactivateQuery)
	if err != nil {
		return err
	}

	setCurrentQuery := `UPDATE terms SET is_current = true WHERE id = $1`
	_, err = tx.Exec(setCurrentQuery, termID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func autoSetCurrentAcademicYear(db *sql.DB) error {
	query := `UPDATE academic_years SET is_current = (start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE)`
	_, err := db.Exec(query)
	return err
}

func autoSetCurrentTerm(db *sql.DB) error {
	query := `UPDATE terms SET is_current = (start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE)`
	_, err := db.Exec(query)
	return err
}

func GetAllAssessmentTypes(db *sql.DB) ([]*models.AssessmentType, error) {
	query := `SELECT id, name, code, color, is_active, created_at, updated_at 
			  FROM assessment_types WHERE deleted_at IS NULL ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.AssessmentType{}, nil
	}
	defer rows.Close()

	var types []*models.AssessmentType
	for rows.Next() {
		t := &models.AssessmentType{}
		err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.Color, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		types = append(types, t)
	}
	return types, nil
}

func CreateAssessmentType(db *sql.DB, t *models.AssessmentType) error {
	query := `INSERT INTO assessment_types (name, code, color, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, t.Name, t.Code, t.Color, t.IsActive).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

func UpdateAssessmentType(db *sql.DB, t *models.AssessmentType) error {
	query := `UPDATE assessment_types SET name = $1, code = $2, color = $3, is_active = $4, updated_at = NOW()
			  WHERE id = $5 AND deleted_at IS NULL`

	_, err := db.Exec(query, t.Name, t.Code, t.Color, t.IsActive, t.ID)
	return err
}

func DeleteAssessmentType(db *sql.DB, id string) error {
	query := `UPDATE assessment_types SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
