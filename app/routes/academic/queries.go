package academic

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"swadiq-schools/app/models"
)

// Data Access Functions

func GetAcademicYearsForTemplate(db *sql.DB) ([]*models.AcademicYear, error) {
	query := `SELECT ay.id, ay.name, ay.start_date, ay.end_date, ay.is_current, ay.is_current as is_active,
			  ay.created_at, ay.updated_at,
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
	query := `SELECT id, name, start_date, end_date, is_current, is_current as is_active, created_at, updated_at
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

	if year.IsCurrent {
		deactivateQuery := `UPDATE academic_years SET is_current = false`
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

	if year.IsCurrent {
		deactivateQuery := `UPDATE academic_years SET is_current = false WHERE id != $1`
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
	query := `SELECT t.id, t.academic_year_id, t.name, t.start_date, t.end_date, t.is_current, 
			  (t.is_current OR (t.start_date <= CURRENT_DATE AND t.end_date >= CURRENT_DATE)) as is_active,
			  t.created_at, t.updated_at, ay.name as academic_year_name
			  FROM terms t
			  INNER JOIN academic_years ay ON t.academic_year_id = ay.id
			  WHERE t.deleted_at IS NULL AND ay.is_current = true
			  ORDER BY t.start_date DESC`

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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify academic year is active
	var isYearActive bool
	err = tx.QueryRow(`SELECT is_active FROM academic_years WHERE id = $1 AND deleted_at IS NULL`, term.AcademicYearID).Scan(&isYearActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("the specified academic year does not exist")
		}
		return fmt.Errorf("failed to verify academic year status: %v", err)
	}
	if !isYearActive {
		return errors.New("terms can only be registered for the currently active academic year")
	}

	if term.IsCurrent {
		deactivateQuery := `UPDATE terms SET is_current = false`
		_, err = tx.Exec(deactivateQuery)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive).Scan(&term.ID, &term.CreatedAt, &term.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func updateTerm(db *sql.DB, term *models.Term) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify academic year is active
	var isYearActive bool
	err = tx.QueryRow(`SELECT is_active FROM academic_years WHERE id = $1 AND deleted_at IS NULL`, term.AcademicYearID).Scan(&isYearActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("the specified academic year does not exist")
		}
		return fmt.Errorf("failed to verify academic year status: %v", err)
	}
	if !isYearActive {
		return errors.New("terms can only be assigned to the currently active academic year")
	}

	if term.IsCurrent {
		deactivateQuery := `UPDATE terms SET is_current = false WHERE id != $1`
		_, err = tx.Exec(deactivateQuery, term.ID)
		if err != nil {
			return err
		}
	}

	query := `UPDATE terms
			  SET academic_year_id = $1, name = $2, start_date = $3, end_date = $4, is_current = $5, is_active = $6, updated_at = NOW()
			  WHERE id = $7 AND deleted_at IS NULL`

	_, err = tx.Exec(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive, term.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func deleteTerm(db *sql.DB, id string) error {
	query := `UPDATE terms SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

func getTermsByAcademicYearID(db *sql.DB, yearID string) ([]*models.Term, error) {
	query := `SELECT id, academic_year_id, name, start_date, end_date, is_current, 
			  (is_current OR (start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE)) as is_active,
			  created_at, updated_at
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

func GetCurrentTerm(db *sql.DB) (*models.Term, error) {
	term := &models.Term{}
	query := `SELECT id, academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at
			  FROM terms WHERE is_current = true AND deleted_at IS NULL LIMIT 1`

	err := db.QueryRow(query).Scan(
		&term.ID, &term.AcademicYearID, &term.Name, &term.StartDate.Time, &term.EndDate.Time,
		&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return term, nil
}

func GetAllAssessmentTypes(db *sql.DB) ([]*models.AssessmentType, error) {
	query := `SELECT t.id, t.name, t.code, t.term_id, t.category_id, c.name as category_name, t.weight, t.color, t.all_classes, t.is_active, t.created_at, t.updated_at,
			  (SELECT json_agg(json_build_object('id', cl.id, 'name', cl.name))
			   FROM assessment_type_classes atc
			   JOIN classes cl ON atc.class_id = cl.id
			   WHERE atc.assessment_type_id = t.id) as classes
			  FROM assessment_types t
			  JOIN assessment_categories c ON t.category_id = c.id
			  WHERE t.deleted_at IS NULL ORDER BY c.name, t.name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := []*models.AssessmentType{}
	for rows.Next() {
		t := &models.AssessmentType{}
		var classesJSON []byte
		err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.TermID, &t.CategoryID, &t.CategoryName, &t.Weight, &t.Color, &t.AllClasses, &t.IsActive, &t.CreatedAt, &t.UpdatedAt, &classesJSON)
		if err != nil {
			log.Printf("Error scanning assessment type: %v", err)
			continue
		}

		if classesJSON != nil {
			if err := json.Unmarshal(classesJSON, &t.Classes); err != nil {
				log.Printf("Error unmarshaling classes for assessment type: %v", err)
			}
		}

		types = append(types, t)
	}
	return types, nil
}

func CreateAssessmentType(db *sql.DB, t *models.AssessmentType) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO assessment_types (name, code, term_id, category_id, weight, color, all_classes, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query, t.Name, t.Code, t.TermID, t.CategoryID, t.Weight, t.Color, t.AllClasses, t.IsActive).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert class links if all_classes is false
	if !t.AllClasses && len(t.Classes) > 0 {
		for _, class := range t.Classes {
			_, err = tx.Exec(`INSERT INTO assessment_type_classes (assessment_type_id, class_id) VALUES ($1, $2)`, t.ID, class.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func UpdateAssessmentType(db *sql.DB, t *models.AssessmentType) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE assessment_types SET name = $1, code = $2, term_id = $3, category_id = $4, weight = $5, color = $6, all_classes = $7, is_active = $8, updated_at = NOW()
			  WHERE id = $9 AND deleted_at IS NULL`

	_, err = tx.Exec(query, t.Name, t.Code, t.TermID, t.CategoryID, t.Weight, t.Color, t.AllClasses, t.IsActive, t.ID)
	if err != nil {
		return err
	}

	// Clear existing class links
	_, err = tx.Exec(`DELETE FROM assessment_type_classes WHERE assessment_type_id = $1`, t.ID)
	if err != nil {
		return err
	}

	// Insert new class links if all_classes is false
	if !t.AllClasses && len(t.Classes) > 0 {
		for _, class := range t.Classes {
			_, err = tx.Exec(`INSERT INTO assessment_type_classes (assessment_type_id, class_id) VALUES ($1, $2)`, t.ID, class.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func DeleteAssessmentType(db *sql.DB, id string) error {
	query := `UPDATE assessment_types SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// AssessmentCategory Queries
func GetAllAssessmentCategories(db *sql.DB) ([]*models.AssessmentCategory, error) {
	query := `SELECT id, name, code, color, display_style, is_active, created_at, updated_at 
			  FROM assessment_categories WHERE deleted_at IS NULL ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := []*models.AssessmentCategory{}
	for rows.Next() {
		c := &models.AssessmentCategory{}
		err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Color, &c.DisplayStyle, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			continue
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func CreateAssessmentCategory(db *sql.DB, c *models.AssessmentCategory) error {
	query := `INSERT INTO assessment_categories (name, code, color, display_style, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, c.Name, c.Code, c.Color, c.DisplayStyle, c.IsActive).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return err
}

func UpdateAssessmentCategory(db *sql.DB, c *models.AssessmentCategory) error {
	query := `UPDATE assessment_categories SET name = $1, code = $2, color = $3, display_style = $4, is_active = $5, updated_at = NOW()
			  WHERE id = $6 AND deleted_at IS NULL`

	_, err := db.Exec(query, c.Name, c.Code, c.Color, c.DisplayStyle, c.IsActive, c.ID)
	return err
}

func DeleteAssessmentCategory(db *sql.DB, id string) error {
	query := `UPDATE assessment_categories SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
