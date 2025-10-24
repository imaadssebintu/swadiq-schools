package classes

import (
	"database/sql"
	"fmt"
	"strings"
	"swadiq-schools/app/models"
	"time"
)

// GetClassByID retrieves a single class by ID with teacher information
func GetClassByID(db *sql.DB, classID string) (*models.Class, error) {
	query := `
		SELECT c.id, c.name, c.code, c.teacher_id, c.is_active, c.created_at, c.updated_at,
		       u.id, u.first_name, u.last_name, u.email
		FROM classes c
		LEFT JOIN users u ON c.teacher_id = u.id
		WHERE c.id = $1 AND c.is_active = true
	`

	var class models.Class
	var teacher models.User
	var teacherID sql.NullString
	var teacherUserID, teacherFirstName, teacherLastName, teacherEmail sql.NullString

	err := db.QueryRow(query, classID).Scan(
		&class.ID, &class.Name, &class.Code, &teacherID, &class.IsActive, &class.CreatedAt, &class.UpdatedAt,
		&teacherUserID, &teacherFirstName, &teacherLastName, &teacherEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("class not found")
		}
		return nil, err
	}

	// Set teacher information if exists
	if teacherID.Valid && teacherUserID.Valid {
		teacher.ID = teacherUserID.String
		teacher.FirstName = teacherFirstName.String
		teacher.LastName = teacherLastName.String
		teacher.Email = teacherEmail.String
		class.Teacher = &teacher
		class.TeacherID = &teacherID.String
	}

	return &class, nil
}

// UpdateClass updates an existing class in the database
func UpdateClass(db *sql.DB, class *models.Class) error {
	query := `
		UPDATE classes 
		SET name = $1, teacher_id = $2, updated_at = NOW()
		WHERE id = $3 AND is_active = true
	`

	var teacherID interface{}
	if class.TeacherID != nil && *class.TeacherID != "" {
		teacherID = *class.TeacherID
	} else {
		teacherID = nil
	}

	_, err := db.Exec(query, class.Name, teacherID, class.ID)
	if err != nil {
		return fmt.Errorf("failed to update class: %v", err)
	}

	// Update the timestamp
	class.UpdatedAt = time.Now()

	return nil
}

// DeleteClass soft deletes a class (sets is_active = false)
func DeleteClass(db *sql.DB, classID string) error {
	query := `
		UPDATE classes 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := db.Exec(query, classID)
	if err != nil {
		return fmt.Errorf("failed to delete class: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("class not found or already deleted")
	}

	return nil
}

// ValidateClassName validates class name format and uniqueness
func ValidateClassName(db *sql.DB, name string, excludeID string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("class name cannot be empty")
	}

	if len(name) < 2 {
		return fmt.Errorf("class name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("class name cannot exceed 50 characters")
	}

	// Check for duplicate names (case-insensitive)
	query := `
		SELECT COUNT(*) 
		FROM classes 
		WHERE LOWER(name) = LOWER($1) 
		AND is_active = true
		AND ($2 = '' OR id != $2)
	`

	var count int
	err := db.QueryRow(query, name, excludeID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate class name: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("class name '%s' already exists", name)
	}

	return nil
}

// GetClassStudentCount returns the number of students in a class
func GetClassStudentCount(db *sql.DB, classID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM students 
		WHERE class_id = $1 AND is_active = true
	`

	var count int
	err := db.QueryRow(query, classID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get student count: %v", err)
	}

	return count, nil
}

// GetClassSubjectCount returns the number of subjects assigned to a class
func GetClassSubjectCount(db *sql.DB, classID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM class_subjects cs
		JOIN subjects s ON cs.subject_id = s.id
		WHERE cs.class_id = $1 AND s.is_active = true
	`

	var count int
	err := db.QueryRow(query, classID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get subject count: %v", err)
	}

	return count, nil
}

// CanDeleteClass checks if a class can be safely deleted
func CanDeleteClass(db *sql.DB, classID string) (bool, string, error) {
	// Check if class has students
	studentCount, err := GetClassStudentCount(db, classID)
	if err != nil {
		return false, "", err
	}

	if studentCount > 0 {
		return false, fmt.Sprintf("Cannot delete class with %d students. Please move students to another class first.", studentCount), nil
	}

	// Check if class has upcoming exams or other dependencies
	// This can be extended based on your needs

	return true, "", nil
}

// FormatClassName formats class name for display
func FormatClassName(name string) string {
	return strings.TrimSpace(name)
}

// GenerateClassCode generates a unique class code if needed
func GenerateClassCode(name string) string {
	// Simple implementation - can be enhanced
	cleanName := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(cleanName) > 5 {
		cleanName = cleanName[:5]
	}
	return cleanName
}

// GetClassPromotionSettings retrieves promotion settings for a class
func GetClassPromotionSettings(db *sql.DB, classID string) (*models.ClassPromotion, error) {
	// Check if table exists first
	var tableExists bool
	checkQuery := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'class_promotions')`
	err := db.QueryRow(checkQuery).Scan(&tableExists)
	if err != nil || !tableExists {
		return nil, nil // Table doesn't exist, return nil
	}

	query := `
		SELECT cp.id, cp.from_class_id, cp.to_class_id, cp.academic_year_id,
		       cp.promotion_criteria, cp.is_active, cp.created_at, cp.updated_at,
		       tc.id, tc.name
		FROM class_promotions cp
		LEFT JOIN classes tc ON cp.to_class_id = tc.id
		WHERE cp.from_class_id = $1 AND cp.is_active = true
		ORDER BY cp.created_at DESC
		LIMIT 1
	`

	promotion := &models.ClassPromotion{
		ToClass: &models.Class{},
	}

	var academicYearID sql.NullString
	var toClassID, toClassName sql.NullString

	err = db.QueryRow(query, classID).Scan(
		&promotion.ID, &promotion.FromClassID, &promotion.ToClassID, &academicYearID,
		&promotion.PromotionCriteria, &promotion.IsActive, &promotion.CreatedAt, &promotion.UpdatedAt,
		&toClassID, &toClassName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No promotion settings found
		}
		return nil, err
	}

	if academicYearID.Valid {
		promotion.AcademicYearID = &academicYearID.String
	}

	if toClassID.Valid && toClassName.Valid {
		promotion.ToClass.ID = toClassID.String
		promotion.ToClass.Name = toClassName.String
	}

	return promotion, nil
}

// GetAvailablePromotionClasses returns classes available for promotion (excluding current class)
func GetAvailablePromotionClasses(db *sql.DB, currentClassID string) ([]*models.Class, error) {
	query := `
		SELECT id, name, teacher_id, is_active, created_at, updated_at
		FROM classes
		WHERE is_active = true AND id != $1
		ORDER BY name
	`

	rows, err := db.Query(query, currentClassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		var teacherID sql.NullString

		err := rows.Scan(
			&class.ID, &class.Name, &teacherID,
			&class.IsActive, &class.CreatedAt, &class.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if teacherID.Valid {
			class.TeacherID = &teacherID.String
		}

		classes = append(classes, class)
	}

	return classes, nil
}

// SaveClassPromotionSettings saves or updates promotion settings for a class
func SaveClassPromotionSettings(db *sql.DB, promotion *models.ClassPromotion) error {
	// Check if table exists first
	var tableExists bool
	checkTableQuery := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'class_promotions')`
	err := db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil || !tableExists {
		return fmt.Errorf("class_promotions table does not exist")
	}

	// Check if promotion settings already exist
	existingID := ""
	checkQuery := `SELECT id FROM class_promotions WHERE from_class_id = $1 AND is_active = true LIMIT 1`
	db.QueryRow(checkQuery, promotion.FromClassID).Scan(&existingID)

	if existingID != "" {
		// Update existing
		query := `
			UPDATE class_promotions
			SET to_class_id = $1, academic_year_id = $2, promotion_criteria = $3, updated_at = NOW()
			WHERE id = $4
		`
		_, err := db.Exec(query, promotion.ToClassID, promotion.AcademicYearID, promotion.PromotionCriteria, existingID)
		promotion.ID = existingID
		return err
	} else {
		// Insert new
		query := `
			INSERT INTO class_promotions (from_class_id, to_class_id, academic_year_id, promotion_criteria, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, true, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`
		err := db.QueryRow(query, promotion.FromClassID, promotion.ToClassID, promotion.AcademicYearID, promotion.PromotionCriteria).Scan(
			&promotion.ID, &promotion.CreatedAt, &promotion.UpdatedAt,
		)
		return err
	}
}
