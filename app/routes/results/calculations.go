package results

import (
	"database/sql"
	"fmt"
	"math"
)

// CalculateFinalSubjectMark computes the weighted total mark for a student in a specific subject
func CalculateFinalSubjectMark(db *sql.DB, studentID, subjectID string, termID, examTypeID *string, classID string) (float64, error) {
	if termID == nil {
		return 0, fmt.Errorf("term_id is required for weighted calculation")
	}

	// 1. Get paper weights for the class, subject, and term
	query := `SELECT paper_id, weight FROM paper_weights WHERE class_id = $1 AND subject_id = $2 AND term_id = $3`
	rows, err := db.Query(query, classID, subjectID, *termID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch paper weights: %v", err)
	}
	defer rows.Close()

	type PaperWeightInfo struct {
		PaperID string
		Weight  int
	}
	var weights []PaperWeightInfo

	for rows.Next() {
		var w PaperWeightInfo
		if err := rows.Scan(&w.PaperID, &w.Weight); err != nil {
			return 0, err
		}
		weights = append(weights, w)
	}

	// If no weights defined, we cannot calculate a weighted total
	if len(weights) == 0 {
		return 0, nil
	}

	// 2. Fetch results for each paper
	resultQuery := `
		SELECT MAX(r.marks) 
		FROM results r
		JOIN exams e ON r.exam_id = e.id
		WHERE r.student_id = $1 AND r.paper_id = $2 AND e.term_id = $3 AND r.deleted_at IS NULL
	`
	args := []interface{}{studentID, "", *termID}

	if examTypeID != nil {
		resultQuery += " AND e.assessment_type_id = $4"
		args = append(args, *examTypeID)
	}

	var finalMark float64
	for _, pw := range weights {
		args[1] = pw.PaperID

		var marks sql.NullFloat64
		err := db.QueryRow(resultQuery, args...).Scan(&marks)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to fetch result for paper %s: %v", pw.PaperID, err)
		}

		if marks.Valid {
			// Calculate contribution: (Marks * Weight) / 100
			contribution := (marks.Float64 * float64(pw.Weight)) / 100.0
			finalMark += contribution
		}
	}

	return math.Round(finalMark*100) / 100, nil
}
