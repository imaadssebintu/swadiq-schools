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
	ClassIDs  string // Support multiple class IDs as comma-separated string
	Gender    string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
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

func GetUserByID(db *sql.DB, userID string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password, first_name, last_name, COALESCE(phone, ''), is_active, created_at, updated_at
			  FROM users WHERE id = $1 AND is_active = true`

	err := db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
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

func CreateSession(db *sql.DB, sessionID interface{}, userID string, expiresAt time.Time) error {
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

func CreatePasswordResetToken(db *sql.DB, email, token string) error {
	query := `INSERT INTO password_reset_tokens (email, token, expires_at, created_at) 
			  VALUES ($1, $2, NOW() + INTERVAL '24 hours', NOW())
			  ON CONFLICT (email) DO UPDATE SET 
			  token = EXCLUDED.token, 
			  expires_at = EXCLUDED.expires_at, 
			  created_at = EXCLUDED.created_at`
	_, err := db.Exec(query, email, token)
	return err
}

func ValidatePasswordResetToken(db *sql.DB, token string) (string, error) {
	var email string
	query := `SELECT email FROM password_reset_tokens 
			  WHERE token = $1 AND expires_at > NOW() AND (used_at IS NULL OR used_at IS NOT NULL)`
	err := db.QueryRow(query, token).Scan(&email)
	return email, err
}

func MarkPasswordResetTokenAsUsed(db *sql.DB, token string) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE token = $1`
	_, err := db.Exec(query, token)
	return err
}

// CreateTeacher creates a new teacher with department assignment
func CreateTeacher(db *sql.DB, user *models.User, departmentID *string) error {
	// Hash password before storing
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}

	// Create user account
	query := `INSERT INTO users (email, password, first_name, last_name, phone, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	err = db.QueryRow(query, user.Email, hashedPassword, user.FirstName, user.LastName, user.Phone).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Assign to department if provided
	if departmentID != nil {
		deptQuery := `INSERT INTO user_departments (user_id, department_id) VALUES ($1, $2)`
		_, err = db.Exec(deptQuery, user.ID, *departmentID)
		if err != nil {
			return err
		}
	}

	// Note: class_teacher role will be assigned only when teacher is assigned to a class
	// No default role assignment here

	user.IsActive = true
	return nil
}

// GetAllTeachers gets all teachers with their department information
func GetAllTeachers(db *sql.DB) ([]*models.User, error) {
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, COALESCE(u.phone, ''), u.is_active, u.created_at, u.updated_at,
			  STRING_AGG(DISTINCT r.name, ', ') as roles,
			  STRING_AGG(DISTINCT d.name, ', ') as department_names,
			  STRING_AGG(DISTINCT c.name, ', ') as class_names
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  LEFT JOIN user_departments ud ON u.id = ud.user_id
			  LEFT JOIN departments d ON ud.department_id = d.id
			  LEFT JOIN (
				  SELECT teacher_id, id, name FROM classes WHERE is_active = true
				  UNION
				  SELECT te.teacher_id, c.id, c.name 
				  FROM timetable_entries te
				  JOIN classes c ON te.class_id = c.id
				  WHERE te.is_active = true AND c.is_active = true
			  ) c ON c.teacher_id = u.id
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  GROUP BY u.id, u.email, u.first_name, u.last_name, u.phone, u.is_active, u.created_at, u.updated_at
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
		var departmentNames *string
		var classNames *string
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName, &teacher.Phone,
			&teacher.IsActive, &teacher.CreatedAt, &teacher.UpdatedAt, &roles, &departmentNames, &classNames,
		)
		if err != nil {
			continue
		}

		if departmentNames != nil && *departmentNames != "" {
			names := strings.Split(*departmentNames, ", ")
			for _, name := range names {
				teacher.Departments = append(teacher.Departments, &models.Department{Name: name})
			}
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

// GetTeacherCountsByRole gets teacher counts grouped by role
func GetTeacherCountsByRole(db *sql.DB) (map[string]int, error) {
	query := `SELECT r.name, COUNT(DISTINCT u.id) as count
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
			  AND u.is_active = true
			  GROUP BY r.name`

	rows, err := db.Query(query)
	if err != nil {
		return make(map[string]int), err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var roleName string
		var count int
		if err := rows.Scan(&roleName, &count); err == nil {
			counts[roleName] = count
		}
	}

	return counts, nil
}

func GetTeachersBySubjectOrPaper(db *sql.DB, subjectID, paperID string) ([]*models.User, error) {
	var query string
	var args []interface{}

	baseQuery := `SELECT DISTINCT u.id, u.first_name, u.last_name, u.email
				  FROM users u
				  INNER JOIN teacher_subjects ts ON u.id = ts.teacher_id
				  WHERE u.is_active = true`

	if paperID != "" {
		query = baseQuery + " AND ts.paper_id = $1 ORDER BY u.first_name"
		args = append(args, paperID)
	} else if subjectID != "" {
		query = baseQuery + " AND ts.subject_id = $1 ORDER BY u.first_name"
		args = append(args, subjectID)
	} else {
		// No filter, return all teachers with roles
		query = `SELECT u.id, u.first_name, u.last_name, u.email
				 FROM users u
				 INNER JOIN user_roles ur ON u.id = ur.user_id
				 INNER JOIN roles r ON ur.role_id = r.id
				 WHERE u.is_active = true AND r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin')
				 ORDER BY u.first_name`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*models.User
	for rows.Next() {
		teacher := &models.User{}
		if err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email); err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

// GetTeacherSubjects fetches all subjects assigned to a teacher
func GetTeacherSubjects(db *sql.DB, teacherID string) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code
			  FROM subjects s
			  INNER JOIN teacher_subjects ts ON s.id = ts.subject_id
			  WHERE ts.teacher_id = $1 AND s.is_active = true
			  ORDER BY s.name`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		s := &models.Subject{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Code); err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}
	return subjects, nil
}

// GetTeacherClasses fetches all classes assigned to a teacher (direct and via timetable)
func GetTeacherClasses(db *sql.DB, teacherID string) ([]*models.Class, error) {
	query := `SELECT id, name, code
			  FROM (
				  SELECT id, name, code FROM classes WHERE teacher_id = $1 AND is_active = true
				  UNION
				  SELECT DISTINCT c.id, c.name, c.code FROM classes c
				  INNER JOIN timetable_entries te ON c.id = te.class_id
				  WHERE te.teacher_id = $1 AND te.is_active = true AND c.is_active = true
			  ) t
			  ORDER BY name`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		c := &models.Class{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Code); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

// SearchTeachersWithPagination searches teachers with pagination
func SearchTeachersWithPagination(db *sql.DB, searchTerm string, limit, offset int) ([]*models.User, int, error) {
	searchPattern := "%" + searchTerm + "%"

	// Count query
	countQuery := `SELECT COUNT(DISTINCT u.id)
				   FROM users u
				   INNER JOIN user_roles ur ON u.id = ur.user_id
				   INNER JOIN roles r ON ur.role_id = r.id
				   WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
				   AND u.is_active = true
				   AND (LOWER(u.first_name) LIKE LOWER($1)
						OR LOWER(u.last_name) LIKE LOWER($1)
						OR LOWER(u.email) LIKE LOWER($1)
						OR LOWER(CONCAT(u.first_name, ' ', u.last_name)) LIKE LOWER($1))`

	var total int
	err := db.QueryRow(countQuery, searchPattern).Scan(&total)
	if err != nil {
		return []*models.User{}, 0, err
	}

	// Data query
	dataQuery := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
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
				  LIMIT $2 OFFSET $3`

	rows, err := db.Query(dataQuery, searchPattern, limit, offset)
	if err != nil {
		return []*models.User{}, total, err
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

	return teachers, total, nil
}

// GetTeacherByID gets a teacher by ID
func GetTeacherByID(db *sql.DB, teacherID string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, first_name, last_name, COALESCE(phone, ''), is_active, created_at, updated_at
			  FROM users WHERE id = $1 AND is_active = true`

	err := db.QueryRow(query, teacherID).Scan(
		&user.ID, &user.Email, &user.FirstName,
		&user.LastName, &user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateTeacher updates an existing teacher's information
func UpdateTeacher(db *sql.DB, user *models.User) error {
	query := `UPDATE users
			  SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = NOW()
			  WHERE id = $5 AND is_active = true`

	_, err := db.Exec(query, user.FirstName, user.LastName, user.Email, user.Phone, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update teacher: %v", err)
	}

	return nil
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

// IsPhoneTaken checks if a phone number is already in use by another active user
func IsPhoneTaken(db *sql.DB, phone string, excludeUserID string) (bool, error) {
	if phone == "" {
		return false, nil
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND id != $2 AND is_active = true)`
	err := db.QueryRow(query, phone, excludeUserID).Scan(&exists)
	return exists, err
}

// GetAllDepartments gets all departments
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

	return departments, nil
}

// GetAllSubjects gets all subjects with paper and class counts
func GetAllSubjects(db *sql.DB) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name,
			  COALESCE(p.paper_count, 0) as paper_count,
			  COALESCE(cs.class_count, 0) as class_count
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as paper_count 
				  FROM papers 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) p ON s.id = p.subject_id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as class_count 
				  FROM class_subjects 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) cs ON s.id = cs.subject_id
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
		var paperCount, classCount int
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName, &paperCount, &classCount,
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

		// Create dummy slices for template compatibility
		if paperCount > 0 {
			subject.Papers = make([]*models.Paper, paperCount)
		}
		if classCount > 0 {
			subject.Classes = make([]*models.Class, classCount)
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

// GetSubjectsByDepartment gets subjects by department with paper and class counts
func GetSubjectsByDepartment(db *sql.DB, departmentID string) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  COALESCE(p.paper_count, 0) as paper_count,
			  COALESCE(cs.class_count, 0) as class_count
			  FROM subjects s
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as paper_count 
				  FROM papers 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) p ON s.id = p.subject_id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as class_count 
				  FROM class_subjects 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) cs ON s.id = cs.subject_id
			  WHERE s.department_id = $1 AND s.is_active = true 
			  ORDER BY s.name`

	rows, err := db.Query(query, departmentID)
	if err != nil {
		return []*models.Subject{}, nil
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var paperCount, classCount int
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &paperCount, &classCount,
		)
		if err != nil {
			continue
		}

		// Create dummy slices for template compatibility
		if paperCount > 0 {
			subject.Papers = make([]*models.Paper, paperCount)
		}
		if classCount > 0 {
			subject.Classes = make([]*models.Class, classCount)
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

// LinkTeacherToSubjects links a teacher to multiple subjects
func LinkTeacherToSubjects(db *sql.DB, teacherID string, subjectIDs []string) error {
	if len(subjectIDs) == 0 {
		return nil // Nothing to link
	}

	// Build the insert query dynamically
	valueStrings := make([]string, 0, len(subjectIDs))
	valueArgs := make([]interface{}, 0, len(subjectIDs)*2) // teacherID + subjectID for each

	for i, subjectID := range subjectIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, teacherID, subjectID)
	}

	query := fmt.Sprintf("INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES %s ON CONFLICT (teacher_id, subject_id) DO NOTHING",
		strings.Join(valueStrings, ","))

	_, err := db.Exec(query, valueArgs...)
	return err
}

// LinkTeacherToDepartments links a teacher to multiple departments
func LinkTeacherToDepartments(db *sql.DB, teacherID string, departmentIDs []string) error {
	if len(departmentIDs) == 0 {
		return nil // Nothing to link
	}

	// Build the insert query dynamically
	valueStrings := make([]string, 0, len(departmentIDs))
	valueArgs := make([]interface{}, 0, len(departmentIDs)*2) // teacherID + departmentID for each

	for i, departmentID := range departmentIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, teacherID, departmentID)
	}

	query := fmt.Sprintf("INSERT INTO user_departments (user_id, department_id) VALUES %s ON CONFLICT (user_id, department_id) DO NOTHING",
		strings.Join(valueStrings, ","))

	_, err := db.Exec(query, valueArgs...)
	return err
}

// GetUserDepartments gets departments for a user
func GetUserDepartments(db *sql.DB, userID string) ([]*models.Department, error) {
	query := `SELECT d.id, d.name, d.code, d.description
			  FROM departments d
			  INNER JOIN user_departments ud ON d.id = ud.department_id
			  WHERE ud.user_id = $1 AND d.is_active = true
			  ORDER BY d.name`

	rows, err := db.Query(query, userID)
	if err != nil {
		return []*models.Department{}, err
	}
	defer rows.Close()

	var departments []*models.Department
	for rows.Next() {
		department := &models.Department{}
		err := rows.Scan(&department.ID, &department.Name, &department.Code, &department.Description)
		if err != nil {
			continue
		}
		departments = append(departments, department)
	}

	if departments == nil {
		departments = []*models.Department{}
	}

	return departments, nil
}

// Placeholder functions to resolve compilation errors
func GetAllParents(db *sql.DB) ([]*models.Parent, error) {
	return []*models.Parent{}, nil
}

func CreateParent(db *sql.DB, parent *models.Parent) error {
	return nil
}

func SearchParents(db *sql.DB, query string) ([]*models.Parent, error) {
	return []*models.Parent{}, nil
}

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
			continue
		}
		students = append(students, student)
	}

	if students == nil {
		students = []*models.Student{}
	}

	return students, nil
}

func GetAttendanceByClassAndDate(db *sql.DB, classID string, date time.Time) ([]*models.Attendance, error) {
	query := `SELECT id, student_id, class_id, timetable_entry_id, paper_id, term_id, date, status, marked_by, created_at, updated_at
			  FROM attendance
			  WHERE class_id = $1 AND date = $2 AND deleted_at IS NULL`

	rows, err := db.Query(query, classID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.Attendance
	for rows.Next() {
		record := &models.Attendance{}
		var status string
		err := rows.Scan(
			&record.ID, &record.StudentID, &record.ClassID, &record.TimetableEntryID,
			&record.PaperID, &record.TermID, &record.Date, &status, &record.MarkedBy,
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			continue
		}
		record.Status = models.AttendanceStatus(status)
		records = append(records, record)
	}

	return records, nil
}

func CreateOrUpdateAttendance(db *sql.DB, attendance *models.Attendance) error {
	var id string
	var err error
	if attendance.TimetableEntryID != nil {
		err = db.QueryRow("SELECT id FROM attendance WHERE student_id = $1 AND date = $2 AND timetable_entry_id = $3 AND deleted_at IS NULL",
			attendance.StudentID, attendance.Date, *attendance.TimetableEntryID).Scan(&id)
	} else if attendance.ClassID != nil {
		err = db.QueryRow("SELECT id FROM attendance WHERE student_id = $1 AND date = $2 AND class_id = $3 AND timetable_entry_id IS NULL AND deleted_at IS NULL",
			attendance.StudentID, attendance.Date, *attendance.ClassID).Scan(&id)
	} else {
		return fmt.Errorf("either class_id or timetable_entry_id must be provided")
	}

	if err == nil {
		// Update
		query := "UPDATE attendance SET status = $1, marked_by = $2, term_id = COALESCE($4, term_id), updated_at = NOW() WHERE id = $3"
		_, err = db.Exec(query, attendance.Status, attendance.MarkedBy, id, attendance.TermID)
		return err
	} else if err == sql.ErrNoRows {
		// Insert
		query := `INSERT INTO attendance (id, student_id, class_id, timetable_entry_id, paper_id, term_id, date, status, marked_by, created_at, updated_at)
				  VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
		_, err = db.Exec(query, attendance.StudentID, attendance.ClassID, attendance.TimetableEntryID, attendance.PaperID, attendance.TermID, attendance.Date, attendance.Status, attendance.MarkedBy)
		return err
	}

	return err
}

func GetAttendanceStats(db *sql.DB, classID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func GetAllClasses(db *sql.DB) ([]*models.Class, error) {
	// Query with student count
	query := `SELECT c.id, c.name, c.code, c.teacher_id, c.is_active, c.created_at, c.updated_at,
			  u.first_name, u.last_name, u.email,
			  COALESCE(s.student_count, 0) as student_count
			  FROM classes c
			  LEFT JOIN users u ON c.teacher_id = u.id
			  LEFT JOIN (
				  SELECT class_id, COUNT(*) as student_count 
				  FROM students 
				  WHERE is_active = true 
				  GROUP BY class_id
			  ) s ON c.id = s.class_id
			  WHERE c.is_active = true
			  ORDER BY c.name`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Class{}, nil
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		var teacherFirstName, teacherLastName, teacherEmail *string
		var studentCount int
		err := rows.Scan(
			&class.ID, &class.Name, &class.Code, &class.TeacherID, &class.IsActive, &class.CreatedAt, &class.UpdatedAt,
			&teacherFirstName, &teacherLastName, &teacherEmail, &studentCount,
		)
		if err != nil {
			continue
		}

		// Set teacher if exists
		if teacherFirstName != nil && teacherLastName != nil && class.TeacherID != nil {
			class.Teacher = &models.User{
				ID:        *class.TeacherID,
				FirstName: *teacherFirstName,
				LastName:  *teacherLastName,
				Email:     *teacherEmail,
			}
		}

		// Set student count
		class.StudentCount = studentCount

		classes = append(classes, class)
	}

	if classes == nil {
		classes = []*models.Class{}
	}

	return classes, nil
}

func GetAllPapers(db *sql.DB) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.deleted_at IS NULL
			  ORDER BY p.name, p.code`

	rows, err := db.Query(query)
	if err != nil {
		return []*models.Paper{}, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		var subjectName, subjectCode *string
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory, &paper.IsActive,
			&paper.CreatedAt, &paper.UpdatedAt, &subjectName, &subjectCode,
		)
		if err != nil {
			continue
		}

		// Set subject if exists
		if subjectName != nil {
			paper.Subject = &models.Subject{
				ID:   paper.SubjectID,
				Name: *subjectName,
				Code: *subjectCode,
			}
		}

		papers = append(papers, paper)
	}

	if papers == nil {
		papers = []*models.Paper{}
	}

	return papers, nil
}

// GetPapersStats returns optimized stats for papers page
func GetPapersStats(db *sql.DB) (map[string]interface{}, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM papers WHERE deleted_at IS NULL) as total_papers,
			(SELECT COUNT(*) FROM papers WHERE deleted_at IS NULL AND is_active = true) as active_papers,
			(SELECT COUNT(*) FROM subjects WHERE is_active = true) as subjects_count,
			(SELECT COUNT(DISTINCT u.id) FROM users u 
			 INNER JOIN user_roles ur ON u.id = ur.user_id 
			 INNER JOIN roles r ON ur.role_id = r.id 
			 WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher') 
			 AND u.is_active = true) as teachers_count
	`

	var totalPapers, activePapers, subjectsCount, teachersCount int
	err := db.QueryRow(query).Scan(&totalPapers, &activePapers, &subjectsCount, &teachersCount)
	if err != nil {
		return nil, err
	}

	// Get subjects for dropdown
	subjects, err := GetAllSubjects(db)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalPapers":   totalPapers,
		"activePapers":  activePapers,
		"subjectsCount": subjectsCount,
		"teachersCount": teachersCount,
		"subjects":      subjects,
	}, nil
}

func CreateClass(db *sql.DB, class *models.Class) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if teacher is already assigned to another class
	if class.TeacherID != nil && *class.TeacherID != "" {
		var existingClassID string
		err := tx.QueryRow("SELECT id FROM classes WHERE teacher_id = $1 AND is_active = true LIMIT 1", *class.TeacherID).Scan(&existingClassID)
		if err == nil {
			return fmt.Errorf("teacher is already assigned to another class")
		}
	}

	query := `INSERT INTO classes (name, code, teacher_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	class.IsActive = true
	err = tx.QueryRow(query, class.Name, class.Code, class.TeacherID, class.IsActive).Scan(
		&class.ID, &class.CreatedAt, &class.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// If teacher is assigned, add class_teacher role
	if class.TeacherID != nil && *class.TeacherID != "" {
		err = assignClassTeacherRoleInDB(tx, *class.TeacherID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetClassStatistics(db *sql.DB, classID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total students
	var totalStudents int
	err := db.QueryRow("SELECT COUNT(*) FROM students WHERE class_id = $1 AND is_active = true", classID).Scan(&totalStudents)
	if err != nil {
		totalStudents = 0
	}

	// Male students
	var maleStudents int
	err = db.QueryRow("SELECT COUNT(*) FROM students WHERE class_id = $1 AND is_active = true AND gender = 'male'", classID).Scan(&maleStudents)
	if err != nil {
		maleStudents = 0
	}

	// Female students
	var femaleStudents int
	err = db.QueryRow("SELECT COUNT(*) FROM students WHERE class_id = $1 AND is_active = true AND gender = 'female'", classID).Scan(&femaleStudents)
	if err != nil {
		femaleStudents = 0
	}

	stats["total_students"] = totalStudents
	stats["male_students"] = maleStudents
	stats["female_students"] = femaleStudents

	return stats, nil
}

func GetClassStudents(db *sql.DB, classID string) ([]*models.Student, error) {
	return GetStudentsByClass(db, classID)
}

func AddSubjectsToClass(db *sql.DB, classID string, subjectIDs []string) error {
	return nil
}

type SubjectAssignment struct {
	SubjectID    string `json:"subject_id"`
	IsCompulsory bool   `json:"is_compulsory"`
}

func AddSubjectsToClassWithCompulsory(db *sql.DB, classID string, subjects []SubjectAssignment) error {
	if len(subjects) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(subjects))
	valueArgs := make([]interface{}, 0, len(subjects)*3)

	for i, subject := range subjects {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, classID, subject.SubjectID, subject.IsCompulsory)
	}

	query := fmt.Sprintf("INSERT INTO class_subjects (class_id, subject_id, is_compulsory) VALUES %s ON CONFLICT (class_id, subject_id) DO UPDATE SET is_compulsory = EXCLUDED.is_compulsory",
		strings.Join(valueStrings, ","))

	_, err := db.Exec(query, valueArgs...)
	return err
}

type PaperAssignmentForSubject struct {
	PaperID   string  `json:"paper_id"`
	TeacherID *string `json:"teacher_id"`
}

type SubjectAssignmentWithPapers struct {
	SubjectID        string                      `json:"subject_id"`
	IsCompulsory     bool                        `json:"is_compulsory"`
	PaperAssignments []PaperAssignmentForSubject `json:"paper_assignments"`
}

func AddSubjectsToClassWithPapers(db *sql.DB, classID string, subjects []SubjectAssignmentWithPapers) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, subject := range subjects {
		// Add subject to class
		query := `INSERT INTO class_subjects (class_id, subject_id, is_compulsory) 
					VALUES ($1, $2, $3) 
					ON CONFLICT (class_id, subject_id) 
					DO UPDATE SET is_compulsory = EXCLUDED.is_compulsory, deleted_at = NULL`
		_, err := tx.Exec(query, classID, subject.SubjectID, subject.IsCompulsory)
		if err != nil {
			return err
		}

		// Add paper assignments
		for _, pa := range subject.PaperAssignments {
			var existingID string
			checkQuery := `SELECT id FROM class_papers WHERE class_id = $1 AND paper_id = $2 AND deleted_at IS NULL`
			err := tx.QueryRow(checkQuery, classID, pa.PaperID).Scan(&existingID)

			if err != nil {
				// Create new class paper
				insertQuery := `INSERT INTO class_papers (class_id, paper_id, teacher_id, created_at, updated_at)
						  VALUES ($1, $2, $3, NOW(), NOW())`
				_, err = tx.Exec(insertQuery, classID, pa.PaperID, pa.TeacherID)
				if err != nil {
					return err
				}
			} else {
				// Update existing class paper
				updateQuery := `UPDATE class_papers SET teacher_id = $1, updated_at = NOW() WHERE id = $2`
				_, err = tx.Exec(updateQuery, pa.TeacherID, existingID)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func GetClassSubjects(db *sql.DB, classID string) ([]*models.Subject, error) {
	// Query to get subjects with their papers and teachers
	query := `SELECT DISTINCT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name, COALESCE(cs.is_compulsory, true) as is_compulsory
			  FROM subjects s
			  INNER JOIN class_subjects cs ON s.id = cs.subject_id
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE cs.class_id = $1 AND s.is_active = true AND cs.deleted_at IS NULL
			  ORDER BY s.name`

	rows, err := db.Query(query, classID)
	if err != nil {
		// If the query fails (possibly due to missing is_compulsory column), try without it
		fallbackQuery := `SELECT DISTINCT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
					  d.name as department_name
					  FROM subjects s
					  INNER JOIN class_subjects cs ON s.id = cs.subject_id
					  LEFT JOIN departments d ON s.department_id = d.id
					  WHERE cs.class_id = $1 AND s.is_active = true AND cs.deleted_at IS NULL
					  ORDER BY s.name`

		rows, err = db.Query(fallbackQuery, classID)
		if err != nil {
			return []*models.Subject{}, err
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

			// Load papers and teachers for this subject - we'll handle this in the API response
			// subject.Papers = loadSubjectPapersWithTeachers(db, classID, subject.ID)

			subjects = append(subjects, subject)
		}

		return subjects, nil
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string
		var isCompulsory bool
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName, &isCompulsory,
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

		// Load papers and teachers for this subject - we'll handle this in the API response
		// subject.Papers = loadSubjectPapersWithTeachers(db, classID, subject.ID)

		subjects = append(subjects, subject)
	}

	return subjects, nil
}

// ClassPaperWithTeacher represents a paper in a class with its assigned teacher
type ClassPaperWithTeacher struct {
	ID           string       `json:"id"`
	SubjectID    string       `json:"subject_id"`
	Name         string       `json:"name"`
	Code         string       `json:"code"`
	IsCompulsory bool         `json:"is_compulsory"`
	IsActive     bool         `json:"is_active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Teacher      *models.User `json:"teacher,omitempty"`
}

// loadSubjectPapersWithTeachers loads papers for a specific subject in a class with teacher information
func loadSubjectPapersWithTeachers(db *sql.DB, classID, subjectID string) []ClassPaperWithTeacher {
	// First get all papers for the subject
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at,
			  cp.teacher_id, u.first_name as teacher_first_name, u.last_name as teacher_last_name, u.email as teacher_email
			  FROM papers p
			  LEFT JOIN class_papers cp ON p.id = cp.paper_id AND cp.class_id = $1 AND cp.deleted_at IS NULL
			  LEFT JOIN users u ON cp.teacher_id = u.id AND u.is_active = true
			  WHERE p.subject_id = $2 AND p.deleted_at IS NULL
			  ORDER BY p.code`

	rows, err := db.Query(query, classID, subjectID)
	if err != nil {
		return []ClassPaperWithTeacher{}
	}
	defer rows.Close()

	var papers []ClassPaperWithTeacher
	for rows.Next() {
		paper := ClassPaperWithTeacher{}
		var teacherID, teacherFirstName, teacherLastName, teacherEmail *string
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory, &paper.IsActive,
			&paper.CreatedAt, &paper.UpdatedAt, &teacherID, &teacherFirstName, &teacherLastName, &teacherEmail,
		)
		if err != nil {
			continue
		}

		// Set teacher if exists
		if teacherID != nil && teacherFirstName != nil && teacherLastName != nil {
			paper.Teacher = &models.User{
				ID:        *teacherID,
				FirstName: *teacherFirstName,
				LastName:  *teacherLastName,
				Email:     *teacherEmail,
			}
		}

		papers = append(papers, paper)
	}

	return papers
}

// SubjectWithPapers represents a subject with its papers and teachers for a specific class
type SubjectWithPapers struct {
	*models.Subject
	Papers []ClassPaperWithTeacher `json:"papers"`
}

// GetClassSubjectsWithPapers gets subjects for a class with their papers and assigned teachers
func GetClassSubjectsWithPapers(db *sql.DB, classID string) ([]SubjectWithPapers, error) {
	// Get subjects assigned to the class (either directly or via papers)
	query := `SELECT DISTINCT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  LEFT JOIN class_subjects cs ON s.id = cs.subject_id AND cs.class_id = $1 AND cs.deleted_at IS NULL
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE (cs.class_id = $1 OR EXISTS (
				  SELECT 1 FROM class_papers cp 
				  INNER JOIN papers p ON cp.paper_id = p.id 
				  WHERE cp.class_id = $1 AND p.subject_id = s.id AND cp.deleted_at IS NULL AND p.deleted_at IS NULL
			  )) AND s.is_active = true
			  ORDER BY s.name`

	rows, err := db.Query(query, classID)
	if err != nil {
		return []SubjectWithPapers{}, err
	}
	defer rows.Close()

	var result []SubjectWithPapers
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

		// Load papers with teacher assignments for this subject
		papers := loadSubjectPapersWithTeachers(db, classID, subject.ID)

		subjectWithPapers := SubjectWithPapers{
			Subject: subject,
			Papers:  papers,
		}
		result = append(result, subjectWithPapers)
	}

	return result, nil
}

func GetClassPapers(db *sql.DB, classID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  INNER JOIN class_papers cp ON p.id = cp.paper_id
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE cp.class_id = $1 AND p.deleted_at IS NULL AND cp.deleted_at IS NULL
			  ORDER BY s.name, p.code`

	rows, err := db.Query(query, classID)
	if err != nil {
		return []*models.Paper{}, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		var subjectName, subjectCode *string
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory, &paper.IsActive,
			&paper.CreatedAt, &paper.UpdatedAt, &subjectName, &subjectCode,
		)
		if err != nil {
			continue
		}

		// Set subject if exists
		if subjectName != nil {
			paper.Subject = &models.Subject{
				ID:   paper.SubjectID,
				Name: *subjectName,
				Code: *subjectCode,
			}
		}

		papers = append(papers, paper)
	}

	return papers, nil
}

type PaperAssignment struct {
	PaperID   string  `json:"paper_id"`
	TeacherID *string `json:"teacher_id"`
}

func AssignPapersToClass(db *sql.DB, classID string, assignments []PaperAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	// First delete existing assignments for this class
	_, err := db.Exec("DELETE FROM class_papers WHERE class_id = $1", classID)
	if err != nil {
		return fmt.Errorf("failed to clear existing assignments: %v", err)
	}

	valueStrings := make([]string, 0, len(assignments))
	valueArgs := make([]interface{}, 0, len(assignments)*3)

	for i, assignment := range assignments {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, classID, assignment.PaperID, assignment.TeacherID)
	}

	query := fmt.Sprintf("INSERT INTO class_papers (class_id, paper_id, teacher_id) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err = db.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert paper assignments: %v", err)
	}
	return nil
}

func GetSubjectPapersForClass(db *sql.DB, classID, subjectID string) ([]*models.Paper, error) {
	return []*models.Paper{}, nil
}

func GetPapersBySubject(db *sql.DB, subjectID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.subject_id = $1 AND p.deleted_at IS NULL
			  ORDER BY p.name, p.code`

	rows, err := db.Query(query, subjectID)
	if err != nil {
		return []*models.Paper{}, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		var subjectName, subjectCode *string
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory, &paper.IsActive,
			&paper.CreatedAt, &paper.UpdatedAt, &subjectName, &subjectCode,
		)
		if err != nil {
			continue
		}

		// Set subject if exists
		if subjectName != nil {
			paper.Subject = &models.Subject{
				ID:   paper.SubjectID,
				Name: *subjectName,
				Code: *subjectCode,
			}
		}

		papers = append(papers, paper)
	}

	return papers, nil
}

func GetPaperByID(db *sql.DB, id string) (*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at,
			  s.name as subject_name, s.code as subject_code
			  FROM papers p
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  WHERE p.id = $1 AND p.deleted_at IS NULL`

	paper := &models.Paper{}
	var subjectName, subjectCode *string
	err := db.QueryRow(query, id).Scan(
		&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory, &paper.IsActive,
		&paper.CreatedAt, &paper.UpdatedAt, &subjectName, &subjectCode,
	)
	if err != nil {
		return nil, err
	}

	// Set subject if exists
	if subjectName != nil {
		paper.Subject = &models.Subject{
			ID:   paper.SubjectID,
			Name: *subjectName,
			Code: *subjectCode,
		}
	}

	return paper, nil
}

func GetSubjectByID(db *sql.DB, id string) (*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name,
			  COALESCE(p.paper_count, 0) as paper_count,
			  COALESCE(cs.class_count, 0) as class_count
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as paper_count 
				  FROM papers 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) p ON s.id = p.subject_id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as class_count 
				  FROM class_subjects 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) cs ON s.id = cs.subject_id
			  WHERE s.id = $1 AND s.is_active = true`

	subject := &models.Subject{}
	var departmentName *string
	var paperCount, classCount int
	err := db.QueryRow(query, id).Scan(
		&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
		&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName, &paperCount, &classCount,
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

	// Create dummy slices for template compatibility
	if paperCount > 0 {
		subject.Papers = make([]*models.Paper, paperCount)
	}
	if classCount > 0 {
		subject.Classes = make([]*models.Class, classCount)
	}

	return subject, nil
}

func RemoveSubjectFromClass(db *sql.DB, classID, subjectID string) error {
	query := `DELETE FROM class_subjects WHERE class_id = $1 AND subject_id = $2`
	_, err := db.Exec(query, classID, subjectID)
	if err != nil {
		return fmt.Errorf("failed to remove subject from class: %v", err)
	}
	return nil
}

func CreatePaper(db *sql.DB, paper *models.Paper) error {
	query := `INSERT INTO papers (subject_id, name, code, is_compulsory, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	paper.IsActive = true
	paper.IsCompulsory = true // Default to compulsory
	err := db.QueryRow(query, paper.SubjectID, paper.Name, paper.Code, paper.IsCompulsory, paper.IsActive).Scan(
		&paper.ID, &paper.CreatedAt, &paper.UpdatedAt,
	)
	return err
}

func UpdatePaper(db *sql.DB, paper *models.Paper) error {
	query := `UPDATE papers
			  SET subject_id = $1, name = $2, code = $3, is_compulsory = $4, is_active = $5, updated_at = NOW()
			  WHERE id = $6 AND deleted_at IS NULL`

	_, err := db.Exec(query, paper.SubjectID, paper.Name, paper.Code, paper.IsCompulsory, paper.IsActive, paper.ID)
	return err
}

func DeletePaper(db *sql.DB, id string) error {
	query := `UPDATE papers SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

func CreateDepartment(db *sql.DB, dept *models.Department) error {
	return nil
}

func UpdateDepartment(db *sql.DB, dept *models.Department) error {
	return nil
}

func DeleteDepartment(db *sql.DB, id string) error {
	return nil
}

func GetStudentsWithDetails(db *sql.DB) ([]*models.Student, error) {
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth, s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name, c.code as class_code,
			  p.id as parent_id, p.first_name as parent_first_name, p.last_name as parent_last_name, p.phone as parent_phone, p.email as parent_email
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  LEFT JOIN student_parents sp ON s.id = sp.student_id
			  LEFT JOIN parents p ON sp.parent_id = p.id
			  WHERE s.is_active = true
			  ORDER BY s.first_name, s.last_name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentMap := make(map[string]*models.Student)
	for rows.Next() {
		var studentID, studentIDCode, firstName, lastName string
		var dateOfBirth *time.Time
		var gender *string
		var address, classID *string
		var isActive bool
		var createdAt, updatedAt time.Time
		var className, classCode *string
		var parentID, parentFirstName, parentLastName, parentPhone, parentEmail *string

		err := rows.Scan(
			&studentID, &studentIDCode, &firstName, &lastName, &dateOfBirth, &gender, &address, &classID, &isActive, &createdAt, &updatedAt,
			&className, &classCode, &parentID, &parentFirstName, &parentLastName, &parentPhone, &parentEmail,
		)
		if err != nil {
			continue
		}

		// Get or create student
		student, exists := studentMap[studentID]
		if !exists {
			student = &models.Student{
				ID:          studentID,
				StudentID:   studentIDCode,
				FirstName:   firstName,
				LastName:    lastName,
				DateOfBirth: dateOfBirth,
				Address:     address,
				ClassID:     classID,
				IsActive:    isActive,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}

			if gender != nil {
				g := models.Gender(*gender)
				student.Gender = &g
			}

			// Set class if available
			if className != nil && classID != nil {
				student.Class = &models.Class{
					ID:   *classID,
					Name: *className,
					Code: classCode,
				}
			}

			studentMap[studentID] = student
		}

		// Add parent if available
		if parentID != nil && parentFirstName != nil && parentLastName != nil {
			parent := &models.Parent{
				ID:        *parentID,
				FirstName: *parentFirstName,
				LastName:  *parentLastName,
				Phone:     parentPhone,
				Email:     parentEmail,
			}
			student.Parents = append(student.Parents, parent)
		}
	}

	// Convert map to slice
	var students []*models.Student
	for _, student := range studentMap {
		students = append(students, student)
	}

	if students == nil {
		students = []*models.Student{}
	}

	return students, nil
}

func GetStudentsStats(db *sql.DB) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total students
	var totalStudents int
	err := db.QueryRow("SELECT COUNT(*) FROM students WHERE is_active = true").Scan(&totalStudents)
	if err != nil {
		totalStudents = 0
	}
	stats["total_students"] = totalStudents

	// Active students (same as total for now)
	stats["active_students"] = totalStudents

	// New students this month
	var newThisMonth int
	err = db.QueryRow("SELECT COUNT(*) FROM students WHERE is_active = true AND created_at >= date_trunc('month', CURRENT_DATE)").Scan(&newThisMonth)
	if err != nil {
		newThisMonth = 0
	}
	stats["new_this_month"] = newThisMonth

	// Pending applications (placeholder - you can implement based on your needs)
	stats["pending_applications"] = 0

	return stats, nil
}

func GetStudentsWithFiltersAndPagination(db *sql.DB, filters StudentFilters) ([]*models.Student, int, error) {
	// First get the total count
	totalCount, err := getStudentsCountWithFilters(db, filters)
	if err != nil {
		return nil, 0, err
	}

	// Then get the paginated results
	students, err := getStudentsWithFiltersInternal(db, filters, true)
	if err != nil {
		return nil, totalCount, err
	}

	return students, totalCount, nil
}

func GetStudentsWithFilters(db *sql.DB, filters StudentFilters) ([]*models.Student, error) {
	return getStudentsWithFiltersInternal(db, filters, false)
}

func getStudentsCountWithFilters(db *sql.DB, filters StudentFilters) (int, error) {
	// Base count query - simplified to avoid duplicates
	baseQuery := `SELECT COUNT(s.id)
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  WHERE s.is_active = true`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Apply the same filters as the main query
	if filters.Search != "" && len(filters.Search) >= 3 {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			LOWER(s.first_name) LIKE $%d
			OR LOWER(s.last_name) LIKE $%d
			OR LOWER(CONCAT(s.first_name, ' ', s.last_name)) LIKE $%d
			OR LOWER(s.student_id) LIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if filters.Status != "" {
		if filters.Status == "active" {
			conditions = append(conditions, "s.is_active = true")
		} else if filters.Status == "inactive" {
			conditions = append(conditions, "s.is_active = false")
		}
	}

	if filters.ClassID != "" {
		conditions = append(conditions, fmt.Sprintf("s.class_id = $%d", argIndex))
		args = append(args, filters.ClassID)
		argIndex++
	} else if filters.ClassIDs != "" {
		// Handle multiple class IDs
		classIDList := strings.Split(filters.ClassIDs, ",")
		if len(classIDList) > 0 {
			placeholders := make([]string, len(classIDList))
			for i, classID := range classIDList {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, strings.TrimSpace(classID))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("s.class_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if filters.Gender != "" {
		conditions = append(conditions, fmt.Sprintf("s.gender = $%d", argIndex))
		args = append(args, filters.Gender)
		argIndex++
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at <= $%d", argIndex))
		args = append(args, filters.DateTo)
		argIndex++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := db.QueryRow(baseQuery, args...).Scan(&count)
	return count, err
}

func getStudentsWithFiltersInternal(db *sql.DB, filters StudentFilters, withPagination bool) ([]*models.Student, error) {
	// Base query - simplified to avoid duplicates
	baseQuery := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth, s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name, c.code as class_code
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  WHERE s.is_active = true`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Search filter - simplified without parent search
	if filters.Search != "" && len(filters.Search) >= 3 {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			LOWER(s.first_name) LIKE $%d
			OR LOWER(s.last_name) LIKE $%d
			OR LOWER(CONCAT(s.first_name, ' ', s.last_name)) LIKE $%d
			OR LOWER(s.student_id) LIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Status filter
	if filters.Status != "" {
		if filters.Status == "active" {
			conditions = append(conditions, "s.is_active = true")
		} else if filters.Status == "inactive" {
			conditions = append(conditions, "s.is_active = false")
		}
	}

	// Class filter (single or multiple)
	if filters.ClassID != "" {
		conditions = append(conditions, fmt.Sprintf("s.class_id = $%d", argIndex))
		args = append(args, filters.ClassID)
		argIndex++
	} else if filters.ClassIDs != "" {
		// Handle multiple class IDs
		classIDList := strings.Split(filters.ClassIDs, ",")
		if len(classIDList) > 0 {
			placeholders := make([]string, len(classIDList))
			for i, classID := range classIDList {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, strings.TrimSpace(classID))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("s.class_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	// Gender filter
	if filters.Gender != "" {
		conditions = append(conditions, fmt.Sprintf("s.gender = $%d", argIndex))
		args = append(args, filters.Gender)
		argIndex++
	}

	// Date filters
	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at >= $%d", argIndex))
		args = append(args, filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("s.created_at <= $%d", argIndex))
		args = append(args, filters.DateTo)
		argIndex++
	}

	// Add conditions to query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add sorting - default to student_id for proper STU-2025-___ ordering
	sortBy := "s.student_id"
	if filters.SortBy == "name" {
		sortBy = "s.first_name"
	} else if filters.SortBy == "class" {
		sortBy = "c.name"
	}

	sortOrder := "ASC"
	if filters.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s, s.first_name ASC", sortBy, sortOrder)

	// Add pagination if requested
	if withPagination && filters.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.Limit, filters.Offset)
	}

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var studentID, studentIDCode, firstName, lastName string
		var dateOfBirth *time.Time
		var gender *string
		var address, classID *string
		var isActive bool
		var createdAt, updatedAt time.Time
		var className, classCode *string

		err := rows.Scan(
			&studentID, &studentIDCode, &firstName, &lastName, &dateOfBirth, &gender, &address, &classID, &isActive, &createdAt, &updatedAt,
			&className, &classCode,
		)
		if err != nil {
			continue
		}

		student := &models.Student{
			ID:          studentID,
			StudentID:   studentIDCode,
			FirstName:   firstName,
			LastName:    lastName,
			DateOfBirth: dateOfBirth,
			Address:     address,
			ClassID:     classID,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		if gender != nil {
			g := models.Gender(*gender)
			student.Gender = &g
		}

		// Set class if available
		if className != nil && classID != nil {
			student.Class = &models.Class{
				ID:   *classID,
				Name: *className,
				Code: classCode,
			}
		}

		students = append(students, student)
	}

	if students == nil {
		students = []*models.Student{}
	}

	return students, nil
}

func GetStudentByID(db *sql.DB, id string) (*models.Student, error) {
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth, s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  WHERE s.id = $1`

	var studentID, studentIDCode, firstName, lastName string
	var dateOfBirth *time.Time
	var gender *string
	var address, classID *string
	var isActive bool
	var createdAt, updatedAt time.Time
	var className *string

	err := db.QueryRow(query, id).Scan(
		&studentID, &studentIDCode, &firstName, &lastName, &dateOfBirth, &gender, &address, &classID, &isActive, &createdAt, &updatedAt,
		&className,
	)
	if err != nil {
		return nil, err
	}

	student := &models.Student{
		ID:          studentID,
		StudentID:   studentIDCode,
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: dateOfBirth,
		Address:     address,
		ClassID:     classID,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if gender != nil {
		g := models.Gender(*gender)
		student.Gender = &g
	}

	if className != nil && classID != nil {
		student.Class = &models.Class{
			ID:   *classID,
			Name: *className,
		}
	}

	// Load parents
	parentRows, err := db.Query(`SELECT p.id, p.first_name, p.last_name, p.phone, p.email 
							 FROM parents p 
							 JOIN student_parents sp ON p.id = sp.parent_id 
							 WHERE sp.student_id = $1`, studentID)
	if err == nil {
		defer parentRows.Close()
		for parentRows.Next() {
			var parentID, parentFirstName, parentLastName string
			var parentPhone, parentEmail *string
			if parentRows.Scan(&parentID, &parentFirstName, &parentLastName, &parentPhone, &parentEmail) == nil {
				parent := &models.Parent{
					ID:        parentID,
					FirstName: parentFirstName,
					LastName:  parentLastName,
					Phone:     parentPhone,
					Email:     parentEmail,
				}
				student.Parents = append(student.Parents, parent)
			}
		}
	}

	return student, nil
}

func GetStudentParentRelationship(db *sql.DB, studentID, relationshipType string) ([]*models.Parent, error) {
	return []*models.Parent{}, nil
}

func GetStudentsByYear(db *sql.DB, year int) ([]*models.Student, error) {
	return []*models.Student{}, nil
}

func CreateStudent(db *sql.DB, student *models.Student) error {
	return nil
}

func LinkStudentToParent(db *sql.DB, studentID, parentID, relationshipType string) error {
	return nil
}

func SearchSubjects(db *sql.DB, query string) ([]*models.Subject, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"
	sqlQuery := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
				  d.name as department_name,
				  COALESCE(p.paper_count, 0) as paper_count,
				  COALESCE(cs.class_count, 0) as class_count
				  FROM subjects s
				  LEFT JOIN departments d ON s.department_id = d.id
				  LEFT JOIN (
					  SELECT subject_id, COUNT(*) as paper_count 
					  FROM papers 
					  WHERE deleted_at IS NULL 
					  GROUP BY subject_id
				  ) p ON s.id = p.subject_id
				  LEFT JOIN (
					  SELECT subject_id, COUNT(*) as class_count 
					  FROM class_subjects 
					  WHERE deleted_at IS NULL 
					  GROUP BY subject_id
				  ) cs ON s.id = cs.subject_id
				  WHERE s.is_active = true
				  AND (LOWER(s.name) LIKE $1 OR LOWER(s.code) LIKE $1)
				  ORDER BY s.name`

	rows, err := db.Query(sqlQuery, searchPattern)
	if err != nil {
		return []*models.Subject{}, err
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		var departmentName *string
		var paperCount, classCount int
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName, &paperCount, &classCount,
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

		// Create dummy slices for template compatibility
		if paperCount > 0 {
			subject.Papers = make([]*models.Paper, paperCount)
		}
		if classCount > 0 {
			subject.Classes = make([]*models.Class, classCount)
		}

		subjects = append(subjects, subject)
	}
	return subjects, nil
}

func CreateSubject(db *sql.DB, subject *models.Subject) error {
	query := `INSERT INTO subjects (name, code, department_id, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, true, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	return db.QueryRow(query, subject.Name, subject.Code, subject.DepartmentID).Scan(
		&subject.ID, &subject.CreatedAt, &subject.UpdatedAt,
	)
}

func UpdateSubject(db *sql.DB, subject *models.Subject) error {
	query := `UPDATE subjects
			  SET name = $1, code = $2, department_id = $3, updated_at = NOW()
			  WHERE id = $4 AND is_active = true`

	_, err := db.Exec(query, subject.Name, subject.Code, subject.DepartmentID, subject.ID)
	return err
}

func DeleteSubject(db *sql.DB, id string) error {
	query := `UPDATE subjects SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

func UpdateStudent(db *sql.DB, student *models.Student) error {
	return nil
}

func ChangeStudentParent(db *sql.DB, studentID, parentID, relationshipType string) error {
	return nil
}

func DeleteStudent(db *sql.DB, id string) error {
	return nil
}

func GetParentsForSelection(db *sql.DB, search string) ([]*models.Parent, error) {
	return []*models.Parent{}, nil
}

func SearchStudents(db *sql.DB, query string) ([]*models.Student, error) {
	if len(query) < 3 {
		return []*models.Student{}, nil
	}

	searchPattern := "%" + strings.ToLower(query) + "%"

	sqlQuery := `SELECT DISTINCT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth, s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at,
			  c.name as class_name, c.code as class_code
			  FROM students s
			  LEFT JOIN classes c ON s.class_id = c.id
			  LEFT JOIN student_parents sp ON s.id = sp.student_id
			  LEFT JOIN parents p ON sp.parent_id = p.id
			  WHERE s.is_active = true
			  AND (
				  LOWER(s.first_name) LIKE $1
				  OR LOWER(s.last_name) LIKE $1
				  OR LOWER(CONCAT(s.first_name, ' ', s.last_name)) LIKE $1
				  OR LOWER(s.student_id) LIKE $1
				  OR LOWER(p.first_name) LIKE $1
				  OR LOWER(p.last_name) LIKE $1
				  OR LOWER(CONCAT(p.first_name, ' ', p.last_name)) LIKE $1
			  )
			  ORDER BY s.first_name, s.last_name
			  LIMIT 50`

	rows, err := db.Query(sqlQuery, searchPattern)
	if err != nil {
		return []*models.Student{}, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var studentID, studentIDCode, firstName, lastName string
		var dateOfBirth *time.Time
		var gender *string
		var address, classID *string
		var isActive bool
		var createdAt, updatedAt time.Time
		var className, classCode *string

		err := rows.Scan(
			&studentID, &studentIDCode, &firstName, &lastName, &dateOfBirth, &gender, &address, &classID, &isActive, &createdAt, &updatedAt,
			&className, &classCode,
		)
		if err != nil {
			continue
		}

		student := &models.Student{
			ID:          studentID,
			StudentID:   studentIDCode,
			FirstName:   firstName,
			LastName:    lastName,
			DateOfBirth: dateOfBirth,
			Address:     address,
			ClassID:     classID,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		if gender != nil {
			g := models.Gender(*gender)
			student.Gender = &g
		}

		if className != nil && classID != nil {
			student.Class = &models.Class{
				ID:   *classID,
				Name: *className,
				Code: classCode,
			}
		}

		students = append(students, student)
	}

	if students == nil {
		students = []*models.Student{}
	}

	return students, nil
}

func GetAllStudents(db *sql.DB) ([]*models.Student, error) {
	return []*models.Student{}, nil
}

// assignClassTeacherRoleInDB assigns the class_teacher role to a user
func assignClassTeacherRoleInDB(tx *sql.Tx, userID string) error {
	// Get class_teacher role ID
	var roleID string
	err := tx.QueryRow("SELECT id FROM roles WHERE name = 'class_teacher' LIMIT 1").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("class_teacher role not found: %v", err)
	}

	// Check if user already has this role
	var existingID string
	err = tx.QueryRow("SELECT id FROM user_roles WHERE user_id = $1 AND role_id = $2 LIMIT 1", userID, roleID).Scan(&existingID)
	if err == nil {
		// Role already exists
		return nil
	}

	// Add the role
	_, err = tx.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())", userID, roleID)
	return err
}

// removeClassTeacherRoleInDB removes the class_teacher role from a user
func removeClassTeacherRoleInDB(tx *sql.Tx, userID string) error {
	// Get class_teacher role ID
	var roleID string
	err := tx.QueryRow("SELECT id FROM roles WHERE name = 'class_teacher' LIMIT 1").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("class_teacher role not found: %v", err)
	}

	// Remove the role
	_, err = tx.Exec("DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID)
	return err
}

// GetTeacherAvailability retrieves the availability for a specific teacher
func GetTeacherAvailability(db *sql.DB, teacherID string) ([]*models.TeacherAvailability, error) {
	query := `SELECT id, teacher_id, day_of_week, is_available, 
			  to_char(start_time, 'HH24:MI') as start_time, 
			  to_char(end_time, 'HH24:MI') as end_time, 
			  created_at, updated_at
			  FROM teacher_availability
			  WHERE teacher_id = $1
			  ORDER BY day_of_week`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var availability []*models.TeacherAvailability
	for rows.Next() {
		item := &models.TeacherAvailability{}
		err := rows.Scan(
			&item.ID, &item.TeacherID, &item.DayOfWeek, &item.IsAvailable,
			&item.StartTime, &item.EndTime, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		availability = append(availability, item)
	}

	return availability, nil
}

// UpdateTeacherAvailability updates or inserts the availability for a teacher for multiple days
func UpdateTeacherAvailability(db *sql.DB, teacherID string, availability []*models.TeacherAvailability) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO teacher_availability (teacher_id, day_of_week, is_available, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (teacher_id, day_of_week) DO UPDATE SET
			is_available = EXCLUDED.is_available,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			updated_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range availability {
		// Handle nullable time values
		var startTime, endTime interface{}
		if item.StartTime.Valid {
			startTime = item.StartTime.String
		} else {
			startTime = nil
		}
		if item.EndTime.Valid {
			endTime = item.EndTime.String
		} else {
			endTime = nil
		}

		_, err := stmt.Exec(teacherID, item.DayOfWeek, item.IsAvailable, startTime, endTime)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAttendanceByTimetableEntryAndDate gets attendance records for a timetable entry and date
func GetAttendanceByTimetableEntryAndDate(db *sql.DB, timetableEntryID string, date time.Time) ([]*models.Attendance, error) {
	query := `SELECT id, student_id, class_id, timetable_entry_id, paper_id, term_id, date, status, marked_by, created_at, updated_at
			  FROM attendance
			  WHERE timetable_entry_id = $1 AND date = $2 AND deleted_at IS NULL`

	rows, err := db.Query(query, timetableEntryID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.Attendance
	for rows.Next() {
		record := &models.Attendance{}
		var status string
		err := rows.Scan(
			&record.ID, &record.StudentID, &record.ClassID, &record.TimetableEntryID,
			&record.PaperID, &record.TermID, &record.Date, &status, &record.MarkedBy,
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			continue
		}
		record.Status = models.AttendanceStatus(status)
		records = append(records, record)
	}

	return records, nil
}

// GetStudentsByTimetableEntry gets students for a timetable entry (based on class)
func GetStudentsByTimetableEntry(db *sql.DB, timetableEntryID string) ([]*models.Student, error) {
	query := `SELECT s.id, s.student_id, s.first_name, s.last_name, s.date_of_birth, s.gender, s.address, s.class_id, s.is_active, s.created_at, s.updated_at
			  FROM students s
			  INNER JOIN timetable_entries te ON s.class_id = te.class_id
			  WHERE te.id = $1 AND s.is_active = true
			  ORDER BY s.first_name, s.last_name`

	rows, err := db.Query(query, timetableEntryID)
	if err != nil {
		return []*models.Student{}, err
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
			continue
		}
		students = append(students, student)
	}

	return students, nil
}

// GetTimetableEntryByID gets a single timetable entry with full details by ID
func GetTimetableEntryByID(db *sql.DB, id string) (*models.TimetableEntryResponse, error) {
	query := `SELECT te.id, te.class_id, te.subject_id, te.teacher_id, te.day_of_week, 
			  CONCAT(to_char(te.start_time, 'HH24:MI'), ' - ', to_char(te.end_time, 'HH24:MI')) as time_slot,
			  te.created_at, te.updated_at, te.paper_id,
			  s.name as subject_name, c.name as class_name, u.first_name, u.last_name,
			  (SELECT COUNT(*) FROM students WHERE class_id = te.class_id AND is_active = true) as student_count,
			  COALESCE(p.code, '') as paper_code
			  FROM timetable_entries te
			  LEFT JOIN subjects s ON te.subject_id = s.id
			  LEFT JOIN classes c ON te.class_id = c.id
			  LEFT JOIN users u ON te.teacher_id = u.id
			  LEFT JOIN papers p ON te.paper_id = p.id
			  WHERE te.id = $1 AND te.is_active = true`

	entry := &models.TimetableEntryResponse{}
	var subjectName, className, teacherFirstName, teacherLastName *string
	err := db.QueryRow(query, id).Scan(
		&entry.ID, &entry.ClassID, &entry.SubjectID, &entry.TeacherID,
		&entry.Day, &entry.TimeSlot, &entry.CreatedAt, &entry.UpdatedAt, &entry.PaperID,
		&subjectName, &className, &teacherFirstName, &teacherLastName,
		&entry.StudentCount, &entry.PaperCode,
	)
	if err != nil {
		return nil, err
	}

	if subjectName != nil {
		entry.SubjectName = *subjectName
	}
	if className != nil {
		entry.ClassName = *className
	}
	if teacherFirstName != nil && teacherLastName != nil {
		entry.TeacherName = *teacherFirstName + " " + *teacherLastName
	}

	return entry, nil
}

// GetTimetableEntriesByTeacherAndDay gets timetable entries for a teacher on a specific day
func GetTimetableEntriesByTeacherAndDay(db *sql.DB, teacherID, dayOfWeek string) ([]*models.TimetableEntryResponse, error) {
	query := `SELECT te.id, te.class_id, te.subject_id, te.teacher_id, te.day_of_week, 
			  CONCAT(to_char(te.start_time, 'HH24:MI'), ' - ', to_char(te.end_time, 'HH24:MI')) as time_slot,
			  te.created_at, te.updated_at, te.paper_id,
			  s.name as subject_name, c.name as class_name, u.first_name, u.last_name,
			  (SELECT COUNT(*) FROM students WHERE class_id = te.class_id AND is_active = true) as student_count,
			  COALESCE(p.code, '') as paper_code
			  FROM timetable_entries te
			  LEFT JOIN subjects s ON te.subject_id = s.id
			  LEFT JOIN classes c ON te.class_id = c.id
			  LEFT JOIN users u ON te.teacher_id = u.id
			  LEFT JOIN papers p ON te.paper_id = p.id
			  WHERE te.teacher_id = $1 AND LOWER(te.day_of_week) = LOWER($2) AND te.is_active = true
			  ORDER BY te.start_time`

	rows, err := db.Query(query, teacherID, dayOfWeek)
	if err != nil {
		return []*models.TimetableEntryResponse{}, err
	}
	defer rows.Close()

	var entries []*models.TimetableEntryResponse
	for rows.Next() {
		entry := &models.TimetableEntryResponse{}
		var subjectName, className, teacherFirstName, teacherLastName *string
		err := rows.Scan(
			&entry.ID, &entry.ClassID, &entry.SubjectID, &entry.TeacherID,
			&entry.Day, &entry.TimeSlot, &entry.CreatedAt, &entry.UpdatedAt, &entry.PaperID,
			&subjectName, &className, &teacherFirstName, &teacherLastName,
			&entry.StudentCount, &entry.PaperCode,
		)
		if err != nil {
			continue
		}

		// Add subject and class names for display
		if subjectName != nil {
			entry.SubjectName = *subjectName
		}
		if className != nil {
			entry.ClassName = *className
		}

		// Add teacher name
		if teacherFirstName != nil && teacherLastName != nil {
			entry.TeacherName = *teacherFirstName + " " + *teacherLastName
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAllTimetableEntriesByDay gets all timetable entries for a specific day (admin/head teacher only)
func GetAllTimetableEntriesByDay(db *sql.DB, dayOfWeek string) ([]*models.TimetableEntryResponse, error) {
	query := `SELECT te.id, te.class_id, te.subject_id, te.teacher_id, te.day_of_week, 
			  CONCAT(to_char(te.start_time, 'HH24:MI'), ' - ', to_char(te.end_time, 'HH24:MI')) as time_slot,
			  te.created_at, te.updated_at,
			  s.name as subject_name, c.name as class_name, u.first_name, u.last_name,
			  (SELECT COUNT(*) FROM students WHERE class_id = te.class_id AND is_active = true) as student_count
			  FROM timetable_entries te
			  LEFT JOIN subjects s ON te.subject_id = s.id
			  LEFT JOIN classes c ON te.class_id = c.id
			  LEFT JOIN users u ON te.teacher_id = u.id
			  WHERE LOWER(te.day_of_week) = LOWER($1) AND te.is_active = true
			  ORDER BY te.start_time, c.name`

	rows, err := db.Query(query, dayOfWeek)
	if err != nil {
		return []*models.TimetableEntryResponse{}, err
	}
	defer rows.Close()

	var entries []*models.TimetableEntryResponse
	for rows.Next() {
		entry := &models.TimetableEntryResponse{}
		var subjectName, className, teacherFirstName, teacherLastName *string
		err := rows.Scan(
			&entry.ID, &entry.ClassID, &entry.SubjectID, &entry.TeacherID,
			&entry.Day, &entry.TimeSlot, &entry.CreatedAt, &entry.UpdatedAt,
			&subjectName, &className, &teacherFirstName, &teacherLastName,
			&entry.StudentCount,
		)
		if err != nil {
			continue
		}

		// Add subject and class names for display
		if subjectName != nil {
			entry.SubjectName = *subjectName
		}
		if className != nil {
			entry.ClassName = *className
		}

		// Add teacher name
		if teacherFirstName != nil && teacherLastName != nil {
			entry.TeacherName = *teacherFirstName + " " + *teacherLastName
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Exam functions
func GetAllExams(db *sql.DB, classID string) ([]*models.Exam, error) {
	query := `SELECT e.id, e.name, e.class_id, e.academic_year_id, e.term_id, e.paper_id, e.assessment_type_id, e.type, e.start_time, e.end_time, e.is_active, e.created_at, e.updated_at,
			  c.name as class_name, ay.name as academic_year_name, t.name as term_name, p.name as paper_name, p.subject_id, s.name as subject_name,
			  at.name as assessment_type_name, ac.display_style as display_style
			  FROM exams e
			  LEFT JOIN classes c ON e.class_id = c.id
			  LEFT JOIN academic_years ay ON e.academic_year_id = ay.id
			  LEFT JOIN terms t ON e.term_id = t.id
			  LEFT JOIN papers p ON e.paper_id = p.id
			  LEFT JOIN subjects s ON p.subject_id = s.id
			  LEFT JOIN assessment_types at ON (e.assessment_type_id = at.id OR (e.assessment_type_id IS NULL AND LOWER(e.type) = LOWER(at.code)))
			  LEFT JOIN assessment_categories ac ON at.category_id = ac.id
			  WHERE e.deleted_at IS NULL`

	var args []interface{}
	if classID != "" {
		query += " AND e.class_id = $1"
		args = append(args, classID)
	}
	query += " ORDER BY e.start_time DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return []*models.Exam{}, err
	}
	defer rows.Close()

	var exams []*models.Exam
	for rows.Next() {
		exam := &models.Exam{}
		var className, academicYearName, termName, paperName, subjectName *string
		var assessmentTypeName, displayStyle *string
		var subjectID *string
		err := rows.Scan(
			&exam.ID, &exam.Name, &exam.ClassID, &exam.AcademicYearID, &exam.TermID, &exam.PaperID, &exam.AssessmentTypeID, &exam.Type,
			&exam.StartTime, &exam.EndTime, &exam.IsActive, &exam.CreatedAt, &exam.UpdatedAt,
			&className, &academicYearName, &termName, &paperName, &subjectID, &subjectName,
			&assessmentTypeName, &displayStyle,
		)
		if err != nil {
			continue
		}

		if className != nil {
			exam.Class = &models.Class{ID: exam.ClassID, Name: *className}
		}
		if academicYearName != nil && exam.AcademicYearID != nil {
			exam.AcademicYear = &models.AcademicYear{ID: *exam.AcademicYearID, Name: *academicYearName}
		}
		if termName != nil && exam.TermID != nil {
			exam.Term = &models.Term{ID: *exam.TermID, Name: *termName}
		}
		if paperName != nil {
			exam.Paper = &models.Paper{ID: exam.PaperID, Name: *paperName}
			if subjectID != nil && subjectName != nil {
				exam.Paper.SubjectID = *subjectID
				exam.Paper.Subject = &models.Subject{ID: *subjectID, Name: *subjectName}
			}
		}
		if assessmentTypeName != nil {
			if exam.AssessmentType == nil {
				exam.AssessmentType = &models.AssessmentType{}
			}
			exam.AssessmentType.Name = *assessmentTypeName
			if displayStyle != nil {
				if exam.AssessmentType.Category == nil {
					exam.AssessmentType.Category = &models.AssessmentCategory{}
				}
				exam.AssessmentType.Category.DisplayStyle = *displayStyle
			}
		}

		exams = append(exams, exam)
	}

	return exams, nil
}

func GetExamByID(db *sql.DB, id string) (*models.Exam, error) {
	query := `SELECT id, name, class_id, academic_year_id, term_id, paper_id, assessment_type_id, type, start_time, end_time, is_active, created_at, updated_at
			  FROM exams WHERE id = $1 AND deleted_at IS NULL`

	exam := &models.Exam{}
	err := db.QueryRow(query, id).Scan(
		&exam.ID, &exam.Name, &exam.ClassID, &exam.AcademicYearID, &exam.TermID, &exam.PaperID, &exam.AssessmentTypeID, &exam.Type,
		&exam.StartTime, &exam.EndTime, &exam.IsActive, &exam.CreatedAt, &exam.UpdatedAt,
	)
	return exam, err
}

func CreateExam(db *sql.DB, exam *models.Exam) error {
	query := `INSERT INTO exams (name, class_id, academic_year_id, term_id, paper_id, assessment_type_id, type, start_time, end_time, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	exam.IsActive = true
	if exam.Type == "" {
		exam.Type = "exam"
	}
	err := db.QueryRow(query, exam.Name, exam.ClassID, exam.AcademicYearID, exam.TermID, exam.PaperID, exam.AssessmentTypeID, exam.Type,
		exam.StartTime, exam.EndTime, exam.IsActive).Scan(&exam.ID, &exam.CreatedAt, &exam.UpdatedAt)
	return err
}

func UpdateExam(db *sql.DB, exam *models.Exam) error {
	query := `UPDATE exams SET name = $1, class_id = $2, academic_year_id = $3, term_id = $4, paper_id = $5,
			  assessment_type_id = $6, type = $7, start_time = $8, end_time = $9, is_active = $10, updated_at = NOW()
			  WHERE id = $11 AND deleted_at IS NULL`

	_, err := db.Exec(query, exam.Name, exam.ClassID, exam.AcademicYearID, exam.TermID, exam.PaperID,
		exam.AssessmentTypeID, exam.Type, exam.StartTime, exam.EndTime, exam.IsActive, exam.ID)
	return err
}

func DeleteExam(db *sql.DB, id string) error {
	query := `UPDATE exams SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// GetSubjectsByClass retrieves all subjects associated with a specific class
func GetSubjectsByClass(db *sql.DB, classID string) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at
			  FROM subjects s
			  JOIN class_subjects cs ON s.id = cs.subject_id
			  WHERE cs.class_id = $1 AND s.deleted_at IS NULL AND cs.deleted_at IS NULL
			  ORDER BY s.name`

	rows, err := db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []*models.Subject
	for rows.Next() {
		subject := &models.Subject{}
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt,
		)
		if err != nil {
			continue
		}
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

// GetPapersByClass retrieves all papers for subjects associated with a specific class
func GetPapersByClass(db *sql.DB, classID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at
			  FROM papers p
			  JOIN class_subjects cs ON p.subject_id = cs.subject_id
			  WHERE cs.class_id = $1 AND p.deleted_at IS NULL AND cs.deleted_at IS NULL
			  ORDER BY p.name`

	rows, err := db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
		)
		if err != nil {
			continue
		}
		papers = append(papers, paper)
	}
	return papers, nil
}

// GetPapersByClassAndSubject retrieves papers for a specific subject restricted by class
func GetPapersByClassAndSubject(db *sql.DB, classID, subjectID string) ([]*models.Paper, error) {
	query := `SELECT p.id, p.subject_id, p.name, p.code, p.is_compulsory, p.is_active, p.created_at, p.updated_at
			  FROM papers p
			  JOIN class_subjects cs ON p.subject_id = cs.subject_id
			  WHERE cs.class_id = $1 AND p.subject_id = $2 AND p.deleted_at IS NULL AND cs.deleted_at IS NULL
			  ORDER BY p.name`

	rows, err := db.Query(query, classID, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var papers []*models.Paper
	for rows.Next() {
		paper := &models.Paper{}
		err := rows.Scan(
			&paper.ID, &paper.SubjectID, &paper.Name, &paper.Code, &paper.IsCompulsory,
			&paper.IsActive, &paper.CreatedAt, &paper.UpdatedAt,
		)
		if err != nil {
			continue
		}
		papers = append(papers, paper)
	}
	return papers, nil
}

func MarkLessonAsConducted(db *sql.DB, logRec *models.ConductedLesson) error {
	query := `INSERT INTO conducted_lessons (id, timetable_entry_id, term_id, date, teacher_id, topic, notes, created_at, updated_at)
			  VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
			  ON CONFLICT (timetable_entry_id, date) 
			  DO UPDATE SET topic = EXCLUDED.topic, notes = EXCLUDED.notes, updated_at = NOW()`
	_, err := db.Exec(query, logRec.TimetableEntryID, logRec.TermID, logRec.Date, logRec.TeacherID, logRec.Topic, logRec.Notes)
	return err
}

func GetConductedLesson(db *sql.DB, timetableEntryID string, date time.Time) (*models.ConductedLesson, error) {
	query := `SELECT id, timetable_entry_id, term_id, date, teacher_id, topic, notes, created_at, updated_at
			  FROM conducted_lessons
			  WHERE timetable_entry_id = $1 AND date = $2`

	l := &models.ConductedLesson{}
	err := db.QueryRow(query, timetableEntryID, date).Scan(
		&l.ID, &l.TimetableEntryID, &l.TermID, &l.Date, &l.TeacherID, &l.Topic, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return l, err
}

func GetStudentLessonAttendanceReport(db *sql.DB, studentID string) ([]map[string]interface{}, error) {
	query := `SELECT 
				cl.id as lesson_log_id,
				cl.date,
				te.day_of_week,
				CONCAT(to_char(te.start_time, 'HH24:MI'), ' - ', to_char(te.end_time, 'HH24:MI')) as time_slot,
				s.name as subject_name,
				COALESCE(p.code, '') as paper_code,
				COALESCE(cl.topic, '') as topic,
				COALESCE(a.status, 'pending') as status
			FROM conducted_lessons cl
			JOIN timetable_entries te ON cl.timetable_entry_id = te.id
			JOIN subjects s ON te.subject_id = s.id
			LEFT JOIN papers p ON te.paper_id = p.id
			LEFT JOIN attendance a ON a.timetable_entry_id = cl.timetable_entry_id AND a.date = cl.date AND a.student_id = $1
			WHERE te.class_id = (SELECT class_id FROM students WHERE id = $1)
			ORDER BY cl.date DESC, te.start_time DESC`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []map[string]interface{}
	for rows.Next() {
		var id, date, day, timeSlot, subject, paper, topic, status string
		err := rows.Scan(&id, &date, &day, &timeSlot, &subject, &paper, &topic, &status)
		if err != nil {
			continue
		}
		report = append(report, map[string]interface{}{
			"id":        id,
			"date":      date,
			"day":       day,
			"time_slot": timeSlot,
			"subject":   subject,
			"paper":     paper,
			"topic":     topic,
			"status":    status,
		})
	}
	return report, nil
}

func GetTimetableEntriesByClassAndDay(db *sql.DB, classID, dayOfWeek string) ([]*models.TimetableEntryResponse, error) {
	query := `SELECT te.id, te.class_id, te.subject_id, te.teacher_id, te.day_of_week, 
			  CONCAT(to_char(te.start_time, 'HH24:MI'), ' - ', to_char(te.end_time, 'HH24:MI')) as time_slot,
			  te.created_at, te.updated_at, te.paper_id,
			  s.name as subject_name, c.name as class_name, u.first_name, u.last_name,
			  (SELECT COUNT(*) FROM students WHERE class_id = te.class_id AND is_active = true) as student_count,
			  COALESCE(p.code, '') as paper_code
			  FROM timetable_entries te
			  LEFT JOIN subjects s ON te.subject_id = s.id
			  LEFT JOIN classes c ON te.class_id = c.id
			  LEFT JOIN users u ON te.teacher_id = u.id
			  LEFT JOIN papers p ON te.paper_id = p.id
			  WHERE te.class_id = $1 AND LOWER(te.day_of_week) = LOWER($2) AND te.is_active = true
			  ORDER BY te.start_time`

	rows, err := db.Query(query, classID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.TimetableEntryResponse
	for rows.Next() {
		entry := &models.TimetableEntryResponse{}
		var subjectName, className, teacherFirstName, teacherLastName *string
		err := rows.Scan(
			&entry.ID, &entry.ClassID, &entry.SubjectID, &entry.TeacherID,
			&entry.Day, &entry.TimeSlot, &entry.CreatedAt, &entry.UpdatedAt, &entry.PaperID,
			&subjectName, &className, &teacherFirstName, &teacherLastName,
			&entry.StudentCount, &entry.PaperCode,
		)
		if err != nil {
			continue
		}
		if subjectName != nil {
			entry.SubjectName = *subjectName
		}
		if className != nil {
			entry.ClassName = *className
		}
		if teacherFirstName != nil && teacherLastName != nil {
			entry.TeacherName = *teacherFirstName + " " + *teacherLastName
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func GetConductedLessonsByClassAndDate(db *sql.DB, classID string, date time.Time) ([]*models.ConductedLesson, error) {
	query := `SELECT cl.id, cl.timetable_entry_id, cl.term_id, cl.date, cl.teacher_id, cl.topic, cl.notes, cl.created_at, cl.updated_at
			  FROM conducted_lessons cl
			  JOIN timetable_entries te ON cl.timetable_entry_id = te.id
			  WHERE te.class_id = $1 AND cl.date = $2`

	rows, err := db.Query(query, classID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := []*models.ConductedLesson{}
	for rows.Next() {
		l := &models.ConductedLesson{}
		err := rows.Scan(
			&l.ID, &l.TimetableEntryID, &l.TermID, &l.Date, &l.TeacherID, &l.Topic, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			continue
		}
		lessons = append(lessons, l)
	}
	return lessons, nil
}

func GetClassTermAttendanceSummary(db *sql.DB, classID, termID string) ([]map[string]interface{}, error) {
	// Query to get aggregated stats for each student in the class for the term
	query := `
		SELECT 
			s.id as student_id,
			s.first_name,
			s.last_name,
			s.student_id as student_code,
			COUNT(DISTINCT cl.id) as total_conducted,
			COUNT(a.id) FILTER (WHERE a.status = 'present') as present_count,
			COUNT(a.id) FILTER (WHERE a.status = 'absent') as absent_count,
			COUNT(a.id) FILTER (WHERE a.status = 'late') as late_count
		FROM students s
		CROSS JOIN conducted_lessons cl
		JOIN timetable_entries te ON cl.timetable_entry_id = te.id
		LEFT JOIN attendance a ON a.student_id = s.id 
			AND a.timetable_entry_id = cl.timetable_entry_id 
			AND a.date = cl.date
		WHERE s.class_id = $1 
			AND te.class_id = $1
			AND cl.term_id = $2
			AND s.is_active = true
		GROUP BY s.id, s.first_name, s.last_name, s.student_id
		ORDER BY s.first_name, s.last_name
	`

	rows, err := db.Query(query, classID, termID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary []map[string]interface{}
	for rows.Next() {
		var sid, fname, lname, scode string
		var total, present, absent, late int
		err := rows.Scan(&sid, &fname, &lname, &scode, &total, &present, &absent, &late)
		if err != nil {
			continue
		}
		summary = append(summary, map[string]interface{}{
			"student_id":      sid,
			"first_name":      fname,
			"last_name":       lname,
			"student_code":    scode,
			"total_conducted": total,
			"present_count":   present,
			"absent_count":    absent,
			"late_count":      late,
		})
	}

	return summary, nil
}
