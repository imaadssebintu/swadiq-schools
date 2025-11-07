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

// CreateTeacher creates a new teacher with department assignment
func CreateTeacher(db *sql.DB, user *models.User, departmentID *string) error {
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
	query := `SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at,
			  STRING_AGG(DISTINCT r.name, ', ') as roles,
			  STRING_AGG(DISTINCT d.name, ', ') as department_names,
			  STRING_AGG(DISTINCT c.name, ', ') as class_names
			  FROM users u
			  INNER JOIN user_roles ur ON u.id = ur.user_id
			  INNER JOIN roles r ON ur.role_id = r.id
			  LEFT JOIN user_departments ud ON u.id = ud.user_id
			  LEFT JOIN departments d ON ud.department_id = d.id
			  LEFT JOIN classes c ON c.teacher_id = u.id AND c.is_active = true
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
		var departmentNames *string
		var classNames *string
		err := rows.Scan(
			&teacher.ID, &teacher.Email, &teacher.FirstName, &teacher.LastName,
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

		// Load departments for this teacher
		departments, _ := GetUserDepartments(db, teacher.ID)
		teacher.Departments = departments

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
			  FROM roles r
			  LEFT JOIN user_roles ur ON r.id = ur.role_id
			  LEFT JOIN users u ON ur.user_id = u.id AND u.is_active = true
			  WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
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

// GetAllSubjects gets all subjects with paper counts
func GetAllSubjects(db *sql.DB) ([]*models.Subject, error) {
	query := `SELECT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name,
			  COALESCE(p.paper_count, 0) as paper_count
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  LEFT JOIN (
				  SELECT subject_id, COUNT(*) as paper_count 
				  FROM papers 
				  WHERE deleted_at IS NULL 
				  GROUP BY subject_id
			  ) p ON s.id = p.subject_id
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
		var paperCount int
		err := rows.Scan(
			&subject.ID, &subject.Name, &subject.Code, &subject.DepartmentID,
			&subject.IsActive, &subject.CreatedAt, &subject.UpdatedAt, &departmentName, &paperCount,
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

		// Create dummy papers slice for template compatibility
		if paperCount > 0 {
			subject.Papers = make([]*models.Paper, paperCount)
		}

		subjects = append(subjects, subject)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}

	return subjects, nil
}

// GetSubjectsByDepartment gets subjects by department
func GetSubjectsByDepartment(db *sql.DB, departmentID string) ([]*models.Subject, error) {
	query := `SELECT id, name, code, department_id, is_active, created_at, updated_at
			  FROM subjects WHERE department_id = $1 AND is_active = true ORDER BY name`

	rows, err := db.Query(query, departmentID)
	if err != nil {
		return []*models.Subject{}, nil
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
		return []*models.Department{}, nil
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

func GetAllAcademicYears(db *sql.DB) ([]*models.AcademicYear, error) {
	return []*models.AcademicYear{}, nil
}

func GetAcademicYearByID(db *sql.DB, id string) (*models.AcademicYear, error) {
	return &models.AcademicYear{}, nil
}

func CreateAcademicYear(db *sql.DB, year *models.AcademicYear) error {
	return nil
}

func UpdateAcademicYear(db *sql.DB, year *models.AcademicYear) error {
	return nil
}

func DeleteAcademicYear(db *sql.DB, id string) error {
	return nil
}

func GetAllTerms(db *sql.DB) ([]*models.AcademicYear, error) {
	return []*models.AcademicYear{}, nil
}

func GetTermByID(db *sql.DB, id string) (*models.AcademicYear, error) {
	return &models.AcademicYear{}, nil
}

func CreateTerm(db *sql.DB, term *models.Term) error {
	return nil
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
	return []*models.Attendance{}, nil
}

func CreateOrUpdateAttendance(db *sql.DB, attendance *models.Attendance) error {
	return nil
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

func GetDashboardStats(db *sql.DB) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
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
	ID           string     `json:"id"`
	SubjectID    string     `json:"subject_id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	IsCompulsory bool       `json:"is_compulsory"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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
	// Get subjects assigned to the class
	query := `SELECT DISTINCT s.id, s.name, s.code, s.department_id, s.is_active, s.created_at, s.updated_at,
			  d.name as department_name
			  FROM subjects s
			  INNER JOIN class_subjects cs ON s.id = cs.subject_id
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE cs.class_id = $1 AND s.is_active = true AND cs.deleted_at IS NULL
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
			  d.name as department_name
			  FROM subjects s
			  LEFT JOIN departments d ON s.department_id = d.id
			  WHERE s.id = $1 AND s.is_active = true`

	subject := &models.Subject{}
	var departmentName *string
	err := db.QueryRow(query, id).Scan(
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
				ID:        studentID,
				StudentID: studentIDCode,
				FirstName: firstName,
				LastName:  lastName,
				DateOfBirth: dateOfBirth,
				Address:   address,
				ClassID:   classID,
				IsActive:  isActive,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
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
			ID:        studentID,
			StudentID: studentIDCode,
			FirstName: firstName,
			LastName:  lastName,
			DateOfBirth: dateOfBirth,
			Address:   address,
			ClassID:   classID,
			IsActive:  isActive,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
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
		ID:        studentID,
		StudentID: studentIDCode,
		FirstName: firstName,
		LastName:  lastName,
		DateOfBirth: dateOfBirth,
		Address:   address,
		ClassID:   classID,
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
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
	return []*models.Subject{}, nil
}

func CreateSubject(db *sql.DB, subject *models.Subject) error {
	return nil
}

func UpdateSubject(db *sql.DB, subject *models.Subject) error {
	return nil
}

func DeleteSubject(db *sql.DB, id string) error {
	return nil
}
// Additional missing functions
func UpdateTerm(db *sql.DB, term *models.Term) error {
	return nil
}

func DeleteTerm(db *sql.DB, id string) error {
	return nil
}

func GetTermsByAcademicYearID(db *sql.DB, yearID string) ([]*models.Term, error) {
	return []*models.Term{}, nil
}

func SetCurrentAcademicYear(db *sql.DB, yearID string) error {
	return nil
}

func SetCurrentTerm(db *sql.DB, termID string) error {
	return nil
}

func AutoSetCurrentAcademicYear(db *sql.DB) error {
	return nil
}

func AutoSetCurrentTerm(db *sql.DB) error {
	return nil
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
			ID:        studentID,
			StudentID: studentIDCode,
			FirstName: firstName,
			LastName:  lastName,
			DateOfBirth: dateOfBirth,
			Address:   address,
			ClassID:   classID,
			IsActive:  isActive,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
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