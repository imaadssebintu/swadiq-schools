package results

import (
	"database/sql"
	"fmt"
	"math"
)

// CalculateFinalSubjectMark computes the weighted total mark for a student in a specific subject
func CalculateFinalSubjectMark(db *sql.DB, studentID, subjectID string, termID, examTypeID *string) (float64, error) {
	// 1. Get all papers for the subject
	query := `SELECT id, weight FROM papers WHERE subject_id = $1 AND deleted_at IS NULL`
	rows, err := db.Query(query, subjectID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch papers: %v", err)
	}
	defer rows.Close()

	type PaperWeight struct {
		ID     string
		Weight int
	}
	var papers []PaperWeight
	totalWeight := 0

	for rows.Next() {
		var p PaperWeight
		if err := rows.Scan(&p.ID, &p.Weight); err != nil {
			return 0, err
		}
		papers = append(papers, p)
		totalWeight += p.Weight
	}

	// If no papers, return 0
	if len(papers) == 0 {
		return 0, nil
	}

	// 2. Fetch results for each paper
	// We need to filter by term and exam type if provided.
	// This query assumes we want the *latest* result for that paper within the constraints.
	// Or we might want to aggregate across multiple exams (e.g., specific exams).
	// For now, let's assume we are calculating based on available results for the given context.

	// Construct query dynamically based on filters
	resultQuery := `
		SELECT r.marks 
		FROM results r
		JOIN exams e ON r.exam_id = e.id
		WHERE r.student_id = $1 AND r.paper_id = $2 AND r.deleted_at IS NULL
	`
	args := []interface{}{studentID, ""} // Placeholder for paperID

	if termID != nil {
		resultQuery += " AND e.term_id = $3"
		args = append(args, *termID)
	}
	if examTypeID != nil {
		resultQuery += fmt.Sprintf(" AND e.assessment_type_id = $%d", len(args)+1)
		args = append(args, *examTypeID)
	}

	// Note: If multiple exams exist for the same paper in the same term/type context,
	// we might need to average them or take the latest.
	// Let's take the MAX for now (best performance) or AVG.
	// Using MAX to avoid penalizing if a re-take happened, but this logic can be adjusted.
	resultQuery = "SELECT MAX(r.marks) " + resultQuery[15:]

	var finalMark float64

	for _, paper := range papers {
		args[1] = paper.ID // Set paperID

		var marks sql.NullFloat64
		err := db.QueryRow(resultQuery, args...).Scan(&marks)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("failed to fetch result for paper %s: %v", paper.ID, err)
		}

		if marks.Valid {
			// Calculate contribution: (Marks * Weight) / 100
			contribution := (marks.Float64 * float64(paper.Weight)) / 100.0
			finalMark += contribution
		}
	}

	// Normalize if total weight is not 100?
	// The user asked "all have to add up to 100".
	// If they sum to 100, the logic matches.
	// If they sum to MORE, the mark could exceed 100.
	// If less, the mark is out of `totalWeight`.

	// Option: Scale to 100 if needed.
	// finalMark = (finalMark / float64(totalWeight)) * 100

	return math.Round(finalMark*100) / 100, nil
}
