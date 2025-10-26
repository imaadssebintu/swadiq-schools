package database

import (
	"database/sql"
	"fmt"
	"strings"
	"swadiq-schools/app/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// StudentFilters represents filtering options for students
type StudentFilters struct {
	Search    string
	Status    string
	ClassID   string
	Gender    string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password, first_name, last_name, is_active, created_at, updated_at 
			  FROM users WHERE email = $1 AND is_active = true`

	err := db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(db *sql.DB, userID string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password, first_name, last_name, is_active, created_at, updated_at
			  FROM users WHERE id = $1 AND is_active = true`

	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserRoles(db *sql.DB, userID string) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func CreateSession(db *sql.DB, sessionID string, userID string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, sessionID, userID, expiresAt, time.Now())
	return err
}

func GetSessionByID(db *sql.DB, sessionID string) (*models.Session, error) {
	session := &models.Session{}
	query := `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1 AND expires_at > NOW()`

	err := db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return session, nil
}

func DeleteSession(db *sql.DB, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := db.Exec(query, sessionID)
	return err
}

func UpdateUserPassword(db *sql.DB, userID string, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.Exec(query, hashedPassword, userID)
	return err
}

func GetAllStudents(db *sql.DB) ([]models.Student, error) {
	// Simple query first to check if table exists
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth,
			  s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at
			  FROM students s
			  WHERE s.is_active = true ORDER BY s.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		// Return empty slice if table doesn't exist
		return []models.Student{}, nil
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student

		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
			&student.DateOfBirth, &student.Gender, &student.Address,
			&student.ClassID, &student.IsActive, &student.CreatedAt, &student.UpdatedAt,
		)
		if err != nil {
			continue
		}

		students = append(students, student)
	}
	return students, nil
}

// GetStudentsWithDetails gets all students with their class and parent information (SUPER OPTIMIZED)
func GetStudentsWithDetails(db *sql.DB) ([]models.Student, error) {
	// Ultra-fast query - get only essential data in one go
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth,
			  s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name,
			  p.first_name as parent_first_name, p.last_name as parent_last_name
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  LEFT JOIN student_parents sp ON s.id = sp.student_id
			  LEFT JOIN parents p ON sp.parent_id = p.id AND p.is_active = true
			  WHERE s.is_active = true
			  ORDER BY s.created_at DESC
			  LIMIT 50`

	rows, err := db.Query(query)
	if err != nil {
		return []models.Student{}, err
	}
	defer rows.Close()

	studentMap := make(map[string]*models.Student)

	for rows.Next() {
		var student models.Student
		var className, parentFirstName, parentLastName *string

		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
			&student.DateOfBirth, &student.Gender, &student.Address,
			&student.ClassID, &student.IsActive, &student.CreatedAt, &student.UpdatedAt,
			&className, &parentFirstName, &parentLastName,
		)
		if err != nil {
			continue
		}

		// Check if student already exists in map
		if existingStudent, exists := studentMap[student.ID]; exists {
			// Add parent to existing student if not already added
			if parentFirstName != nil && parentLastName != nil {
				parentExists := false
				parentName := *parentFirstName + " " + *parentLastName
				for _, p := range existingStudent.Parents {
					if p.FirstName+" "+p.LastName == parentName {
						parentExists = true
						break
					}
				}
				if !parentExists {
					parent := &models.Parent{
						FirstName: *parentFirstName,
						LastName:  *parentLastName,
					}
					existingStudent.Parents = append(existingStudent.Parents, parent)
				}
			}
		} else {
			// Create new student entry
			if className != nil {
				student.Class = &models.Class{
					Name: *className,
				}
				if student.ClassID != nil {
					student.Class.ID = *student.ClassID
				}
			}

			// Add parent if exists
			if parentFirstName != nil && parentLastName != nil {
				parent := &models.Parent{
					FirstName: *parentFirstName,
					LastName:  *parentLastName,
				}
				student.Parents = []*models.Parent{parent}
			}

			studentMap[student.ID] = &student
		}
	}

	// Convert map to slice
	var students []models.Student
	for _, student := range studentMap {
		students = append(students, *student)
	}

	return students, nil
}

// GetStudentsForTable gets students optimized for table display (minimal data, fast query)
func GetStudentsForTable(db *sql.DB, limit int, offset int) ([]models.Student, error) {
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.gender, 
			  s.is_active, s.created_at, c.name as class_name
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  ORDER BY s.created_at DESC
			  LIMIT $1 OFFSET $2`
	
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		var className sql.NullString
		
		err := rows.Scan(
			&student.ID,
			&student.StudentID,
			&student.FirstName,
			&student.LastName,
			&student.Gender,
			&student.IsActive,
			&student.CreatedAt,
			&className,
		)
		if err != nil {
			return nil, err
		}

		if className.Valid {
			student.Class = &models.Class{Name: className.String}
		}

		students = append(students, student)
	}

	return students, nil
}

// SearchStudents searches for students by name or student ID
func SearchStudents(db *sql.DB, query string) ([]models.Student, error) {
	searchQuery := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.gender, 
					s.is_active, s.created_at, c.name as class_name
					FROM students s
					LEFT JOIN classes c ON s.class_id = c.id
					WHERE s.is_active = true 
					AND (LOWER(s.first_name) LIKE LOWER($1) 
						OR LOWER(s.last_name) LIKE LOWER($1)
						OR LOWER(s.student_id) LIKE LOWER($1)
						OR LOWER(CONCAT(s.first_name, ' ', s.last_name)) LIKE LOWER($1))
					ORDER BY s.first_name, s.last_name
					LIMIT 20`
	
	searchTerm := "%" + query + "%"
	rows, err := db.Query(searchQuery, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		var className sql.NullString
		
		err := rows.Scan(
			&student.ID,
			&student.StudentID,
			&student.FirstName,
			&student.LastName,
			&student.Gender,
			&student.IsActive,
			&student.CreatedAt,
			&className,
		)
		if err != nil {
			return nil, err
		}

		if className.Valid {
			student.Class = &models.Class{Name: className.String}
		}

		students = append(students, student)
	}

	return students, nil
}

func GetStudentsWithFilters(db *sql.DB, filters StudentFilters) ([]models.Student, error) {
	// Build the base query
	baseQuery := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth,
				  s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
				  c.name as class_name,
				  p.first_name as parent_first_name, p.last_name as parent_last_name
				  FROM students s
				  LEFT JOIN classes c ON s.class_id = c.id
				  LEFT JOIN student_parents sp ON s.id = sp.student_id
				  LEFT JOIN parents p ON sp.parent_id = p.id AND p.is_active = true`

	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Always filter for active students unless specifically looking for inactive
	if filters.Status == "inactive" {
		conditions = append(conditions, "s.is_active = false")
	} else if filters.Status == "active" || filters.Status == "" {
		conditions = append(conditions, "s.is_active = true")
	}

	// Search filter (name, student ID, or parent name)
	if filters.Search != "" {
		searchCondition := fmt.Sprintf(`(
			LOWER(s.first_name) LIKE LOWER($%d) OR
			LOWER(s.last_name) LIKE LOWER($%d) OR
			LOWER(s.student_id) LIKE LOWER($%d) OR
			LOWER(p.first_name || ' ' || p.last_name) LIKE LOWER($%d)
		)`, argIndex, argIndex, argIndex, argIndex)
		conditions = append(conditions, searchCondition)
		args = append(args, "%"+filters.Search+"%")
		argIndex++
	}

	// Class filter
	if filters.ClassID != "" {
		conditions = append(conditions, fmt.Sprintf("s.class_id = $%d", argIndex))
		args = append(args, filters.ClassID)
		argIndex++
	}

	// Gender filter
	if filters.Gender != "" {
		conditions = append(conditions, fmt.Sprintf("s.gender = $%d", argIndex))
		args = append(args, filters.Gender)
		argIndex++
	}

	// Date range filters
	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at <= $%d", argIndex))
		args = append(args, filters.DateTo+" 23:59:59")
		argIndex++
	}

	// Add WHERE clause if we have conditions
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY clause
	orderBy := "s.created_at DESC" // default
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "name":
			orderBy = "s.first_name, s.last_name"
		case "student_id":
			orderBy = "s.student_id"
		case "class":
			orderBy = "c.name"
		default:
			orderBy = "s.created_at DESC"
		}

		if filters.SortOrder == "desc" {
			orderBy += " DESC"
		} else {
			orderBy += " ASC"
		}
	}

	query += " ORDER BY " + orderBy + " LIMIT 100" // Limit for performance

	rows, err := db.Query(query, args...)
	if err != nil {
		return []models.Student{}, err
	}
	defer rows.Close()

	// Use a map to handle duplicate students (due to multiple parents)
	studentMap := make(map[string]*models.Student)

	for rows.Next() {
		var student models.Student
		var className, parentFirstName, parentLastName *string

		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
			&student.DateOfBirth, &student.Gender, &student.Address,
			&student.ClassID, &student.IsActive, &student.CreatedAt, &student.UpdatedAt,
			&className, &parentFirstName, &parentLastName,
		)
		if err != nil {
			continue
		}

		// Check if student already exists in map
		if existingStudent, exists := studentMap[student.ID]; exists {
			// Add parent to existing student if not already added
			if parentFirstName != nil && parentLastName != nil {
				parentExists := false
				parentName := *parentFirstName + " " + *parentLastName
				for _, p := range existingStudent.Parents {
					if p.FirstName+" "+p.LastName == parentName {
						parentExists = true
						break
					}
				}
				if !parentExists {
					parent := &models.Parent{
						FirstName: *parentFirstName,
						LastName:  *parentLastName,
					}
					existingStudent.Parents = append(existingStudent.Parents, parent)
				}
			}
		} else {
			// Create new student entry
			if className != nil {
				student.Class = &models.Class{
					Name: *className,
				}
				if student.ClassID != nil {
					student.Class.ID = *student.ClassID
				}
			}

			// Add parent if exists
			if parentFirstName != nil && parentLastName != nil {
				parent := &models.Parent{
					FirstName: *parentFirstName,
					LastName:  *parentLastName,
				}
				student.Parents = []*models.Parent{parent}
			}

			studentMap[student.ID] = &student
		}
	}

	// Convert map to slice
	var students []models.Student
	for _, student := range studentMap {
		students = append(students, *student)
	}

	return students, nil
}

// Helper function to get class names for students
func getClassNamesForStudents(db *sql.DB, students []models.Student) (map[string]string, error) {
	classMap := make(map[string]string)

	// Get unique class IDs
	classIDs := make(map[string]bool)
	for _, student := range students {
		if student.ClassID != nil {
			classIDs[*student.ClassID] = true
		}
	}

	if len(classIDs) == 0 {
		return classMap, nil
	}

	// Build query with placeholders
	placeholders := make([]string, 0, len(classIDs))
	args := make([]interface{}, 0, len(classIDs))
	i := 1
	for classID := range classIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, classID)
		i++
	}

	query := fmt.Sprintf("SELECT id, name FROM classes WHERE id IN (%s)",
		strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return classMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err == nil {
			classMap[id] = name
		}
	}

	return classMap, nil
}

// Helper function to get parents for students in batch
func getParentsForStudents(db *sql.DB, studentIDs []string) (map[string][]*models.Parent, error) {
	parentMap := make(map[string][]*models.Parent)

	if len(studentIDs) == 0 {
		return parentMap, nil
	}

	// Build query with placeholders
	placeholders := make([]string, len(studentIDs))
	args := make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT sp.student_id, p.id, p.first_name, p.last_name,
						  p.phone, p.email, sp.relationship
						  FROM student_parents sp
						  JOIN parents p ON sp.parent_id = p.id
						  WHERE sp.student_id IN (%s) AND p.is_active = true
						  ORDER BY sp.student_id, sp.created_at`,
		strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return parentMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var studentID string
		var parent models.Parent
		var relationship string

		err := rows.Scan(&studentID, &parent.ID, &parent.FirstName, &parent.LastName,
			&parent.Phone, &parent.Email, &relationship)
		if err != nil {
			continue
		}

		parentMap[studentID] = append(parentMap[studentID], &parent)
	}

	return parentMap, nil
}

// GetDashboardStats returns statistics for the dashboard (OPTIMIZED)
func GetDashboardStats(db *sql.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Single optimized query to get all student statistics
	query := `
		SELECT
			COUNT(*) as total_students,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_students,
			COUNT(CASE WHEN is_active = false THEN 1 END) as pending_applications,
			COUNT(CASE WHEN DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as new_this_month,
			COUNT(CASE WHEN gender = 'male' AND is_active = true THEN 1 END) as male_students,
			COUNT(CASE WHEN gender = 'female' AND is_active = true THEN 1 END) as female_students,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as recent_activity
		FROM students
	`

	var totalStudents, activeStudents, pendingApplications, newThisMonth int
	var maleStudents, femaleStudents, recentActivity int

	err := db.QueryRow(query).Scan(
		&totalStudents, &activeStudents, &pendingApplications, &newThisMonth,
		&maleStudents, &femaleStudents, &recentActivity,
	)
	if err != nil {
		return nil, err
	}

	// Set student statistics
	stats["total_students"] = totalStudents
	stats["active_students"] = activeStudents
	stats["pending_applications"] = pendingApplications
	stats["new_this_month"] = newThisMonth
	stats["male_students"] = maleStudents
	stats["female_students"] = femaleStudents
	stats["recent_activity"] = recentActivity

	// Get other statistics in parallel (these are typically small tables)
	// Total Parents
	var totalParents int
	err = db.QueryRow("SELECT COUNT(*) FROM parents WHERE is_active = true").Scan(&totalParents)
	if err != nil {
		return nil, err
	}
	stats["total_parents"] = totalParents

	// Total Classes
	var totalClasses int
	err = db.QueryRow("SELECT COUNT(*) FROM classes WHERE is_active = true").Scan(&totalClasses)
	if err != nil {
		return nil, err
	}
	stats["total_classes"] = totalClasses

	return stats, nil
}

// GetStudentsStats returns statistics specifically for the students page (OPTIMIZED)
func GetStudentsStats(db *sql.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Single optimized query to get all student statistics
	query := `
		SELECT
			COUNT(*) as total_students,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_students,
			COUNT(CASE WHEN is_active = false THEN 1 END) as pending_applications,
			COUNT(CASE WHEN DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as new_this_month,
			COUNT(CASE WHEN gender = 'male' AND is_active = true THEN 1 END) as male_students,
			COUNT(CASE WHEN gender = 'female' AND is_active = true THEN 1 END) as female_students,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as recent_activity
		FROM students
	`

	var totalStudents, activeStudents, pendingApplications, newThisMonth int
	var maleStudents, femaleStudents, recentActivity int

	err := db.QueryRow(query).Scan(
		&totalStudents, &activeStudents, &pendingApplications, &newThisMonth,
		&maleStudents, &femaleStudents, &recentActivity,
	)
	if err != nil {
		return nil, err
	}

	// Set student statistics
	stats["total_students"] = totalStudents
	stats["active_students"] = activeStudents
	stats["pending_applications"] = pendingApplications
	stats["new_this_month"] = newThisMonth
	stats["male_students"] = maleStudents
	stats["female_students"] = femaleStudents
	stats["recent_activity"] = recentActivity

	return stats, nil
}

// GetStudentsByYear gets all students for a specific year
func GetStudentsByYear(db *sql.DB, year int) ([]models.Student, error) {
	yearPrefix := fmt.Sprintf("STU-%d-%%", year)

	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth,
			  s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at
			  FROM students s
			  WHERE s.student_id LIKE $1 AND s.is_active = true
			  ORDER BY s.student_id ASC`

	rows, err := db.Query(query, yearPrefix)
	if err != nil {
		return []models.Student{}, nil
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student

		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
			&student.DateOfBirth, &student.Gender, &student.Address,
			&student.ClassID, &student.IsActive, &student.CreatedAt, &student.UpdatedAt,
		)
		if err != nil {
			continue
		}

		students = append(students, student)
	}
	return students, nil
}

// GetStudentsByClass gets all students for a specific class
func GetStudentsByClass(db *sql.DB, classID string) ([]*models.Student, error) {
	query := `SELECT id, student_id, first_name, last_name, date_of_birth, gender, address, class_id, is_active, created_at, updated_at
			  FROM students
			  WHERE class_id = $1 AND is_active = true
			  ORDER BY first_name, last_name`

	rows, err := db.Query(query, classID)
	if err != nil {
		return []*models.Student{}, nil
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		student := &models.Student{}
		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
			&student.DateOfBirth, &student.Gender, &student.Address, &student.ClassID,
			&student.IsActive, &student.CreatedAt, &student.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// GetStudentByID gets a single student by ID with details
func GetStudentByID(db *sql.DB, studentID string) (*models.Student, error) {
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth,
			  s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  WHERE s.id = $1 AND s.is_active = true`

	var student models.Student
	var className *string

	err := db.QueryRow(query, studentID).Scan(
		&student.ID, &student.StudentID, &student.FirstName, &student.LastName,
		&student.DateOfBirth, &student.Gender, &student.Address,
		&student.ClassID, &student.IsActive, &student.CreatedAt, &student.UpdatedAt,
		&className,
	)

	if err != nil {
		return nil, err
	}

	// Set class if exists
	if className != nil {
		student.Class = &models.Class{
			Name: *className,
		}
		if student.ClassID != nil {
			student.Class.ID = *student.ClassID
		}
	}

	// Get parents for this student
	parentQuery := `SELECT p.id, p.first_name, p.last_name, p.phone, p.email, sp.relationship
					FROM parents p
					INNER JOIN student_parents sp ON p.id = sp.parent_id
					WHERE sp.student_id = $1 AND p.is_active = true`

	rows, err := db.Query(parentQuery, studentID)
	if err == nil {
		defer rows.Close()
		var parents []*models.Parent
		for rows.Next() {
			parent := &models.Parent{}
			var relationship string
			err := rows.Scan(
				&parent.ID, &parent.FirstName, &parent.LastName,
				&parent.Phone, &parent.Email, &relationship,
			)
			if err == nil {
				parents = append(parents, parent)
			}
		}
		student.Parents = parents
	}

	return &student, nil
}

func CreateStudent(db *sql.DB, student *models.Student) error {
	query := `INSERT INTO students (student_id, first_name, last_name, date_of_birth,
			  gender, address, class_id)
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, student.StudentID, student.FirstName, student.LastName,
		student.DateOfBirth, student.Gender, student.Address,
		student.ClassID).Scan(&student.ID, &student.CreatedAt, &student.UpdatedAt)

	return err
}

// UpdateStudent updates an existing student in the database
func UpdateStudent(db *sql.DB, student *models.Student) error {
	query := `UPDATE students SET
			  first_name = $1, last_name = $2, date_of_birth = $3,
			  gender = $4, address = $5, class_id = $6, updated_at = NOW()
			  WHERE id = $7`

	_, err := db.Exec(query, student.FirstName, student.LastName, student.DateOfBirth,
		student.Gender, student.Address, student.ClassID, student.ID)

	return err
}

// DeleteStudent soft deletes a student by setting is_active to false
func DeleteStudent(db *sql.DB, studentID string) error {
	query := `UPDATE students SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, studentID)
	return err
}

// CheckStudentIDExists checks if a student ID already exists in the database
func CheckStudentIDExists(db *sql.DB, studentID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE student_id = $1 AND is_active = true)`
	err := db.QueryRow(query, studentID).Scan(&exists)
	return exists, err
}

func LinkStudentToParent(db *sql.DB, studentID string, parentID string, relationship string) error {
	query := `INSERT INTO student_parents (student_id, parent_id, relationship) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, studentID, parentID, relationship)
	return err
}

// UpdateStudentParent updates parent information for a student
func UpdateStudentParent(db *sql.DB, studentID string, parentID string, parentInfo interface{}) error {
	// First, update the parent record
	type ParentInfo struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		Relationship string `json:"relationship"`
	}

	parent, ok := parentInfo.(ParentInfo)
	if !ok {
		return fmt.Errorf("invalid parent info type")
	}

	// Update parent details
	var email, phone, address *string
	if parent.Email != "" {
		email = &parent.Email
	}
	if parent.Phone != "" {
		phone = &parent.Phone
	}
	if parent.Address != "" {
		address = &parent.Address
	}

	query := `UPDATE parents SET
			  first_name = $1, last_name = $2, email = $3, phone = $4, address = $5, updated_at = NOW()
			  WHERE id = $6`

	_, err := db.Exec(query, parent.FirstName, parent.LastName, email, phone, address, parentID)
	if err != nil {
		return err
	}

	// Update relationship if provided
	if parent.Relationship != "" {
		relationQuery := `UPDATE student_parents SET relationship = $1, updated_at = NOW()
						  WHERE student_id = $2 AND parent_id = $3`
		_, err = db.Exec(relationQuery, parent.Relationship, studentID, parentID)
	}

	return err
}

// CreateAndLinkParent creates a new parent and links them to a student
func CreateAndLinkParent(db *sql.DB, studentID string, parentInfo interface{}) error {
	type ParentInfo struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		Relationship string `json:"relationship"`
	}

	parent, ok := parentInfo.(ParentInfo)
	if !ok {
		return fmt.Errorf("invalid parent info type")
	}

	// Create parent
	var email, phone, address *string
	if parent.Email != "" {
		email = &parent.Email
	}
	if parent.Phone != "" {
		phone = &parent.Phone
	}
	if parent.Address != "" {
		address = &parent.Address
	}

	var parentID string
	query := `INSERT INTO parents (first_name, last_name, email, phone, address)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := db.QueryRow(query, parent.FirstName, parent.LastName, email, phone, address).Scan(&parentID)
	if err != nil {
		return err
	}

	// Link to student
	relationship := parent.Relationship
	if relationship == "" {
		relationship = "guardian" // Default relationship
	}

	return LinkStudentToParent(db, studentID, parentID, relationship)
}

// UpdateStudentParentRelationship updates the relationship between a student and parent
func UpdateStudentParentRelationship(db *sql.DB, studentID string, parentID string, relationship string) error {
	query := `UPDATE student_parents SET relationship = $1, updated_at = NOW()
			  WHERE student_id = $2 AND parent_id = $3`
	_, err := db.Exec(query, relationship, studentID, parentID)
	return err
}

// ChangeStudentParent changes the parent for a student (removes old, adds new)
func ChangeStudentParent(db *sql.DB, studentID string, newParentID string, relationship string) error {
	// First, remove any existing parent relationships for this student
	deleteQuery := `DELETE FROM student_parents WHERE student_id = $1`
	_, err := db.Exec(deleteQuery, studentID)
	if err != nil {
		return err
	}

	// If newParentID is provided, create new relationship
	if newParentID != "" {
		return LinkStudentToParent(db, studentID, newParentID, relationship)
	}

	return nil
}

// GetStudentParentRelationship gets the relationship between a student and parent
func GetStudentParentRelationship(db *sql.DB, studentID string, parentID string) (string, error) {
	var relationship string
	query := `SELECT relationship FROM student_parents
			  WHERE student_id = $1 AND parent_id = $2`
	err := db.QueryRow(query, studentID, parentID).Scan(&relationship)
	return relationship, err
}

// GetParentsForSelection returns all parents for selection with optional search
func GetParentsForSelection(db *sql.DB, search string) ([]*models.Parent, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT id, first_name, last_name, email, phone, address
				 FROM parents
				 WHERE is_active = true
				 AND (LOWER(first_name) LIKE LOWER($1)
				      OR LOWER(last_name) LIKE LOWER($1)
				      OR LOWER(email) LIKE LOWER($1)
				      OR LOWER(phone) LIKE LOWER($1))
				 ORDER BY first_name, last_name
				 LIMIT 50`
		args = append(args, "%"+search+"%")
	} else {
		query = `SELECT id, first_name, last_name, email, phone, address
				 FROM parents
				 WHERE is_active = true
				 ORDER BY first_name, last_name
				 LIMIT 50`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []*models.Parent
	for rows.Next() {
		parent := &models.Parent{}
		err := rows.Scan(
			&parent.ID, &parent.FirstName, &parent.LastName,
			&parent.Email, &parent.Phone, &parent.Address,
		)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}

	return parents, nil
}

func GetAllParents(db *sql.DB) ([]*models.Parent, error) {
	query := `SELECT id, first_name, last_name, phone, email, address, is_active, created_at, updated_at 
			  FROM parents WHERE is_active = true ORDER BY first_name, last_name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []*models.Parent
	for rows.Next() {
		parent := &models.Parent{}
		err := rows.Scan(
			&parent.ID, &parent.FirstName, &parent.LastName, &parent.Phone,
			&parent.Email, &parent.Address, &parent.IsActive, &parent.CreatedAt, &parent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}

	return parents, nil
}

func CreateParent(db *sql.DB, parent *models.Parent) error {
	query := `INSERT INTO parents (id, first_name, last_name, phone, email, address, is_active, created_at, updated_at) 
			  VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, true, NOW(), NOW()) 
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, parent.FirstName, parent.LastName, parent.Phone, parent.Email, parent.Address).Scan(
		&parent.ID, &parent.CreatedAt, &parent.UpdatedAt,
	)

	if err != nil {
		return err
	}

	parent.IsActive = true
	return nil
}

// SearchParents searches for parents by name, phone, or email
func SearchParents(db *sql.DB, query string) ([]*models.Parent, error) {
	searchPattern := "%" + query + "%"

	sqlQuery := `SELECT id, first_name, last_name, phone, email, address, is_active, created_at, updated_at
				 FROM parents
				 WHERE is_active = true AND (
					LOWER(first_name) LIKE LOWER($1)
					OR LOWER(last_name) LIKE LOWER($1)
					OR LOWER(CONCAT(first_name, ' ', last_name)) LIKE LOWER($1)
					OR phone LIKE $1
					OR LOWER(email) LIKE LOWER($1)
				 )
				 ORDER BY first_name, last_name
				 LIMIT 20`

	rows, err := db.Query(sqlQuery, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []*models.Parent
	for rows.Next() {
		parent := &models.Parent{}
		err := rows.Scan(
			&parent.ID, &parent.FirstName, &parent.LastName, &parent.Phone,
			&parent.Email, &parent.Address, &parent.IsActive, &parent.CreatedAt, &parent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}

	return parents, nil
}

func GetAllClasses(db *sql.DB) ([]*models.Class, error) {
	query := `SELECT c.id, c.name, c.code, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			  u.first_name, u.last_name, u.email
			  FROM classes c
			  LEFT JOIN users u ON c.teacher_id = u.id
			  WHERE c.is_active = true
			  ORDER BY c.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Class{}, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		var teacherFirstName, teacherLastName, teacherEmail sql.NullString

		err := rows.Scan(
			&class.ID, &class.Name, &class.Code, &class.TeacherID,
			&class.IsActive, &class.CreatedAt, &class.UpdatedAt,
			&teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			continue
		}

		// Set teacher info if exists
		if class.TeacherID != nil && teacherFirstName.Valid {
			class.Teacher = &models.User{
				ID:        *class.TeacherID,
				FirstName: teacherFirstName.String,
				LastName:  teacherLastName.String,
				Email:     teacherEmail.String,
			}
		}

		classes = append(classes, class)
	}

	if classes == nil {
		classes = []*models.Class{}
	}

	return classes, nil
}

func CreateClass(db *sql.DB, class *models.Class) error {
	var teacherID *string
	if class.TeacherID != nil && *class.TeacherID != "" {
		// Verify teacher exists
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", *class.TeacherID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("teacher not found")
		}
		teacherID = class.TeacherID
	}

	query := `INSERT INTO classes (name, code, teacher_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, class.Name, class.Code, teacherID).Scan(
		&class.ID, &class.CreatedAt, &class.UpdatedAt,
	)

	if err != nil {
		return err
	}

	class.IsActive = true
	return nil
}

func GetClassByID(db *sql.DB, classID string) (*models.Class, error) {
	query := `SELECT c.id, c.name, c.code, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			  u.first_name, u.last_name, u.email
			  FROM classes c
			  LEFT JOIN users u ON c.teacher_id = u.id
			  WHERE c.id = $1`

	class := &models.Class{}
	var teacherFirstName, teacherLastName, teacherEmail sql.NullString

	err := db.QueryRow(query, classID).Scan(
		&class.ID,
		&class.Name,
		&class.Code,
		&class.TeacherID,
		&class.IsActive,
		&class.CreatedAt,
		&class.UpdatedAt,
		&teacherFirstName,
		&teacherLastName,
		&teacherEmail,
	)

	if err != nil {
		return nil, err
	}

	if teacherFirstName.Valid {
		class.Teacher = &models.User{
			ID:        *class.TeacherID,
			FirstName: teacherFirstName.String,
			LastName:  teacherLastName.String,
			Email:     teacherEmail.String,
		}
	}

	return class, nil
}

func UpdateClass(db *sql.DB, class *models.Class) error {
	query := `UPDATE classes SET name = $1, code = $2, teacher_id = $3, updated_at = NOW()
			  WHERE id = $4`

	_, err := db.Exec(query, class.Name, class.Code, class.TeacherID, class.ID)
	return err
}

func DeleteClass(db *sql.DB, classID string) error {
	query := `UPDATE classes SET is_active = false, updated_at = NOW()
			  WHERE id = $1`

	_, err := db.Exec(query, classID)
	return err
}

// CheckClassExists checks if a class already exists in the database
func CheckClassExists(db *sql.DB, className string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM classes WHERE name = $1 AND is_active = true)`
	err := db.QueryRow(query, className).Scan(&exists)
	return exists, err
}

// Teacher-related functions
// Teacher-related functions
func GetAllTeachers(db *sql.DB) ([]*models.User, error) {
	query := `SELECT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  ORDER BY u.first_name, u.last_name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.User{}, nil
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Load roles separately to avoid complex joins
		roleQuery := `SELECT r.name FROM roles r 
					  INNER JOIN user_roles ur ON r.id = ur.role_id 
					  WHERE ur.user_id = $1`
		roleRows, err := db.Query(roleQuery, teacher.ID)
		if err == nil {
			for roleRows.Next() {
				var roleName string
				if err := roleRows.Scan(&roleName); err == nil {
					teacher.Roles = append(teacher.Roles, &models.Role{Name: roleName})
				}
			}
			roleRows.Close()
		}

		teachers = append(teachers, teacher)
	}

	if teachers == nil {
		teachers = []*models.User{}
	}

	return teachers, nil
}

func GetTeacherCountsByRole(db *sql.DB) (map[string]int, error) {
	query := `SELECT r.name, COUNT(DISTINCT u.id) as count
			  FROM roles r
			  INNER JOIN user_roles ur ON r.id = ur.role_id
			  INNER JOIN users u ON ur.user_id = u.id 
			  WHERE r.name IN ('class_teacher', 'subject_teacher') 
			  AND u.is_active = true
			  AND ur.deleted_at IS NULL
			  GROUP BY r.name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	// Initialize with 0 counts
	counts["class_teacher"] = 0
	counts["subject_teacher"] = 0
	
	for rows.Next() {
		var roleName string
		var count int
		if err := rows.Scan(&roleName, &count); err != nil {
			return nil, err
		}
		counts[roleName] = count
	}

	return counts, nil
}

func SearchTeachersWithPagination(db *sql.DB, query string, limit, offset int) ([]*models.User, int, error) {
	// Count total matching teachers
	countQuery := `SELECT COUNT(DISTINCT u.id) FROM users u 
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin') 
		AND u.is_active = true 
		AND (u.first_name ILIKE $1 OR u.last_name ILIKE $1 OR u.email ILIKE $1)`
	
	var total int
	err := db.QueryRow(countQuery, "%"+query+"%").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated teachers
	searchQuery := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
		FROM users u 
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin') 
		AND u.is_active = true 
		AND (u.first_name ILIKE $1 OR u.last_name ILIKE $1 OR u.email ILIKE $1)
		ORDER BY u.first_name, u.last_name
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(searchQuery, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(
			&teacher.ID,
			&teacher.Email,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.IsActive,
			&teacher.CreatedAt,
			&teacher.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		teachers = append(teachers, teacher)
	}

	if teachers == nil {
		teachers = []*models.User{}
	}

	return teachers, total, nil
}

func GetTeachersWithDetails(db *sql.DB) ([]*models.User, error) {
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at,
			  STRING_AGG(DISTINCT r.name, ', ') as roles,
			  COUNT(DISTINCT c.id) as class_count,
			  COUNT(DISTINCT s.id) as subject_count
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  LEFT JOIN classes c ON u.id = c.teacher_id AND c.is_active = true
			  LEFT JOIN papers p ON u.id = p.teacher_id AND p.is_active = true
			  LEFT JOIN subjects s ON p.subject_id = s.id AND s.is_active = true
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  GROUP BY u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
			  ORDER BY u.first_name, u.last_name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.User{}, nil
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		var roles string
		var classCount, subjectCount int
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt, &roles,
			&classCount, &subjectCount,
		)
		if err != nil {
			continue
		}
		teachers = append(teachers, teacher)
	}

	if teachers == nil {
		teachers = []*models.User{}
	}

	return teachers, nil
}

func CreateTeacher(db *sql.DB, user *models.User) error {
	// Hash password before storing
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}

	// Create user account
	query := `INSERT INTO users (email, password, first_name, last_name, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = db.QueryRow(query, user.Email, hashedPassword, user.FirstName, user.LastName).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Assign class_teacher role by default
	roleQuery := `INSERT INTO user_roles (user_id, role_id, created_at)
				  SELECT $1, r.id, NOW()
				  FROM roles r
				  WHERE r.name = 'class_teacher'`

	_, err = db.Exec(roleQuery, user.ID)
	if err != nil {
		return err
	}

	user.IsActive = true
	return nil
}

// GetTeachersByRole gets teachers filtered by specific role
func GetTeachersByRole(db *sql.DB, roleName string) ([]*models.User, error) {
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE r.name = $1 AND u.is_active = true
			  ORDER BY u.first_name, u.last_name`

	rows, err := db.Query(query, roleName)
	if err != nil {
		return []*models.User{}, nil
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
		)
		if err != nil {
			continue
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

// AssignTeacherRole assigns a role to a teacher
func AssignTeacherRole(db *sql.DB, teacherID string, roleName string) error {
	query := `INSERT INTO user_roles (user_id, role_id, created_at)
			  SELECT $1, r.id, NOW()
			  FROM roles r
			  WHERE r.name = $2
			  ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := db.Exec(query, teacherID, roleName)
	return err
}

// RemoveTeacherRole removes a role from a teacher
func RemoveTeacherRole(db *sql.DB, teacherID string, roleName string) error {
	query := `DELETE FROM user_roles 
			  WHERE user_id = $1 
			  AND role_id = (SELECT id FROM roles WHERE name = $2)`

	_, err := db.Exec(query, teacherID, roleName)
	return err
}

// GetTeacherClasses gets all classes assigned to a teacher
func GetTeacherClasses(db *sql.DB, teacherID string) ([]*models.Class, error) {
	query := `SELECT c.id, c.name, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			  COUNT(DISTINCT s.id) as student_count
			  FROM classes c
			  LEFT JOIN students s ON c.id = s.class_id AND s.is_active = true
			  WHERE c.teacher_id = $1 AND c.is_active = true
			  GROUP BY c.id, c.name, c.teacher_id, c.is_active, c.created_at, c.updated_at
			  ORDER BY c.name`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return []*models.Class{}, nil
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		var studentCount int
		err := rows.Scan(
			&class.ID, &class.Name, &class.TeacherID,
			&class.IsActive, &class.CreatedAt, &class.UpdatedAt, &studentCount,
		)
		if err != nil {
			continue
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// GetTeacherSubjects gets all subjects taught by a teacher
func GetTeacherSubjects(db *sql.DB, teacherID string) ([]*models.Subject, error) {
	query := `SELECT DISTINCT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  INNER JOIN papers p ON s.id = p.subject_id
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE p.teacher_id = $1 AND s.is_active = true AND p.is_active = true
			  ORDER BY s.name`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return []*models.Subject{}, nil
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName,
		)
		if err != nil {
			continue
		}

		if departmentName != nil && subject.DepartmentID != nil {
			subject.Department = &models.Department{
				ID:   *subject.DepartmentID,
				Name: *departmentName,
			}
		}

		subjects = append(subjects, subject)
	}

	return subjects, nil
}

// SearchTeachers searches for teachers by name or email
func SearchTeachers(db *sql.DB, searchTerm string) ([]*models.User, error) {
	searchPattern := "%" + searchTerm + "%"
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  AND (LOWER(u.first_name) LIKE LOWER($1)
			       OR LOWER(u.last_name) LIKE LOWER($1)
			       OR LOWER(u.email) LIKE LOWER($1)
			       OR LOWER(CONCAT(u.first_name, ' ', u.last_name)) LIKE LOWER($1))
			  ORDER BY u.first_name, u.last_name
			  LIMIT 20`

	rows, err := db.Query(query, searchPattern)
	if err != nil {
		return []*models.User{}, nil
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt,
		)
		if err != nil {
			continue
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

// UpdateTeacher updates an existing teacher's information
func UpdateTeacher(db *sql.DB, user *models.User) error {
	query := `UPDATE users
			  SET first_name = $1, last_name = $2, email = $3, updated_at = NOW()
			  WHERE id = $4 AND is_active = true`

	_, err := db.Exec(query, user.FirstName, user.LastName, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update teacher: %v", err)
	}

	return nil
}

func GetTeacherByID(db *sql.DB, teacherID string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, first_name, last_name, is_active, created_at, updated_at
			  FROM users WHERE id = $1 AND is_active = true`

	err := db.QueryRow(query, teacherID).Scan(
		&user.ID, &user.Email, &user.FirstName,
		&user.LastName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteTeacher soft deletes a teacher (sets is_active = false)
func DeleteTeacher(db *sql.DB, teacherID string) error {
	query := `UPDATE users
			  SET is_active = false, updated_at = NOW()
			  WHERE id = $1`

	result, err := db.Exec(query, teacherID)
	if err != nil {
		return fmt.Errorf("failed to delete teacher: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("teacher not found")
	}

	return nil
}

// Department-related functions
func GetAllDepartments(db *sql.DB) ([]*models.Department, error) {
	query := `SELECT d.id, d.name, d.code, d.description, d.head_of_department_id, d.assistant_head_id,
			  d.is_active, d.created_at, d.updated_at,
			  h.first_name as head_first_name, h.last_name as head_last_name, h.email as head_email,
			  a.first_name as assistant_first_name, a.last_name as assistant_last_name, a.email as assistant_email
			  FROM departments d
			  LEFT JOIN users h ON d.head_of_department_id = h.id
			  LEFT JOIN users a ON d.assistant_head_id = a.id
			  WHERE d.is_active = true ORDER BY d.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Department{}, nil
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		department := &models.Department{}
		var headFirstName, headLastName, headEmail *string
		var assistantFirstName, assistantLastName, assistantEmail *string

		err := rows.Scan(
			&department.ID, &department.Name, &department.Code, &department.Description,
			&department.HeadOfDepartmentID, &department.AssistantHeadID,
			&department.IsActive, &department.CreatedAt, &department.UpdatedAt,
			&headFirstName, &headLastName, &headEmail,
			&assistantFirstName, &assistantLastName, &assistantEmail,
		)
		if err != nil {
			continue
		}

		// Set head of department if exists
		if headFirstName != nil && headLastName != nil && department.HeadOfDepartmentID != nil {
			department.HeadOfDepartment = &models.User{
				ID:        *department.HeadOfDepartmentID,
				FirstName: *headFirstName,
				LastName:  *headLastName,
				Email:     *headEmail,
			}
		}

		// Set assistant head if exists
		if assistantFirstName != nil && assistantLastName != nil && department.AssistantHeadID != nil {
			department.AssistantHead = &models.User{
				ID:        *department.AssistantHeadID,
				FirstName: *assistantFirstName,
				LastName:  *assistantLastName,
				Email:     *assistantEmail,
			}
		}

		departments = append(departments, department)
	}

	if departments == nil {
		departments = []*models.Department{}
	}

	// Load subjects for each department
	for _, department := range departments {
		subjects, err := GetSubjectsByDepartment(db, department.ID)
		if err == nil {
			department.Subjects = subjects
		}
	}

	return departments, nil
}

func CreateDepartment(db *sql.DB, department *models.Department) error {
	query := `INSERT INTO departments (name, code, description, head_of_department_id, assistant_head_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, department.Name, department.Code, department.Description, department.HeadOfDepartmentID, department.AssistantHeadID).Scan(
		&department.ID, &department.CreatedAt, &department.UpdatedAt,
	)

	if err != nil {
		return err
	}

	department.IsActive = true
	return nil
}

func UpdateDepartment(db *sql.DB, department *models.Department) error {
	query := `UPDATE departments SET name = $1, code = $2, description = $3, head_of_department_id = $4, assistant_head_id = $5, updated_at = NOW() WHERE id = $6`
	_, err := db.Exec(query, department.Name, department.Code, department.Description, department.HeadOfDepartmentID, department.AssistantHeadID, department.ID)
	return err
}

func DeleteDepartment(db *sql.DB, departmentID string) error {
	query := `UPDATE departments SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, departmentID)
	return err
}

// Subject-related functions
func GetAllSubjects(db *sql.DB) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE s.is_active = true ORDER BY s.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Subject{}, nil
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName,
		)
		if err != nil {
			continue
		}

		// Set department if exists
		if departmentName != nil && subject.DepartmentID != nil {
			subject.Department = &models.Department{
				ID:   *subject.DepartmentID,
				Name: *departmentName,
			}
		}

		// Load papers for the subject
		papers, err := GetPapersBySubject(db, subject.ID)
		if err == nil {
			subject.Papers = papers
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

func SearchSubjects(db *sql.DB, query string) ([]*models.Subject, error) {
	sqlQuery := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
		FROM subjects s
		LEFT JOIN departments d ON s.department_id = d.id
		WHERE s.name ILIKE $1 OR s.code ILIKE $1
		ORDER BY s.name`
	
	rows, err := db.Query(sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string

		err := rows.Scan(
			&subject.ID,
			&subject.Name,
			&subject.Code,
			&subject.DepartmentID,
			&subject.IsActive,
			&subject.CreatedAt,
			&subject.UpdatedAt,
			&departmentName,
		)
		if err != nil {
			return nil, err
		}

		if departmentName != nil {
			subject.Department = &models.Department{
				Name: *departmentName,
			}
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

func GetSubjectByID(db *sql.DB, subjectID string) (*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE s.id = $1`

	subject := &models.Subject{}
	var departmentName *string

	err := db.QueryRow(query, subjectID).Scan(
		&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
		&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName,
	)
	if err != nil {
		return nil, err
	}

	// Set department if exists
	if departmentName != nil && subject.DepartmentID != nil {
		subject.Department = &models.Department{
			ID:   *subject.DepartmentID,
			Name: *departmentName,
		}
	}

	// Load papers for the subject
	papers, err := GetPapersBySubject(db, subject.ID)
	if err == nil {
		subject.Papers = papers
	}

	return subject, nil
}

func CreateSubject(db *sql.DB, subject *models.Subject) error {
	query := `INSERT INTO subjects (name, code, department_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, subject.Name, subject.Code, subject.DepartmentID).Scan(
		&subject.ID, &subject.CreatedAt, &subject.UpdatedAt,
	)

	if err != nil {
		return err
	}

	subject.IsActive = true
	return nil
}

func UpdateSubject(db *sql.DB, subject *models.Subject) error {
	query := `UPDATE subjects SET name = $1, code = $2, department_id = $3, is_active = $4, updated_at = NOW()
			  WHERE id = $5`

	_, err := db.Exec(query, subject.Name, subject.Code, subject.DepartmentID, subject.IsActive, subject.ID)
	return err
}

func DeleteSubject(db *sql.DB, subjectID string) error {
	query := `UPDATE subjects SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, subjectID)
	return err
}

func CreatePaper(db *sql.DB, paper *models.Paper) error {
	query := `INSERT INTO papers (subject_id, code, is_active, created_at, updated_at)
			  VALUES ($1, $2, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, paper.SubjectID, paper.Code).Scan(
		&paper.ID, &paper.CreatedAt, &paper.UpdatedAt,
	)

	if err != nil {
		return err
	}

	paper.IsActive = true
	return nil
}

func GetPaperByID(db *sql.DB, paperID string) (*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.id = $1`

	paper := &models.Paper{
		Subject: &models.Subject{},
	}
	
	err := db.QueryRow(query, paperID).Scan(
		&paper.ID, &paper.SubjectID, &paper.Code,
		&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
		&paper.Subject.Name, &paper.Subject.Code,
	)
	if err != nil {
		return nil, err
	}

	return paper, nil
}

func UpdatePaper(db *sql.DB, paper *models.Paper) error {
	query := `UPDATE papers SET subject_id = $1, code = $2, is_active = $3, updated_at = NOW()
			  WHERE id = $4`

	_, err := db.Exec(query, paper.SubjectID, paper.Code, paper.IsActive, paper.ID)
	return err
}

func DeletePaper(db *sql.DB, paperID string) error {
	query := `UPDATE papers SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, paperID)
	return err
}

// Get papers by subject
func GetPapersBySubject(db *sql.DB, subjectID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.subject_id = $1 AND p.is_active = true ORDER BY p.code`

	rows, err := db.Query(query, subjectID)
	if err != nil {
		return []*models.Paper{}, nil
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{
			Subject: &models.Subject{},
		}

		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Code,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
			&paper.Subject.Name, &paper.Subject.Code,
		)
		if err != nil {
			continue
		}

		papers = append(papers, paper)
	}

	if papers == nil {
		papers = []*models.Paper{}
	}

	return papers, nil
}

// GetAllPapers gets all papers with subject information
func GetAllPapers(db *sql.DB) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.is_active = true ORDER BY s.name, p.code`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Paper{}, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{
			Subject: &models.Subject{},
		}

		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Code,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
			&paper.Subject.Name, &paper.Subject.Code,
		)
		if err != nil {
			return nil, err
		}

		papers = append(papers, paper)
	}

	return papers, nil
}

// ClassPaper-related functions
func CreateClassPaper(db *sql.DB, classPaper *models.ClassPaper) error {
	query := `INSERT INTO class_papers (class_id, paper_id, teacher_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, classPaper.ClassID, classPaper.PaperID, classPaper.TeacherID).Scan(
		&classPaper.ID, &classPaper.CreatedAt, &classPaper.UpdatedAt,
	)

	if err != nil {
		return err
	}

	classPaper.IsActive = true
	return nil
}

func GetClassPapersByClass(db *sql.DB, classID string) ([]*models.ClassPaper, error) {
	query := `SELECT cp.id, cp.class_id, cp.paper_id, cp.teacher_id, cp.is_active, cp.created_at, cp.updated_at,
			  c.name as class_name,
			  p.code as paper_code,
			  u.first_name, u.last_name, u.email
			  FROM class_papers cp
			  LEFT JOIN classes c ON cp.class_id = c.id
			  LEFT JOIN papers p ON cp.paper_id = p.id
			  LEFT JOIN users u ON cp.teacher_id = u.id
			  WHERE cp.class_id = $1 AND cp.is_active = true ORDER BY p.code`

	rows, err := db.Query(query, classID)
	if err != nil {
		return []*models.ClassPaper{}, nil
	}
	defer rows.Close()

	var classPapers []*models.ClassPaper
	for rows.Next() {
		classPaper := &models.ClassPaper{
			Class:   &models.Class{},
			Paper:   &models.Paper{},
			Teacher: &models.User{},
		}
		var className, paperCode sql.NullString
		var teacherFirstName, teacherLastName, teacherEmail sql.NullString

		err := rows.Scan(
			&classPaper.ID, &classPaper.ClassID, &classPaper.PaperID, &classPaper.TeacherID,
			&classPaper.IsActive, &classPaper.CreatedAt, &classPaper.UpdatedAt,
			&className, &paperCode,
			&teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			continue
		}

		if className.Valid {
			classPaper.Class.Name = className.String
		}

		if paperCode.Valid {
			classPaper.Paper.Code = paperCode.String
		}

		if teacherFirstName.Valid {
			classPaper.Teacher.FirstName = teacherFirstName.String
			classPaper.Teacher.LastName = teacherLastName.String
			classPaper.Teacher.Email = teacherEmail.String
		} else {
			classPaper.Teacher = nil
		}

		classPapers = append(classPapers, classPaper)
	}

	if classPapers == nil {
		classPapers = []*models.ClassPaper{}
	}

	return classPapers, nil
}

func GetClassPapersByPaper(db *sql.DB, paperID string) ([]*models.ClassPaper, error) {
	query := `SELECT cp.id, cp.class_id, cp.paper_id, cp.teacher_id, cp.is_active, cp.created_at, cp.updated_at,
			  c.name as class_name,
			  p.code as paper_code,
			  u.first_name, u.last_name, u.email
			  FROM class_papers cp
			  LEFT JOIN classes c ON cp.class_id = c.id
			  LEFT JOIN papers p ON cp.paper_id = p.id
			  LEFT JOIN users u ON cp.teacher_id = u.id
			  WHERE cp.paper_id = $1 AND cp.is_active = true ORDER BY c.name`

	rows, err := db.Query(query, paperID)
	if err != nil {
		return []*models.ClassPaper{}, nil
	}
	defer rows.Close()

	var classPapers []*models.ClassPaper
	for rows.Next() {
		classPaper := &models.ClassPaper{
			Class:   &models.Class{},
			Paper:   &models.Paper{},
			Teacher: &models.User{},
		}
		var className, paperCode sql.NullString
		var teacherFirstName, teacherLastName, teacherEmail sql.NullString

		err := rows.Scan(
			&classPaper.ID, &classPaper.ClassID, &classPaper.PaperID, &classPaper.TeacherID,
			&classPaper.IsActive, &classPaper.CreatedAt, &classPaper.UpdatedAt,
			&className, &paperCode,
			&teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			continue
		}

		if className.Valid {
			classPaper.Class.Name = className.String
		}

		if paperCode.Valid {
			classPaper.Paper.Code = paperCode.String
		}

		if teacherFirstName.Valid {
			classPaper.Teacher.FirstName = teacherFirstName.String
			classPaper.Teacher.LastName = teacherLastName.String
			classPaper.Teacher.Email = teacherEmail.String
		} else {
			classPaper.Teacher = nil
		}

		classPapers = append(classPapers, classPaper)
	}

	if classPapers == nil {
		classPapers = []*models.ClassPaper{}
	}

	return classPapers, nil
}

func UpdateClassPaper(db *sql.DB, classPaper *models.ClassPaper) error {
	query := `UPDATE class_papers SET class_id = $1, paper_id = $2, teacher_id = $3, is_active = $4, updated_at = NOW()
			  WHERE id = $5`

	_, err := db.Exec(query, classPaper.ClassID, classPaper.PaperID, classPaper.TeacherID, classPaper.IsActive, classPaper.ID)
	return err
}

func DeleteClassPaper(db *sql.DB, classPaperID string) error {
	query := `UPDATE class_papers SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, classPaperID)
	return err
}

// Get subjects by department
func GetSubjectsByDepartment(db *sql.DB, departmentID string) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE s.department_id = $1 AND s.is_active = true ORDER BY s.name`

	rows, err := db.Query(query, departmentID)
	if err != nil {
		return []*models.Subject{}, nil
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName,
		)
		if err != nil {
			continue
		}

		// Set department if exists
		if departmentName != nil && subject.DepartmentID != nil {
			subject.Department = &models.Department{
				ID:   *subject.DepartmentID,
				Name: *departmentName,
			}
		}

		// Load papers for the subject
		papers, err := GetPapersBySubject(db, subject.ID)
		if err == nil {
			subject.Papers = papers
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

// GetAttendanceByClassAndDate gets attendance records for a specific class and date
func GetAttendanceByClassAndDate(db *sql.DB, classID string, date time.Time) ([]*models.Attendance, error) {
	query := `SELECT a.id, a.student_id, a.class_id, a.date, a.status, a.created_at, a.updated_at,
			  s.student_id as student_number, s.first_name, s.last_name
			  FROM attendance a
			  INNER JOIN students s ON a.student_id = s.id
			  WHERE a.class_id = $1 AND a.date = $2
			  ORDER BY s.first_name, s.last_name`

	rows, err := db.Query(query, classID, date)
	if err != nil {
		return []*models.Attendance{}, nil
	}
	defer rows.Close()

	var attendanceRecords []*models.Attendance
	for rows.Next() {
		attendance := &models.Attendance{
			Student: &models.Student{},
		}
		err := rows.Scan(
			&attendance.ID, &attendance.StudentID, &attendance.ClassID, &attendance.Date, &attendance.Status,
			&attendance.CreatedAt, &attendance.UpdatedAt,
			&attendance.Student.StudentID, &attendance.Student.FirstName, &attendance.Student.LastName,
		)
		if err != nil {
			return nil, err
		}
		attendanceRecords = append(attendanceRecords, attendance)
	}

	return attendanceRecords, nil
}

// GetAttendanceStats gets attendance statistics for a class within a date range
func GetAttendanceStats(db *sql.DB, classID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `SELECT
				COUNT(*) as total_records,
				COUNT(CASE WHEN status = 'present' THEN 1 END) as present_count,
				COUNT(CASE WHEN status = 'absent' THEN 1 END) as absent_count,
				COUNT(CASE WHEN status = 'late' THEN 1 END) as late_count
			  FROM attendance
			  WHERE class_id = $1 AND date BETWEEN $2 AND $3`

	var total, present, absent, late int
	err := db.QueryRow(query, classID, startDate, endDate).Scan(&total, &present, &absent, &late)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":   total,
		"present": present,
		"absent":  absent,
		"late":    late,
	}

	if total > 0 {
		stats["present_percentage"] = float64(present) / float64(total) * 100
		stats["absent_percentage"] = float64(absent) / float64(total) * 100
		stats["late_percentage"] = float64(late) / float64(total) * 100
	}

	return stats, nil
}

// CreateOrUpdateAttendance creates or updates an attendance record
func CreateOrUpdateAttendance(db *sql.DB, attendance *models.Attendance) error {
	// First check if attendance record exists for this student, class, and date
	var existingID string
	checkQuery := `SELECT id FROM attendance WHERE student_id = $1 AND class_id = $2 AND date = $3`
	err := db.QueryRow(checkQuery, attendance.StudentID, attendance.ClassID, attendance.Date).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new record
		insertQuery := `INSERT INTO attendance (student_id, class_id, date, status, created_at, updated_at)
						VALUES ($1, $2, $3, $4, NOW(), NOW())
						RETURNING id, created_at, updated_at`
		err = db.QueryRow(insertQuery, attendance.StudentID, attendance.ClassID, attendance.Date, attendance.Status).Scan(
			&attendance.ID, &attendance.CreatedAt, &attendance.UpdatedAt,
		)
		return err
	} else if err != nil {
		return err
	} else {
		// Update existing record
		updateQuery := `UPDATE attendance SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err = db.Exec(updateQuery, attendance.Status, existingID)
		attendance.ID = existingID
		return err
	}
}

// GetAllAcademicYears gets all academic years
func GetAllAcademicYears(db *sql.DB) ([]*models.AcademicYear, error) {
	query := `SELECT id, name, start_date, end_date, is_current, is_active, created_at, updated_at
			  FROM academic_years 
			  ORDER BY start_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.AcademicYear{}, nil
	}
	defer rows.Close()

	var academicYears []*models.AcademicYear
	for rows.Next() {
		academicYear := &models.AcademicYear{}
		var startDate, endDate time.Time
		err := rows.Scan(
			&academicYear.ID, &academicYear.Name, &startDate, &endDate,
			&academicYear.IsCurrent, &academicYear.IsActive, &academicYear.CreatedAt, &academicYear.UpdatedAt,
		)
		if err != nil {
			continue
		}
		// Convert time.Time to CustomTime
		academicYear.StartDate = models.CustomTime{Time: startDate}
		academicYear.EndDate = models.CustomTime{Time: endDate}

		// Load associated terms
		terms, err := GetTermsByAcademicYearID(db, academicYear.ID)
		if err == nil {
			academicYear.Terms = terms
		}

		academicYears = append(academicYears, academicYear)
	}

	if academicYears == nil {
		academicYears = []*models.AcademicYear{}
	}

	return academicYears, nil
}

// GetAcademicYearByID gets an academic year by ID
func GetAcademicYearByID(db *sql.DB, academicYearID string) (*models.AcademicYear, error) {
	academicYear := &models.AcademicYear{}
	var startDate, endDate time.Time
	query := `SELECT id, name, start_date, end_date, is_current, is_active, created_at, updated_at
			  FROM academic_years WHERE id = $1`

	err := db.QueryRow(query, academicYearID).Scan(
		&academicYear.ID, &academicYear.Name, &startDate, &endDate,
		&academicYear.IsCurrent, &academicYear.IsActive, &academicYear.CreatedAt, &academicYear.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Convert time.Time to CustomTime
	academicYear.StartDate = models.CustomTime{Time: startDate}
	academicYear.EndDate = models.CustomTime{Time: endDate}

	// Load associated terms
	terms, err := GetTermsByAcademicYearID(db, academicYearID)
	if err == nil {
		academicYear.Terms = terms
	}

	return academicYear, nil
}

// CreateAcademicYear creates a new academic year
func CreateAcademicYear(db *sql.DB, academicYear *models.AcademicYear) error {
	query := `INSERT INTO academic_years (name, start_date, end_date, is_current, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, academicYear.Name, academicYear.StartDate.Time, academicYear.EndDate.Time,
		academicYear.IsCurrent, academicYear.IsActive).Scan(
		&academicYear.ID, &academicYear.CreatedAt, &academicYear.UpdatedAt,
	)

	return err
}

// UpdateAcademicYear updates an existing academic year
func UpdateAcademicYear(db *sql.DB, academicYear *models.AcademicYear) error {
	query := `UPDATE academic_years SET name = $1, start_date = $2, end_date = $3, 
			  is_current = $4, is_active = $5, updated_at = NOW() WHERE id = $6`
	_, err := db.Exec(query, academicYear.Name, academicYear.StartDate.Time, academicYear.EndDate.Time,
		academicYear.IsCurrent, academicYear.IsActive, academicYear.ID)
	return err
}

// DeleteAcademicYear deletes an academic year
func DeleteAcademicYear(db *sql.DB, academicYearID string) error {
	query := `DELETE FROM academic_years WHERE id = $1`
	_, err := db.Exec(query, academicYearID)
	return err
}

// GetTermsByAcademicYearID gets all terms for an academic year
func GetTermsByAcademicYearID(db *sql.DB, academicYearID string) ([]*models.Term, error) {
	query := `SELECT id, academic_year_id, name, start_date, end_date, 
			  is_current, is_active, created_at, updated_at
			  FROM terms 
			  WHERE academic_year_id = $1
			  ORDER BY start_date`

	rows, err := db.Query(query, academicYearID)
	if err != nil {
		return []*models.Term{}, nil
	}
	defer rows.Close()

	var terms []*models.Term
	for rows.Next() {
		term := &models.Term{}
		var startDate, endDate time.Time
		err := rows.Scan(
			&term.ID, &term.AcademicYearID, &term.Name, &startDate, &endDate,
			&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt,
		)
		if err != nil {
			continue
		}
		// Convert time.Time to CustomTime
		term.StartDate = models.CustomTime{Time: startDate}
		term.EndDate = models.CustomTime{Time: endDate}
		terms = append(terms, term)
	}

	if terms == nil {
		terms = []*models.Term{}
	}

	return terms, nil
}

// CreateTerm creates a new term
func CreateTerm(db *sql.DB, term *models.Term) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If this term is being set as current, make all other terms not current
	if term.IsCurrent {
		_, err = tx.Exec("UPDATE terms SET is_current = false")
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive).Scan(
		&term.ID, &term.CreatedAt, &term.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateTerm updates an existing term
func UpdateTerm(db *sql.DB, term *models.Term) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If this term is being set as current, make all other terms not current
	if term.IsCurrent {
		_, err = tx.Exec("UPDATE terms SET is_current = false WHERE id != $1", term.ID)
		if err != nil {
			return err
		}
	}

	query := `UPDATE terms SET academic_year_id = $1, name = $2, start_date = $3, end_date = $4,
			  is_current = $5, is_active = $6, updated_at = NOW() WHERE id = $7`
	_, err = tx.Exec(query, term.AcademicYearID, term.Name, term.StartDate.Time, term.EndDate.Time,
		term.IsCurrent, term.IsActive, term.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteTerm deletes a term
func DeleteTerm(db *sql.DB, termID string) error {
	query := `DELETE FROM terms WHERE id = $1`
	_, err := db.Exec(query, termID)
	return err
}

// Get all terms
func GetAllTerms(db *sql.DB) ([]*models.Term, error) {
	query := `SELECT t.id, t.academic_year_id, t.name, t.start_date, t.end_date, 
			  t.is_current, t.is_active, t.created_at, t.updated_at,
			  a.name as academic_year_name
			  FROM terms t
			  LEFT JOIN academic_years a ON t.academic_year_id = a.id
			  ORDER BY t.start_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Term{}, nil
	}
	defer rows.Close()

	var terms []*models.Term
	for rows.Next() {
		term := &models.Term{
			AcademicYear: &models.AcademicYear{},
		}
		var academicYearName *string
		var startDate, endDate time.Time
		err := rows.Scan(
			&term.ID, &term.AcademicYearID, &term.Name, &startDate, &endDate,
			&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt,
			&academicYearName,
		)
		if err != nil {
			continue
		}

		// Convert time.Time to CustomTime
		term.StartDate = models.CustomTime{Time: startDate}
		term.EndDate = models.CustomTime{Time: endDate}

		if academicYearName != nil {
			term.AcademicYear.Name = *academicYearName
		}

		terms = append(terms, term)
	}

	if terms == nil {
		terms = []*models.Term{}
	}

	return terms, nil
}

func GetTermByID(db *sql.DB, termID string) (*models.Term, error) {
	term := &models.Term{
		AcademicYear: &models.AcademicYear{},
	}
	query := `SELECT t.id, t.academic_year_id, t.name, t.start_date, t.end_date, 
			  t.is_current, t.is_active, t.created_at, t.updated_at,
			  a.name as academic_year_name
			  FROM terms t
			  LEFT JOIN academic_years a ON t.academic_year_id = a.id
			  WHERE t.id = $1`

	var academicYearName *string
	var startDate, endDate time.Time
	err := db.QueryRow(query, termID).Scan(
		&term.ID, &term.AcademicYearID, &term.Name, &startDate, &endDate,
		&term.IsCurrent, &term.IsActive, &term.CreatedAt, &term.UpdatedAt,
		&academicYearName,
	)

	if err != nil {
		return nil, err
	}

	// Convert time.Time to CustomTime
	term.StartDate = models.CustomTime{Time: startDate}
	term.EndDate = models.CustomTime{Time: endDate}

	if academicYearName != nil {
		term.AcademicYear.Name = *academicYearName
	}

	return term, nil
}

// SetCurrentAcademicYear sets one academic year as current and all others as not current
func SetCurrentAcademicYear(db *sql.DB, academicYearID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set all academic years as not current
	_, err = tx.Exec("UPDATE academic_years SET is_current = false")
	if err != nil {
		return err
	}

	// Set the specified academic year as current
	_, err = tx.Exec("UPDATE academic_years SET is_current = true WHERE id = $1", academicYearID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SetCurrentTerm sets one term as current and all others as not current
func SetCurrentTerm(db *sql.DB, termID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set all terms as not current
	_, err = tx.Exec("UPDATE terms SET is_current = false")
	if err != nil {
		return err
	}

	// Set the specified term as current
	_, err = tx.Exec("UPDATE terms SET is_current = true WHERE id = $1", termID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AutoSetCurrentAcademicYear automatically sets current academic year based on current date
func AutoSetCurrentAcademicYear(db *sql.DB) error {
	query := `UPDATE academic_years SET is_current = (NOW() BETWEEN start_date AND end_date)`
	_, err := db.Exec(query)
	return err
}

// AutoSetCurrentTerm automatically sets current term based on current date
func AutoSetCurrentTerm(db *sql.DB) error {
	query := `UPDATE terms SET is_current = (NOW() BETWEEN start_date AND end_date)`
	_, err := db.Exec(query)
	return err
}

// GetTeachersStats returns statistics for the teachers page cards
func GetTeachersStats(db *sql.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Single optimized query to get all teacher statistics
	query := `
		SELECT
			COUNT(DISTINCT u.id) as total_teachers,
			COUNT(DISTINCT CASE WHEN u.is_active = true THEN u.id END) as active_teachers,
			COUNT(DISTINCT CASE WHEN r.name = 'class_teacher' AND u.is_active = true THEN u.id END) as class_teachers,
			COUNT(DISTINCT CASE WHEN r.name = 'subject_teacher' AND u.is_active = true THEN u.id END) as subject_teachers
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
		AND ur.deleted_at IS NULL
	`

	var totalTeachers, activeTeachers, classTeachers, subjectTeachers int

	err := db.QueryRow(query).Scan(
		&totalTeachers, &activeTeachers, &classTeachers, &subjectTeachers,
	)
	if err != nil {
		return nil, err
	}

	stats["total_teachers"] = totalTeachers
	stats["active_teachers"] = activeTeachers
	stats["class_teachers"] = classTeachers
	stats["subject_teachers"] = subjectTeachers

	return stats, nil
}

// GetClassSubjects returns all subjects assigned to a class with their papers
func GetClassSubjects(db *sql.DB, classID string) ([]*models.Subject, error) {
	query := `
		SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at
		FROM subjects s
		INNER JOIN class_subjects cs ON s.id = cs.subject_id
		WHERE cs.class_id = $1 AND s.is_active = true AND cs.deleted_at IS NULL
		ORDER BY s.name
	`

	rows, err := db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code,
			&subject.DepartmentID, &subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get papers for this subject in this class
		paperQuery := `
			SELECT p.id, p.code, p.subject_id, p.is_active, p.created_at, p.updated_at
			FROM papers p
			INNER JOIN class_papers cp ON p.id = cp.paper_id
			WHERE cp.class_id = $1 AND p.subject_id = $2 AND p.is_active = true AND cp.is_active = true
			ORDER BY p.code
		`
		
		paperRows, err := db.Query(paperQuery, classID, subject.ID)
		if err == nil {
			var papers []*models.Paper
			for paperRows.Next() {
				paper := &models.Paper{}
				err := paperRows.Scan(
					&paper.ID, &paper.Code, &paper.SubjectID, 
					&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
				)
				if err == nil {
					papers = append(papers, paper)
				}
			}
			paperRows.Close()
			subject.Papers = papers
		}

		subjects = append(subjects, subject)
	}

	return subjects, nil
}

// GetClassPapers returns papers assigned to a specific class
func GetClassPapers(db *sql.DB, classID string) ([]*models.Paper, error) {
	query := `
		SELECT p.id, p.code, p.subject_id, p.is_active, p.created_at, p.updated_at
		FROM papers p
		INNER JOIN class_papers cp ON p.id = cp.paper_id
		WHERE cp.class_id = $1 AND p.is_active = true AND cp.is_active = true
		ORDER BY p.code
	`

	rows, err := db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		err := rows.Scan(
			&paper.ID, &paper.Code,
			&paper.SubjectID, &paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		papers = append(papers, paper)
	}

	return papers, nil
}

// PaperAssignment represents a paper assignment with teacher
type PaperAssignment struct {
	PaperID   string  `json:"paper_id"`
	TeacherID *string `json:"teacher_id"`
}

// GetSubjectPapersForClass returns papers for a subject with assignment status
func GetSubjectPapersForClass(db *sql.DB, classID, subjectID string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			p.id, p.code, p.subject_id,
			CASE WHEN cp.id IS NOT NULL THEN true ELSE false END as is_assigned,
			cp.teacher_id,
			u.first_name, u.last_name
		FROM papers p
		LEFT JOIN class_papers cp ON p.id = cp.paper_id AND cp.class_id = $1 AND cp.is_active = true
		LEFT JOIN users u ON cp.teacher_id = u.id
		WHERE p.subject_id = $2 AND p.is_active = true
		ORDER BY p.code
	`

	rows, err := db.Query(query, classID, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []map[string]interface{}
	for rows.Next() {
		var id, code, subjectId string
		var isAssigned bool
		var teacherId, firstName, lastName *string
		
		err := rows.Scan(&id, &code, &subjectId, &isAssigned, &teacherId, &firstName, &lastName)
		if err != nil {
			return nil, err
		}

		paper := map[string]interface{}{
			"id":          id,
			"code":        code,
			"subject_id":  subjectId,
			"is_assigned": isAssigned,
			"teacher_id":  teacherId,
		}

		if firstName != nil && lastName != nil {
			paper["teacher_name"] = *firstName + " " + *lastName
		}

		papers = append(papers, paper)
	}

	return papers, nil
}

// AssignPapersToClass assigns papers to a class for a specific subject
func AssignPapersToClass(db *sql.DB, classID, subjectID string, paperAssignments []PaperAssignment) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, remove existing assignments for this subject in this class
	_, err = tx.Exec(`
		UPDATE class_papers 
		SET is_active = false 
		WHERE class_id = $1 AND paper_id IN (
			SELECT id FROM papers WHERE subject_id = $2
		)
	`, classID, subjectID)
	if err != nil {
		return err
	}

	// Then add new assignments
	for _, assignment := range paperAssignments {
		_, err = tx.Exec(`
			INSERT INTO class_papers (class_id, paper_id, teacher_id, is_active) 
			VALUES ($1, $2, $3, true)
			ON CONFLICT (class_id, paper_id) 
			DO UPDATE SET teacher_id = $3, is_active = true
		`, classID, assignment.PaperID, assignment.TeacherID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddSubjectsToClass adds multiple subjects to a class
func AddSubjectsToClass(db *sql.DB, classID string, subjectIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if class exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1 AND is_active = true)", classID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("class not found")
	}

	// Insert class-subject relationships
	for _, subjectID := range subjectIDs {
		// Check if relationship already exists
		var relationExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM class_subjects 
			WHERE class_id = $1 AND subject_id = $2 AND deleted_at IS NULL)
		`, classID, subjectID).Scan(&relationExists)
		if err != nil {
			return err
		}

		if !relationExists {
			_, err = tx.Exec(`
				INSERT INTO class_subjects (class_id, subject_id, created_at)
				VALUES ($1, $2, NOW())
			`, classID, subjectID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// RemoveSubjectFromClass removes a subject from a class
func RemoveSubjectFromClass(db *sql.DB, classID, subjectID string) error {
	query := `
		UPDATE class_subjects 
		SET deleted_at = NOW() 
		WHERE class_id = $1 AND subject_id = $2 AND deleted_at IS NULL
	`

	result, err := db.Exec(query, classID, subjectID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subject not found in class or already removed")
	}

	return nil
}

// GetClassStatistics gets detailed statistics for a specific class
func GetClassStatistics(db *sql.DB, classID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total students
	var totalStudents int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM students 
		WHERE class_id = $1 AND is_active = true
	`, classID).Scan(&totalStudents)
	if err != nil {
		return nil, err
	}

	// Get male students
	var maleStudents int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM students 
		WHERE class_id = $1 AND is_active = true AND gender = 'male'
	`, classID).Scan(&maleStudents)
	if err != nil {
		maleStudents = 0
	}

	// Get female students
	var femaleStudents int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM students 
		WHERE class_id = $1 AND is_active = true AND gender = 'female'
	`, classID).Scan(&femaleStudents)
	if err != nil {
		femaleStudents = 0
	}

	stats["total_students"] = totalStudents
	stats["male_students"] = maleStudents
	stats["female_students"] = femaleStudents

	return stats, nil
}

// GetClassStudents gets all active students for a specific class with accurate filtering
func GetClassStudents(db *sql.DB, classID string) ([]map[string]interface{}, error) {
	query := `
		SELECT id, student_id, first_name, last_name, date_of_birth, gender, address 
		FROM students 
		WHERE class_id = $1 AND is_active = true 
		ORDER BY first_name, last_name
	`

	rows, err := db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []map[string]interface{}
	for rows.Next() {
		var id, studentID, firstName, lastName, gender *string
		var dateOfBirth *time.Time
		var address *string
		
		err := rows.Scan(&id, &studentID, &firstName, &lastName, &dateOfBirth, &gender, &address)
		if err != nil {
			continue
		}
		
		// Skip students with missing required fields
		if id == nil || firstName == nil || lastName == nil {
			continue
		}
		
		addressStr := ""
		if address != nil {
			addressStr = *address
		}
		
		dateStr := ""
		if dateOfBirth != nil {
			dateStr = dateOfBirth.Format("2006-01-02")
		}
		
		genderStr := ""
		if gender != nil {
			genderStr = *gender
		}
		
		student := map[string]interface{}{
			"id":            *id,
			"student_id":    studentID,
			"first_name":    *firstName,
			"last_name":     *lastName,
			"date_of_birth": dateStr,
			"gender":        genderStr,
			"address":       addressStr,
		}
		
		students = append(students, student)
	}

	return students, nil
}
