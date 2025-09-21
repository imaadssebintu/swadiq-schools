package teachers

import (
	"database/sql"
	"fmt"
	"strings"
	"swadiq-schools/app/models"
)

// GenerateTeacherID generates a unique teacher ID based on name
func GenerateTeacherID(firstName, lastName string) string {
	// Convert to uppercase and take first 3 characters of each name
	firstInitial := strings.ToUpper(firstName)
	lastInitial := strings.ToUpper(lastName)

	if len(firstInitial) > 3 {
		firstInitial = firstInitial[:3]
	}
	if len(lastInitial) > 3 {
		lastInitial = lastInitial[:3]
	}

	// Generate ID in format: TCH-FIRSTLAST-001
	return fmt.Sprintf("TCH-%s%s-001", firstInitial, lastInitial)
}

// ValidateTeacherData validates teacher input data
func ValidateTeacherData(teacher *models.User) []string {
	var errors []string

	if teacher.FirstName == "" {
		errors = append(errors, "First name is required")
	}

	if teacher.LastName == "" {
		errors = append(errors, "Last name is required")
	}

	if teacher.Email == "" {
		errors = append(errors, "Email is required")
	}

	if !isValidEmail(teacher.Email) {
		errors = append(errors, "Invalid email format")
	}

	return errors
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// FormatTeacherName formats teacher name for display
func FormatTeacherName(firstName, lastName string) string {
	return fmt.Sprintf("%s %s", strings.Title(firstName), strings.Title(lastName))
}

// GetTeacherRoles returns the default roles for a teacher
func GetTeacherRoles() []string {
	return []string{"teacher", "class_teacher"}
}

// GetTeacherByID retrieves a single teacher by ID
func GetTeacherByID(db *sql.DB, teacherID string) (*models.User, error) {
	query := `SELECT u.id, u.email, u.first_name, u.last_name, u.phone, u.is_active, u.created_at, u.updated_at
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE u.id = $1 AND r.name IN ('class_teacher', 'subject_teacher')
			  AND u.is_active = true`

	teacher := &models.User{}
	err := db.QueryRow(query, teacherID).Scan(
		&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
		&teacher.Phone, &teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return teacher, nil
}

// SearchTeachers searches for teachers by name or email
func SearchTeachers(db *sql.DB, query string, limit int) ([]*models.User, error) {
	teachers, _, err := SearchTeachersWithPagination(db, query, limit, 0)
	return teachers, err
}

// SearchTeachersWithPagination searches for teachers with pagination support
func SearchTeachersWithPagination(db *sql.DB, query string, limit, offset int) ([]*models.User, int, error) {
	var totalCount int
	var teachers []*models.User

	if query == "" {
		// Count total teachers
		countQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher')
			AND u.is_active = true
		`
		err := db.QueryRow(countQuery).Scan(&totalCount)
		if err != nil {
			return nil, 0, err
		}

		// Get teachers with pagination
		sqlQuery := `
			SELECT u.id, u.email, u.first_name, u.last_name, u.phone, u.is_active, u.created_at, u.updated_at
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher')
			AND u.is_active = true
			ORDER BY u.created_at DESC
			LIMIT $1 OFFSET $2
		`
		rows, err := db.Query(sqlQuery, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		teachers, err = scanTeachers(rows)
		if err != nil {
			return nil, 0, err
		}
	} else {
		// Search by name or email with count
		searchPattern := "%" + strings.ToLower(query) + "%"

		// Count matching teachers
		countQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher')
			AND u.is_active = true
			AND (LOWER(u.first_name) LIKE $1 OR LOWER(u.last_name) LIKE $1 OR LOWER(u.email) LIKE $1
				 OR LOWER(CONCAT(u.first_name, ' ', u.last_name)) LIKE $1)
		`
		err := db.QueryRow(countQuery, searchPattern).Scan(&totalCount)
		if err != nil {
			return nil, 0, err
		}

		// Get matching teachers with pagination
		sqlQuery := `
			SELECT u.id, u.email, u.first_name, u.last_name, u.phone, u.is_active, u.created_at, u.updated_at
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher')
			AND u.is_active = true
			AND (LOWER(u.first_name) LIKE $1 OR LOWER(u.last_name) LIKE $1 OR LOWER(u.email) LIKE $1
				 OR LOWER(CONCAT(u.first_name, ' ', u.last_name)) LIKE $1)
			ORDER BY u.first_name, u.last_name
			LIMIT $2 OFFSET $3
		`
		rows, err := db.Query(sqlQuery, searchPattern, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		teachers, err = scanTeachers(rows)
		if err != nil {
			return nil, 0, err
		}
	}

	return teachers, totalCount, nil
}

// scanTeachers scans database rows into User models
func scanTeachers(rows *sql.Rows) ([]*models.User, error) {
	var teachers []*models.User

	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(
			&teacher.ID,
			&teacher.Email,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Phone,
			&teacher.IsActive,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}
