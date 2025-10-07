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

// GetStudentsWithFilters gets students with filtering, searching, and sorting support
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
	query := `SELECT c.id, c.name, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			  u.first_name, u.last_name, u.email,
			  COUNT(DISTINCT s.id) as student_count,
			  COUNT(DISTINCT cs.subject_id) as subject_count
			  FROM classes c
			  LEFT JOIN users u ON c.teacher_id = u.id
			  LEFT JOIN students s ON c.id = s.class_id AND s.is_active = true
			  LEFT JOIN class_subjects cs ON c.id = cs.class_id
			  WHERE c.is_active = true
			  GROUP BY c.id, c.name, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			           u.first_name, u.last_name, u.email
			  ORDER BY c.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Class{}, nil
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		var teacherID *int
		var teacherFirstName, teacherLastName, teacherEmail *string
		var studentCount, subjectCount int

		err := rows.Scan(
			&class.ID, &class.Name, &teacherID,
			&class.IsActive, &class.CreatedAt, &class.UpdatedAt,
			&teacherFirstName, &teacherLastName, &teacherEmail,
			&studentCount, &subjectCount,
		)
		if err != nil {
			continue
		}

		// Convert teacher ID to string if exists
		if teacherID != nil {
			teacherIDStr := fmt.Sprintf("%d", *teacherID)
			class.TeacherID = &teacherIDStr

			// Set teacher info if exists
			if teacherFirstName != nil && teacherLastName != nil {
				class.Teacher = &models.User{
					ID:        teacherIDStr,
					FirstName: *teacherFirstName,
					LastName:  *teacherLastName,
					Email:     *teacherEmail,
				}
			}
		}

		// Create dummy slices for template compatibility
		if studentCount > 0 {
			class.Students = make([]*models.Student, studentCount)
		}
		if subjectCount > 0 {
			class.Subjects = make([]*models.Subject, subjectCount)
		}

		classes = append(classes, class)
	}

	if classes == nil {
		classes = []*models.Class{}
	}

	return classes, nil
}

func CreateClass(db *sql.DB, class *models.Class) error {
	var teacherID *int
	if class.TeacherID != nil && *class.TeacherID != "" {
		// Convert string teacher ID to integer
		var tid int
		if err := db.QueryRow("SELECT id FROM users WHERE id = $1", *class.TeacherID).Scan(&tid); err != nil {
			return err
		}
		teacherID = &tid
	}

	query := `INSERT INTO classes (name, teacher_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, class.Name, teacherID).Scan(
		&class.ID, &class.CreatedAt, &class.UpdatedAt,
	)

	if err != nil {
		return err
	}

	class.IsActive = true
	return nil
}

// Teacher-related functions
func GetAllTeachers(db *sql.DB) ([]*models.User, error) {
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at,
			  STRING_AGG(DISTINCT r.name, ', ') as roles,
			  d.name as department_name,
			  STRING_AGG(DISTINCT c.name, ', ') as class_names
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  LEFT JOIN departments d ON (d.head_of_department_id = u.id OR d.assistant_head_id = u.id)
			  LEFT JOIN classes c ON c.teacher_id = u.id AND c.is_active = true
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  GROUP BY u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at, d.name
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
		var departmentName *string
		var classNames *string
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt, &roles, &departmentName, &classNames,
		)
		if err != nil {
			continue
		}

		if departmentName != nil {
			teacher.Department = &models.Department{Name: *departmentName}
		}

		if roles != "" {
			roleNames := strings.Split(roles, ", ")
			for _, roleName := range roleNames {
				teacher.Roles = append(teacher.Roles, &models.Role{Name: roleName})
			}
		}

		if classNames != nil && *classNames != "" {
			names := strings.Split(*classNames, ", ")
			for _, name := range names {
				teacher.Classes = append(teacher.Classes, &models.Class{Name: name})
			}
		}

		teachers = append(teachers, teacher)
	}

	if teachers == nil {
		teachers = []*models.User{}
	}

	return teachers, nil
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

// Paper-related functions
func GetAllPapers(db *sql.DB) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.teacher_id, p.name, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code,
			  u.first_name, u.last_name, u.email
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  LEFT JOIN users u ON p.teacher_id = u.id
			  WHERE p.is_active = true ORDER BY s.name, p.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Paper{}, nil
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{
			Subject: &models.Subject{},
			Teacher: &models.User{},
		}
		var teacherFirstName, teacherLastName, teacherEmail sql.NullString

		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.TeacherID, &paper.Name, &paper.Code,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
			&paper.Subject.Name, &paper.Subject.Code,
			&teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			return nil, err
		}

		if teacherFirstName.Valid {
			paper.Teacher.FirstName = teacherFirstName.String
			paper.Teacher.LastName = teacherLastName.String
			paper.Teacher.Email = teacherEmail.String
		} else {
			paper.Teacher = nil
		}

		papers = append(papers, paper)
	}

	return papers, nil
}

func CreatePaper(db *sql.DB, paper *models.Paper) error {
	query := `INSERT INTO papers (subject_id, teacher_id, name, code, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err := db.QueryRow(query, paper.SubjectID, paper.TeacherID, paper.Name, paper.Code).Scan(
		&paper.ID, &paper.CreatedAt, &paper.UpdatedAt,
	)

	if err != nil {
		return err
	}

	paper.IsActive = true
	return nil
}

func GetPaperByID(db *sql.DB, paperID string) (*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.teacher_id, p.name, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code,
			  u.first_name, u.last_name, u.email
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  LEFT JOIN users u ON p.teacher_id = u.id
			  WHERE p.id = $1`

	paper := &models.Paper{
		Subject: &models.Subject{},
		Teacher: &models.User{},
	}
	var teacherFirstName, teacherLastName, teacherEmail sql.NullString

	err := db.QueryRow(query, paperID).Scan(
		&paper.ID, &paper.SubjectID, &paper.TeacherID, &paper.Name, &paper.Code,
		&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
		&paper.Subject.Name, &paper.Subject.Code,
		&teacherFirstName, &teacherLastName, &teacherEmail,
	)
	if err != nil {
		return nil, err
	}

	if teacherFirstName.Valid {
		paper.Teacher.FirstName = teacherFirstName.String
		paper.Teacher.LastName = teacherLastName.String
		paper.Teacher.Email = teacherEmail.String
	} else {
		paper.Teacher = nil
	}

	return paper, nil
}

func UpdatePaper(db *sql.DB, paper *models.Paper) error {
	query := `UPDATE papers SET subject_id = $1, teacher_id = $2, name = $3, code = $4, is_active = $5, updated_at = NOW()
			  WHERE id = $6`

	_, err := db.Exec(query, paper.SubjectID, paper.TeacherID, paper.Name, paper.Code, paper.IsActive, paper.ID)
	return err
}

func DeletePaper(db *sql.DB, paperID string) error {
	query := `UPDATE papers SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, paperID)
	return err
}

// Get papers by subject
func GetPapersBySubject(db *sql.DB, subjectID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.teacher_id, p.name, p.code, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code,
			  u.first_name, u.last_name, u.email
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  LEFT JOIN users u ON p.teacher_id = u.id
			  WHERE p.subject_id = $1 AND p.is_active = true ORDER BY p.name`

	rows, err := db.Query(query, subjectID)
	if err != nil {
		return []*models.Paper{}, nil
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{
			Subject: &models.Subject{},
			Teacher: &models.User{},
		}
		var teacherFirstName, teacherLastName, teacherEmail sql.NullString

		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.TeacherID, &paper.Name, &paper.Code,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
			&paper.Subject.Name, &paper.Subject.Code,
			&teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			continue
		}

		if teacherFirstName.Valid {
			paper.Teacher.FirstName = teacherFirstName.String
			paper.Teacher.LastName = teacherLastName.String
			paper.Teacher.Email = teacherEmail.String
		} else {
			paper.Teacher = nil
		}

		papers = append(papers, paper)
	}

	if papers == nil {
		papers = []*models.Paper{}
	}

	return papers, nil
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
