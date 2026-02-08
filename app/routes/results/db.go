package results

import (
	"database/sql"
	"fmt"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
)

// GetResultsByExamID fetches all results for a specific exam
func GetResultsByExamID(db *sql.DB, examID string) ([]*models.Result, error) {
	query := `
		SELECT 
			r.id, r.exam_id, r.student_id, r.paper_id, r.marks, r.grade_id,
			r.created_at, r.updated_at, r.deleted_at,
			s.id, s.student_id, s.first_name, s.last_name, s.gender,
			p.id, p.name, p.code
		FROM results r
		LEFT JOIN students s ON r.student_id = s.id
		LEFT JOIN papers p ON r.paper_id = p.id
		WHERE r.exam_id = $1 AND r.deleted_at IS NULL
		ORDER BY s.first_name, s.last_name
	`

	rows, err := db.Query(query, examID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}
	defer rows.Close()

	var results []*models.Result
	for rows.Next() {
		var result models.Result
		var student models.Student
		var paper models.Paper
		var deletedAt sql.NullTime

		err := rows.Scan(
			&result.ID, &result.ExamID, &result.StudentID, &result.PaperID,
			&result.Marks, &result.GradeID, &result.CreatedAt, &result.UpdatedAt, &deletedAt,
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName, &student.Gender,
			&paper.ID, &paper.Name, &paper.Code,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if deletedAt.Valid {
			result.DeletedAt = &deletedAt.Time
		}

		result.Student = &student
		result.Paper = &paper
		results = append(results, &result)
	}

	return results, nil
}

// GetResultByExamAndStudent fetches a specific student's result for an exam
func GetResultByExamAndStudent(db *sql.DB, examID, studentID string) (*models.Result, error) {
	query := `
		SELECT 
			r.id, r.exam_id, r.student_id, r.paper_id, r.marks, r.grade_id,
			r.created_at, r.updated_at, r.deleted_at
		FROM results r
		WHERE r.exam_id = $1 AND r.student_id = $2 AND r.deleted_at IS NULL
	`

	var result models.Result
	var deletedAt sql.NullTime

	err := db.QueryRow(query, examID, studentID).Scan(
		&result.ID, &result.ExamID, &result.StudentID, &result.PaperID,
		&result.Marks, &result.GradeID, &result.CreatedAt, &result.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No result found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch result: %w", err)
	}

	if deletedAt.Valid {
		result.DeletedAt = &deletedAt.Time
	}

	return &result, nil
}

// CreateResult inserts a new result record
func CreateResult(db *sql.DB, result *models.Result) error {
	query := `
		INSERT INTO results (exam_id, student_id, paper_id, term_id, marks, grade_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		result.ExamID,
		result.StudentID,
		result.PaperID,
		result.TermID,
		result.Marks,
		result.GradeID,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create result: %w", err)
	}

	return nil
}

// UpdateResult updates an existing result
func UpdateResult(db *sql.DB, result *models.Result) error {
	query := `
		UPDATE results
		SET marks = $1, grade_id = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := db.QueryRow(query, result.Marks, result.GradeID, result.ID).Scan(&result.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update result: %w", err)
	}

	return nil
}

// DeleteResult soft deletes a result
func DeleteResult(db *sql.DB, resultID string) error {
	query := `
		UPDATE results
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := db.Exec(query, resultID)
	if err != nil {
		return fmt.Errorf("failed to delete result: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("result not found or already deleted")
	}

	return nil
}

// BatchCreateOrUpdateResults efficiently saves multiple results at once
func BatchCreateOrUpdateResults(db *sql.DB, results []*models.Result) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements
	checkStmt, err := tx.Prepare(`
		SELECT id FROM results 
		WHERE exam_id = $1 AND student_id = $2 AND deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare check statement: %w", err)
	}
	defer checkStmt.Close()

	insertStmt, err := tx.Prepare(`
		INSERT INTO results (exam_id, student_id, paper_id, term_id, marks, grade_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	updateStmt, err := tx.Prepare(`
		UPDATE results
		SET marks = $1, grade_id = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	// Process each result
	for _, result := range results {
		var existingID string
		err := checkStmt.QueryRow(result.ExamID, result.StudentID).Scan(&existingID)

		if err == sql.ErrNoRows {
			// Create new result
			err = insertStmt.QueryRow(
				result.ExamID,
				result.StudentID,
				result.PaperID,
				result.TermID,
				result.Marks,
				result.GradeID,
			).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

			if err != nil {
				return fmt.Errorf("failed to insert result for student %s: %w", result.StudentID, err)
			}
		} else if err == nil {
			// Update existing result
			result.ID = existingID
			err = updateStmt.QueryRow(result.Marks, result.GradeID, existingID).Scan(&result.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to update result for student %s: %w", result.StudentID, err)
			}
		} else {
			return fmt.Errorf("failed to check existing result: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStudentsByClassID fetches all active students in a class
func GetStudentsByClassID(db *sql.DB, classID string, limit, offset int) ([]*models.Student, int, error) {
	// First get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM students WHERE class_id = $1 AND is_active = true AND deleted_at IS NULL", classID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch total students: %w", err)
	}

	query := `
		SELECT 
			id, student_id, first_name, last_name, date_of_birth, 
			gender, address, class_id, is_active, created_at, updated_at
		FROM students
		WHERE class_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY first_name, last_name
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, classID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch students: %w", err)
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var student models.Student
		var dob sql.NullTime
		var address sql.NullString
		var gender sql.NullString
		var classID sql.NullString

		err := rows.Scan(
			&student.ID,
			&student.StudentID,
			&student.FirstName,
			&student.LastName,
			&dob,
			&gender,
			&address,
			&classID,
			&student.IsActive,
			&student.CreatedAt,
			&student.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan student: %w", err)
		}

		if dob.Valid {
			student.DateOfBirth = &dob.Time
		}
		if address.Valid {
			student.Address = &address.String
		}
		if gender.Valid {
			genderVal := models.Gender(gender.String)
			student.Gender = &genderVal
		}
		if classID.Valid {
			student.ClassID = &classID.String
		}

		students = append(students, &student)
	}

	return students, total, nil
}

// StudentWithResult represents a student with their result for an exam
type StudentWithResult struct {
	Student *models.Student `json:"student"`
	Result  *models.Result  `json:"result,omitempty"`
}

// GetStudentsWithResultsByExam fetches all students in a class with their results for an exam
func GetStudentsWithResultsByExam(db *sql.DB, examID, classID string) ([]*StudentWithResult, error) {
	query := `
		SELECT 
			s.id, s.student_id, s.first_name, s.last_name, s.gender,
			r.id, r.marks, r.grade_id, r.created_at, r.updated_at
		FROM students s
		LEFT JOIN results r ON s.id = r.student_id AND r.exam_id = $1 AND r.deleted_at IS NULL
		WHERE s.class_id = $2 AND s.is_active = true AND s.deleted_at IS NULL
		ORDER BY s.first_name, s.last_name
	`

	rows, err := db.Query(query, examID, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students with results: %w", err)
	}
	defer rows.Close()

	var studentsWithResults []*StudentWithResult
	for rows.Next() {
		var student models.Student
		var resultID sql.NullString
		var marks sql.NullFloat64
		var gradeID sql.NullString
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var gender sql.NullString

		err := rows.Scan(
			&student.ID,
			&student.StudentID,
			&student.FirstName,
			&student.LastName,
			&gender,
			&resultID,
			&marks,
			&gradeID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student with result: %w", err)
		}

		if gender.Valid {
			genderVal := models.Gender(gender.String)
			student.Gender = &genderVal
		}

		swr := &StudentWithResult{
			Student: &student,
		}

		// If result exists, populate it
		if resultID.Valid {
			result := &models.Result{
				ID:        resultID.String,
				ExamID:    examID,
				StudentID: student.ID,
				Marks:     marks.Float64,
			}

			if gradeID.Valid {
				result.GradeID = &gradeID.String
			}
			if createdAt.Valid {
				result.CreatedAt = createdAt.Time
			}
			if updatedAt.Valid {
				result.UpdatedAt = updatedAt.Time
			}

			swr.Result = result
		}

		studentsWithResults = append(studentsWithResults, swr)
	}

	return studentsWithResults, nil
}

// ClassResultsMatrix represents data for the grid view
type ClassResultsMatrix struct {
	Students []*models.Student `json:"students"`
	Papers   []*models.Paper   `json:"papers"`
	Results  []*models.Result  `json:"results"`
}

// SubjectResultMatrix contains all data needed for the subject-level mark sheet
type SubjectResultMatrix struct {
	Subject       *models.Subject       `json:"subject"`
	Students      []*models.Student     `json:"students"`
	Papers        []*models.Paper       `json:"papers"`
	Weights       []*models.PaperWeight `json:"weights"`
	Results       []*models.Result      `json:"results"`
	Grades        []*models.Grade       `json:"grades"`
	TermID        string                `json:"term_id"`
	TotalStudents int                   `json:"total_students"`
}

// GetSubjectResultMatrix fetches data for the subject mark sheet
func GetSubjectResultMatrix(db *sql.DB, classID, subjectID, termID, assessmentTypeID string, limit, offset int) (*SubjectResultMatrix, error) {
	matrix := &SubjectResultMatrix{
		Students: []*models.Student{},
		Papers:   []*models.Paper{},
		Weights:  []*models.PaperWeight{},
		Results:  []*models.Result{},
		Grades:   []*models.Grade{},
		TermID:   termID,
	}

	// 1. Fetch Subject Info
	subject, err := database.GetSubjectByID(db, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subject: %v", err)
	}
	matrix.Subject = subject

	// 2. Fetch Students in Class (Paginated)
	students, total, err := GetStudentsByClassID(db, classID, limit, offset)
	if err != nil {
		return nil, err
	}
	matrix.Students = students
	matrix.TotalStudents = total

	// 3. Fetch Weights (and papers via weights)
	weightsQuery := `
		SELECT pw.id, pw.paper_id, pw.weight, p.name, p.code
		FROM paper_weights pw
		JOIN papers p ON pw.paper_id = p.id
		WHERE pw.class_id = $1 AND pw.subject_id = $2 AND pw.term_id = $3
		ORDER BY p.code
	`
	var pRows *sql.Rows
	pRows, err = db.Query(weightsQuery, classID, subjectID, termID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weights and papers: %v", err)
	}
	defer pRows.Close()

	for pRows.Next() {
		w := models.PaperWeight{}
		p := models.Paper{}
		if err := pRows.Scan(&w.ID, &w.PaperID, &w.Weight, &p.Name, &p.Code); err != nil {
			return nil, err
		}
		p.ID = w.PaperID
		matrix.Weights = append(matrix.Weights, &w)
		matrix.Papers = append(matrix.Papers, &p)
	}

	// Fallback: If no weights defined, fetch ALL papers for this subject to ensure columns appear
	if len(matrix.Papers) == 0 {
		papersQuery := `
			SELECT id, name, code 
			FROM papers 
			WHERE subject_id = $1 AND deleted_at IS NULL
			ORDER BY code
		`
		paperRows, err := db.Query(papersQuery, subjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch fallback papers: %v", err)
		}
		defer paperRows.Close()

		for paperRows.Next() {
			p := models.Paper{}
			if err := paperRows.Scan(&p.ID, &p.Name, &p.Code); err != nil {
				return nil, err
			}
			// Create dummy weight for matrix logic
			w := models.PaperWeight{
				PaperID:   p.ID,
				Weight:    0, // Default to 0
				ClassID:   classID,
				SubjectID: subjectID,
				TermID:    termID,
			}
			matrix.Papers = append(matrix.Papers, &p)
			matrix.Weights = append(matrix.Weights, &w)
		}
	}

	// 4. Fetch Results (Targeted fetching mode)
	if len(matrix.Papers) == 0 || len(matrix.Students) == 0 {
		return matrix, nil // No papers or students, no results to fetch
	}

	paperIDs := make([]string, len(matrix.Papers))
	for i, p := range matrix.Papers {
		paperIDs[i] = p.ID
	}

	studentIDs := make([]string, len(matrix.Students))
	for i, s := range matrix.Students {
		studentIDs[i] = s.ID
	}

	var resultsRows *sql.Rows
	if assessmentTypeID == "all" || assessmentTypeID == "" {
		// Optimized targeted query: only fetch results for the papers and students we actually care about in this page
		resultsQuery := `
			SELECT r.id, COALESCE(r.exam_id::text, ''), r.student_id, r.paper_id, r.marks, 
				COALESCE(p.name, pe.name, 'Unknown Paper'), 
				COALESCE(p.code, pe.code, '')
			FROM results r
			JOIN students s ON r.student_id = s.id
			LEFT JOIN papers p ON r.paper_id = p.id
			LEFT JOIN exams e ON r.exam_id = e.id
			LEFT JOIN papers pe ON e.paper_id = pe.id 
			WHERE r.student_id = ANY($1) 
			AND (r.paper_id = ANY($2) OR e.paper_id = ANY($2))
			AND r.deleted_at IS NULL
			ORDER BY r.created_at DESC
		`
		resultsRows, err = db.Query(resultsQuery, database.ToPostgresArray(studentIDs), database.ToPostgresArray(paperIDs))
	} else {
		// Still filter by assessment type if explicitly requested
		resultsQuery := `
			SELECT r.id, COALESCE(r.exam_id::text, ''), r.student_id, r.paper_id, r.marks, 
				COALESCE(p.name, pe.name, 'Unknown Paper'), 
				COALESCE(p.code, pe.code, '')
			FROM results r
			JOIN students s ON r.student_id = s.id
			LEFT JOIN papers p ON r.paper_id = p.id
			LEFT JOIN exams e ON r.exam_id = e.id
			LEFT JOIN papers pe ON e.paper_id = pe.id
			WHERE r.student_id = ANY($1) 
			AND e.assessment_type_id = $2
			AND (r.paper_id = ANY($3) OR e.paper_id = ANY($3))
			AND r.deleted_at IS NULL
			ORDER BY r.created_at DESC
		`
		resultsRows, err = db.Query(resultsQuery, database.ToPostgresArray(studentIDs), assessmentTypeID, database.ToPostgresArray(paperIDs))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %v", err)
	}
	defer resultsRows.Close()

	for resultsRows.Next() {
		var r models.Result
		var p models.Paper
		if err := resultsRows.Scan(&r.ID, &r.ExamID, &r.StudentID, &r.PaperID, &r.Marks, &p.Name, &p.Code); err != nil {
			return nil, err
		}
		p.ID = r.PaperID
		r.Paper = &p
		matrix.Results = append(matrix.Results, &r)
	}

	// 5. Fetch Grades
	grades, err := GetAllGrades(db)
	if err != nil {
		return nil, err
	}
	matrix.Grades = grades

	return matrix, nil
}

// GetResultsByStudentID fetches all results for a specific student with joined relations
func GetResultsByStudentID(db *sql.DB, studentID string) ([]*models.Result, error) {
	query := `
		SELECT 
			r.id, r.exam_id, r.student_id, r.paper_id, r.marks, r.grade_id,
			r.created_at, r.updated_at,
			e.id, e.name, e.type,
			p.id, p.name, p.code,
			s.id, s.name, s.code,
			g.id, g.name
		FROM results r
		LEFT JOIN exams e ON r.exam_id = e.id
		LEFT JOIN papers p ON r.paper_id = p.id
		LEFT JOIN subjects s ON p.subject_id = s.id
		LEFT JOIN grades g ON r.grade_id = g.id
		WHERE r.student_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch student results: %w", err)
	}
	defer rows.Close()

	var results []*models.Result
	for rows.Next() {
		var r models.Result
		var e models.Exam
		var p models.Paper
		var sub models.Subject
		var g models.Grade
		var gradeID, gradeName sql.NullString
		var gid sql.NullString

		err := rows.Scan(
			&r.ID, &r.ExamID, &r.StudentID, &r.PaperID, &r.Marks, &gradeID,
			&r.CreatedAt, &r.UpdatedAt,
			&e.ID, &e.Name, &e.Type,
			&p.ID, &p.Name, &p.Code,
			&sub.ID, &sub.Name, &sub.Code,
			&gid, &gradeName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student result: %w", err)
		}

		if gradeID.Valid {
			r.GradeID = &gradeID.String
			if gradeName.Valid {
				g.ID = gid.String
				g.Name = gradeName.String
				r.Grade = &g
			}
		}

		p.Subject = &sub
		r.Exam = &e
		r.Paper = &p
		results = append(results, &r)
	}

	return results, nil
}

// GetStudentAssessmentHistory fetches all exams for the student's class and their specific results
func GetStudentAssessmentHistory(db *sql.DB, studentID string) ([]*models.Result, error) {
	query := `
		SELECT 
			e.id, e.name, e.type, e.term_id,
			p.id, p.name, p.code,
			s.id, s.name, s.code,
			r.id, r.marks, r.grade_id,
			g.id, g.name
		FROM students st
		JOIN exams e ON st.class_id = e.class_id
		LEFT JOIN papers p ON e.paper_id = p.id
		LEFT JOIN subjects s ON p.subject_id = s.id
		LEFT JOIN results r ON e.id = r.exam_id AND st.id = r.student_id AND r.deleted_at IS NULL
		LEFT JOIN grades g ON r.grade_id = g.id
		WHERE st.id = $1 AND st.deleted_at IS NULL AND e.deleted_at IS NULL
		ORDER BY e.start_time DESC
	`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assessment history: %w", err)
	}
	defer rows.Close()

	var results []*models.Result
	for rows.Next() {
		var r models.Result
		var e models.Exam
		var p models.Paper
		var sub models.Subject
		var g models.Grade
		var resID, gradeID, gradeName, gid, termID sql.NullString
		var marks sql.NullFloat64

		err := rows.Scan(
			&e.ID, &e.Name, &e.Type, &termID,
			&p.ID, &p.Name, &p.Code,
			&sub.ID, &sub.Name, &sub.Code,
			&resID, &marks, &gradeID,
			&gid, &gradeName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assessment history: %w", err)
		}

		if termID.Valid {
			e.TermID = &termID.String
			r.TermID = &termID.String
		}

		p.Subject = &sub
		e.Paper = &p
		e.PaperID = p.ID
		r.Exam = &e
		r.Paper = &p
		r.PaperID = p.ID
		r.StudentID = studentID
		r.ExamID = e.ID

		if resID.Valid {
			r.ID = resID.String
			r.Marks = marks.Float64
			if gradeID.Valid {
				r.GradeID = &gradeID.String
				if gradeName.Valid {
					g.ID = gid.String
					g.Name = gradeName.String
					r.Grade = &g
				}
			}
		} else {
			// Mark as "no result" by setting ID to empty or leaving marks as 0
			// The frontend will check if ID is present or marks are null-ish
			r.ID = ""
		}

		results = append(results, &r)
	}

	return results, nil
}

// GetAllGrades fetches all grades from the system
func GetAllGrades(db *sql.DB) ([]*models.Grade, error) {
	query := `
		SELECT id, name, min_marks, max_marks, grade_value, is_active, created_at, updated_at
		FROM grades
		WHERE deleted_at IS NULL
		ORDER BY min_marks DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch grades: %w", err)
	}
	defer rows.Close()

	var grades []*models.Grade
	for rows.Next() {
		var g models.Grade
		err := rows.Scan(
			&g.ID, &g.Name, &g.MinMarks, &g.MaxMarks, &g.GradeValue,
			&g.IsActive, &g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, &g)
	}

	return grades, nil
}

// CreateGrade inserts a new grade record
func CreateGrade(db *sql.DB, g *models.Grade) error {
	query := `
		INSERT INTO grades (name, min_marks, max_marks, grade_value, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		g.Name, g.MinMarks, g.MaxMarks, g.GradeValue, g.IsActive,
	).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create grade: %w", err)
	}

	return nil
}

// UpdateGrade updates an existing grade record
func UpdateGrade(db *sql.DB, g *models.Grade) error {
	query := `
		UPDATE grades
		SET name = $1, min_marks = $2, max_marks = $3, grade_value = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`

	_, err := db.Exec(
		query,
		g.Name, g.MinMarks, g.MaxMarks, g.GradeValue, g.IsActive, g.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update grade: %w", err)
	}

	return nil
}

// DeleteGrade soft deletes a grade record
func DeleteGrade(db *sql.DB, id string) error {
	query := `UPDATE grades SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
