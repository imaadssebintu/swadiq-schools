package teachers

import (
	"database/sql"
	"fmt"
	"log"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetTeachersAPI(c *fiber.Ctx) error {
	teachers, err := database.GetAllTeachers(config.GetDB())
	if err != nil {
		return c.JSON(fiber.Map{
			"teachers": []interface{}{},
			"count":    0,
		})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
		"count":    len(teachers),
	})
}

func GetTeachersForSelectionAPI(c *fiber.Ctx) error {
	search := c.Query("search", "")

	// Simple query for teacher selection - only essential fields
	db := config.GetDB()
	query := `SELECT id, first_name, last_name, email FROM users 
			  WHERE deleted_at IS NULL AND is_active = true`
	args := []interface{}{}

	if search != "" {
		query += ` AND (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	query += ` ORDER BY first_name LIMIT 20`

	rows, err := db.Query(query, args...)
	if err != nil {
		return c.JSON(fiber.Map{"teachers": []interface{}{}, "count": 0})
	}
	defer rows.Close()

	var teachers []fiber.Map
	for rows.Next() {
		var id, firstName, lastName, email string
		if err := rows.Scan(&id, &firstName, &lastName, &email); err != nil {
			continue
		}
		teachers = append(teachers, fiber.Map{
			"id":         id,
			"first_name": firstName,
			"last_name":  lastName,
			"email":      email,
		})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
		"count":    len(teachers),
	})
}

func GetTeachersForTimetableAPI(c *fiber.Ctx) error {
	subjectID := c.Query("subject_id")
	paperID := c.Query("paper_id")
	classID := c.Query("class_id")
	dayOfWeek := c.Query("day_of_week")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// If availability parameters are provided, use the availability-aware query
	if dayOfWeek != "" && startTime != "" && endTime != "" {
		db := config.GetDB()

		// Convert day name to number
		dayMap := map[string]int{
			"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3,
			"thursday": 4, "friday": 5, "saturday": 6,
		}
		dayNum, exists := dayMap[dayOfWeek]
		if !exists {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid day_of_week"})
		}

		query := `
			SELECT DISTINCT u.id, u.first_name, u.last_name, u.email
			FROM users u
			INNER JOIN user_roles ur ON u.id = ur.user_id
			INNER JOIN roles r ON ur.role_id = r.id
			INNER JOIN teacher_subjects ts ON u.id = ts.teacher_id
			LEFT JOIN teacher_availability ta ON u.id = ta.teacher_id AND ta.day_of_week = $1
			WHERE u.is_active = true 
			  AND r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin')
			  AND (ta.id IS NULL OR (ta.is_available = true AND ta.start_time <= $2::time AND ta.end_time >= $3::time))
			  AND u.id NOT IN (
				  SELECT te.teacher_id FROM timetable_entries te
				  WHERE te.day_of_week = $4 
					AND te.is_active = true
					AND te.class_id != $5
					AND (
						(te.start_time <= $2::time AND te.end_time > $2::time) OR
						(te.start_time < $3::time AND te.end_time >= $3::time) OR
						(te.start_time >= $2::time AND te.end_time <= $3::time)
					)
			  )`

		args := []interface{}{dayNum, startTime, endTime, dayOfWeek, classID}

		// Add subject/paper filtering
		if subjectID != "" {
			query += ` AND ts.subject_id = $6`
			args = append(args, subjectID)
		}

		if paperID != "" {
			query += ` AND ts.paper_id = $` + fmt.Sprintf("%d", len(args)+1)
			args = append(args, paperID)
		}

		query += ` ORDER BY u.first_name, u.last_name`

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Available teachers fetch error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch available teachers"})
		}
		defer rows.Close()

		var teachers []fiber.Map
		for rows.Next() {
			var id, firstName, lastName, email string
			if err := rows.Scan(&id, &firstName, &lastName, &email); err != nil {
				continue
			}
			teachers = append(teachers, fiber.Map{
				"id":         id,
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"full_name":  firstName + " " + lastName,
			})
		}

		return c.JSON(fiber.Map{
			"teachers": teachers,
		})
	}

	// Fallback to original logic if no availability parameters
	teachers, err := database.GetTeachersBySubjectOrPaper(config.GetDB(), subjectID, paperID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
	})
}

func GetAllTeachersForPaperAPI(c *fiber.Ctx) error {
	paperID := c.Query("paper_id")
	if paperID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Paper ID is required"})
	}

	db := config.GetDB()
	query := `
		SELECT DISTINCT u.id, u.first_name, u.last_name, u.email
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		INNER JOIN teacher_subjects ts ON u.id = ts.teacher_id
		WHERE u.is_active = true 
		  AND r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin')
		  AND ts.paper_id = $1
		ORDER BY u.first_name, u.last_name`

	rows, err := db.Query(query, paperID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch teachers"})
	}
	defer rows.Close()

	var teachers []fiber.Map
	for rows.Next() {
		var id, firstName, lastName, email string
		if err := rows.Scan(&id, &firstName, &lastName, &email); err != nil {
			continue
		}
		teachers = append(teachers, fiber.Map{
			"id":         id,
			"first_name": firstName,
			"last_name":  lastName,
			"email":      email,
			"full_name":  firstName + " " + lastName,
		})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
	})
}

func GetTeacherCountsAPI(c *fiber.Ctx) error {
	counts, err := database.GetTeacherCountsByRole(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch teacher counts"})
	}

	return c.JSON(counts)
}

func GetTeacherStatsAPI(c *fiber.Ctx) error {
	db := config.GetDB()

	// Get teacher counts by role
	roleCounts, err := database.GetTeacherCountsByRole(db)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch teacher counts"})
	}

	// Get total teachers count
	totalQuery := `SELECT COUNT(DISTINCT u.id) FROM users u 
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin') 
		AND u.is_active = true`

	var totalTeachers int
	err = db.QueryRow(totalQuery).Scan(&totalTeachers)
	if err != nil {
		totalTeachers = 0
	}

	return c.JSON(fiber.Map{
		"total_teachers":   totalTeachers,
		"active_teachers":  totalTeachers,
		"class_teachers":   roleCounts["class_teacher"],
		"subject_teachers": roleCounts["subject_teacher"],
	})
}

func CreateTeacherAPI(c *fiber.Ctx) error {
	type CreateTeacherRequest struct {
		FirstName     string   `json:"first_name"`
		LastName      string   `json:"last_name"`
		Email         string   `json:"email"`
		Password      string   `json:"password"`
		Phone         string   `json:"phone"`
		Role          string   `json:"role"`
		DepartmentIDs []string `json:"department_ids"`
		SubjectIDs    []string `json:"subject_ids"`
	}

	var req CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name, last name, email, and password are required"})
	}

	// Default role if not provided
	if req.Role == "" {
		req.Role = "class_teacher"
	}

	// Check phone uniqueness
	if req.Phone != "" {
		taken, err := database.IsPhoneTaken(config.GetDB(), req.Phone, "")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Validation engine failure"})
		}
		if taken {
			return c.Status(400).JSON(fiber.Map{"error": "Phone number is already associated with another account"})
		}
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
	}

	if err := database.CreateTeacher(config.GetDB(), user, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create teacher",
			"details": err.Error(),
		})
	}

	// Link to departments if provided
	if len(req.DepartmentIDs) > 0 {
		if err := database.LinkTeacherToDepartments(config.GetDB(), user.ID, req.DepartmentIDs); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Teacher created but failed to link departments",
				"details": err.Error(),
			})
		}
	}

	// Assign the specified role
	if req.Role != "class_teacher" {
		if err := database.AssignTeacherRole(config.GetDB(), user.ID, req.Role); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Teacher created but failed to assign role",
				"details": err.Error(),
			})
		}
	}

	// Link subjects if provided
	if len(req.SubjectIDs) > 0 {
		if err := database.LinkTeacherToSubjects(config.GetDB(), user.ID, req.SubjectIDs); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Teacher created but failed to link subjects",
				"details": err.Error(),
			})
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Teacher created successfully",
		"teacher": user,
	})
}

func SearchTeachersAPI(c *fiber.Ctx) error {
	query := c.Query("q", "")
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	teachers, total, err := database.SearchTeachersWithPagination(config.GetDB(), query, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to search teachers",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
		"count":    len(teachers),
		"total":    total,
		"query":    query,
		"limit":    limit,
		"offset":   offset,
		"has_more": offset+len(teachers) < total,
	})
}

func GetTeacherAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	teacher, err := database.GetTeacherByID(config.GetDB(), teacherID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Teacher not found"})
	}

	// Fetch classes assigned to this teacher with student counts
	db := config.GetDB()
	query := `SELECT id, name, 
			  (SELECT COUNT(*) FROM students s WHERE s.class_id = t.id AND s.is_active = true) as student_count
			  FROM (
				  SELECT id, name FROM classes WHERE teacher_id = $1 AND is_active = true
				  UNION
				  SELECT DISTINCT c.id, c.name FROM classes c
				  INNER JOIN timetable_entries te ON c.id = te.class_id
				  WHERE te.teacher_id = $1 AND te.is_active = true AND c.is_active = true
			  ) t`

	rows, err := db.Query(query, teacherID)
	if err == nil {
		defer rows.Close()
		classes := make([]*models.Class, 0)
		for rows.Next() {
			cls := &models.Class{}
			if err := rows.Scan(&cls.ID, &cls.Name, &cls.StudentCount); err == nil {
				classes = append(classes, cls)
			}
		}
		teacher.Classes = classes
	}

	return c.JSON(fiber.Map{
		"teacher": teacher,
	})
}

func UpdateTeacherAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	type UpdateTeacherRequest struct {
		FirstName string   `json:"first_name"`
		LastName  string   `json:"last_name"`
		Email     string   `json:"email"`
		Phone     string   `json:"phone"`
		Role      string   `json:"role"`  // Maintain for backward compatibility
		Roles     []string `json:"roles"` // Support multi-role
	}

	var req UpdateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Check phone uniqueness (excluding self)
	if req.Phone != "" {
		taken, err := database.IsPhoneTaken(config.GetDB(), req.Phone, teacherID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Validation engine failure"})
		}
		if taken {
			return c.Status(400).JSON(fiber.Map{"error": "Phone number is already associated with another account"})
		}
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name, last name, and email are required"})
	}

	// Check if teacher exists
	existingTeacher, err := database.GetTeacherByID(config.GetDB(), teacherID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Teacher not found"})
	}

	// Update teacher data
	existingTeacher.FirstName = req.FirstName
	existingTeacher.LastName = req.LastName
	existingTeacher.Email = req.Email
	existingTeacher.Phone = req.Phone

	if err := database.UpdateTeacher(config.GetDB(), existingTeacher); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update teacher"})
	}

	// Update roles if provided
	rolesToAssign := req.Roles
	if len(rolesToAssign) == 0 && req.Role != "" {
		rolesToAssign = []string{req.Role}
	}

	if len(rolesToAssign) > 0 {
		// Remove existing teacher roles first
		db := config.GetDB()
		_, err := db.Exec("DELETE FROM user_roles WHERE user_id = $1 AND role_id IN (SELECT id FROM roles WHERE name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin'))", teacherID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to reset clearances"})
		}

		// Assign new roles
		for _, roleName := range rolesToAssign {
			if err := database.AssignTeacherRole(config.GetDB(), teacherID, roleName); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to assign clearance: " + roleName})
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Teacher updated successfully",
		"teacher": existingTeacher,
	})
}

func CheckPhoneUniquenessAPI(c *fiber.Ctx) error {
	phone := c.Query("phone")
	excludeID := c.Query("exclude_id")

	if phone == "" {
		return c.JSON(fiber.Map{"taken": false})
	}

	taken, err := database.IsPhoneTaken(config.GetDB(), phone, excludeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal validation error"})
	}

	return c.JSON(fiber.Map{
		"taken": taken,
		"phone": phone,
	})
}

func DeleteTeacherAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	// Check if teacher exists
	_, err := database.GetTeacherByID(config.GetDB(), teacherID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Teacher not found"})
	}

	// TODO: Check if teacher has assigned classes before deleting
	// For now, we'll do a soft delete
	if err := database.DeleteTeacher(config.GetDB(), teacherID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete teacher"})
	}

	return c.JSON(fiber.Map{
		"message": "Teacher deleted successfully",
	})
}

func GetRolesAPI(c *fiber.Ctx) error {
	query := `SELECT id, name, is_active, created_at, updated_at FROM roles ORDER BY name`
	rows, err := config.GetDB().Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch roles"})
	}
	defer rows.Close()

	var roles []map[string]interface{}
	for rows.Next() {
		var id, name string
		var isActive bool
		var createdAt, updatedAt string

		err := rows.Scan(&id, &name, &isActive, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		roles = append(roles, map[string]interface{}{
			"id":         id,
			"name":       name,
			"is_active":  isActive,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
		"count": len(roles),
	})
}

func CreateRoleAPI(c *fiber.Ctx) error {
	type CreateRoleRequest struct {
		Name string `json:"name"`
	}

	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role name is required"})
	}

	query := `INSERT INTO roles (name, is_active, created_at, updated_at) VALUES ($1, true, NOW(), NOW())`
	_, err := config.GetDB().Exec(query, req.Name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create role"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Role created successfully",
	})
}

func GetDepartmentOverviewAPI(c *fiber.Ctx) error {
	db := config.GetDB()

	query := `SELECT d.id, d.name, d.code,
		h.first_name as head_first_name, h.last_name as head_last_name,
		a.first_name as assistant_first_name, a.last_name as assistant_last_name,
		COUNT(DISTINCT ud.user_id) as teacher_count
		FROM departments d
		LEFT JOIN users h ON d.head_of_department_id = h.id AND h.is_active = true
		LEFT JOIN users a ON d.assistant_head_id = a.id AND a.is_active = true
		LEFT JOIN user_departments ud ON d.id = ud.department_id
		LEFT JOIN users u ON ud.user_id = u.id AND u.is_active = true
		WHERE d.is_active = true
		GROUP BY d.id, d.name, d.code, h.first_name, h.last_name, a.first_name, a.last_name
		ORDER BY d.name`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch department overview"})
	}
	defer rows.Close()

	departments := make([]map[string]interface{}, 0)
	for rows.Next() {
		var deptID, deptName, deptCode string
		var headFirstName, headLastName, assistantFirstName, assistantLastName *string
		var teacherCount int

		if err := rows.Scan(&deptID, &deptName, &deptCode, &headFirstName, &headLastName, &assistantFirstName, &assistantLastName, &teacherCount); err == nil {
			headName := "Not assigned"
			if headFirstName != nil && headLastName != nil {
				headName = *headFirstName + " " + *headLastName
			}

			assistantName := "Not assigned"
			if assistantFirstName != nil && assistantLastName != nil {
				assistantName = *assistantFirstName + " " + *assistantLastName
			}

			departments = append(departments, map[string]interface{}{
				"id":             deptID,
				"name":           deptName,
				"code":           deptCode,
				"teacher_count":  teacherCount,
				"head_name":      headName,
				"assistant_name": assistantName,
			})
		}
	}

	return c.JSON(fiber.Map{
		"departments": departments,
		"count":       len(departments),
	})
}

func GetTeacherSubjectsAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	db := config.GetDB()

	query := `
		SELECT 
			s.id as subject_id, 
			s.name as subject_name, 
			s.code as subject_code, 
			s.department_id,
			p.id as paper_id,
			p.name as paper_name,
			p.code as paper_code
		FROM 
			subjects s
		JOIN 
			teacher_subjects ts ON s.id = ts.subject_id
		LEFT JOIN 
			papers p ON ts.paper_id = p.id AND p.deleted_at IS NULL
		WHERE 
			ts.teacher_id = $1 AND s.deleted_at IS NULL
		ORDER BY 
			s.name, p.name;
	`

	rows, err := db.Query(query, teacherID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch teacher subjects"})
	}
	defer rows.Close()

	subjectsMap := make(map[string]fiber.Map)
	var subjectsOrder []string

	for rows.Next() {
		var subjectID, subjectName, subjectCode, departmentID string
		var paperID, paperName, paperCode *string // Use pointers for nullable fields

		if err := rows.Scan(&subjectID, &subjectName, &subjectCode, &departmentID, &paperID, &paperName, &paperCode); err != nil {
			continue
		}

		if _, ok := subjectsMap[subjectID]; !ok {
			subjectsMap[subjectID] = fiber.Map{
				"id":            subjectID,
				"name":          subjectName,
				"code":          subjectCode,
				"department_id": departmentID,
				"papers":        []fiber.Map{},
			}
			subjectsOrder = append(subjectsOrder, subjectID)
		}

		if paperID != nil {
			papers := subjectsMap[subjectID]["papers"].([]fiber.Map)
			subjectsMap[subjectID]["papers"] = append(papers, fiber.Map{
				"id":   *paperID,
				"name": *paperName,
				"code": *paperCode,
			})
		}
	}

	// Create the final subjects slice in order
	subjects := make([]fiber.Map, len(subjectsOrder))
	for i, subjectID := range subjectsOrder {
		subjects[i] = subjectsMap[subjectID]
	}

	return c.JSON(fiber.Map{
		"subjects": subjects,
		"count":    len(subjects),
	})
}

func AssignTeacherSubjectsAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	db := config.GetDB()

	type AssignSubjectsRequest struct {
		SubjectIDs []string            `json:"subject_ids"`
		Papers     map[string][]string `json:"papers"` // subject_id -> paper_ids
	}

	var req AssignSubjectsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Delete existing assignments
	_, err = tx.Exec("DELETE FROM teacher_subjects WHERE teacher_id = $1", teacherID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear existing subjects"})
	}

	// Insert assignments
	for _, subjectID := range req.SubjectIDs {
		if paperIDs, exists := req.Papers[subjectID]; exists && len(paperIDs) > 0 {
			// Subject has papers, insert only paper assignments
			for _, paperID := range paperIDs {
				_, err = tx.Exec("INSERT INTO teacher_subjects (teacher_id, subject_id, paper_id) VALUES ($1, $2, $3)", teacherID, subjectID, paperID)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to assign paper"})
				}
			}
		} else {
			// Subject has no papers, insert subject assignment
			_, err = tx.Exec("INSERT INTO teacher_subjects (teacher_id, subject_id) VALUES ($1, $2)", teacherID, subjectID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to assign subject"})
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit changes"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subjects and papers assigned successfully",
	})
}

func GetTeacherAvailabilityAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	availability, err := database.GetTeacherAvailability(config.GetDB(), teacherID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get teacher availability"})
	}

	// If no availability is found, return a default schedule for all 7 days
	if len(availability) == 0 {
		defaultAvailability := make([]*models.TeacherAvailability, 7)
		for i := 0; i < 7; i++ {
			defaultAvailability[i] = &models.TeacherAvailability{
				TeacherID:   teacherID,
				DayOfWeek:   i,
				IsAvailable: false,
				StartTime:   sql.NullString{},
				EndTime:     sql.NullString{},
			}
		}
		return c.JSON(fiber.Map{
			"availability": defaultAvailability,
		})
	}

	return c.JSON(fiber.Map{
		"availability": availability,
	})
}

func UpdateTeacherAvailabilityAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	var availability []*models.TeacherAvailability
	if err := c.BodyParser(&availability); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := database.UpdateTeacherAvailability(config.GetDB(), teacherID, availability); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update teacher availability"})
	}

	return c.JSON(fiber.Map{
		"message": "Availability updated successfully",
	})
}

func RemoveTeacherSubjectAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	subjectID := c.Params("subjectId")
	paperID := c.Query("paper_id")

	db := config.GetDB()

	var query string
	var args []interface{}

	if paperID != "" {
		query = "DELETE FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2 AND paper_id = $3"
		args = []interface{}{teacherID, subjectID, paperID}
	} else {
		query = "DELETE FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2 AND paper_id IS NULL"
		args = []interface{}{teacherID, subjectID}
	}

	result, err := db.Exec(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to remove subject"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Assignment not found"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subject removed successfully",
	})
}

// Salary API Handlers

func GetTeacherSalaryAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	salary, err := database.GetTeacherSalary(config.GetDB(), teacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty/null if no salary set, not 404
			return c.JSON(fiber.Map{"salary": nil})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch salary"})
	}

	// Calculate Payroll Status for current month
	// Default to current month viewing
	// TODO: Allow query params for custom range?
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	// Use today as end date for "Accrued to Date"
	// Or use end of month? Accrued usually means "earned so far". So today.
	endOfPeriod := now

	payrollStatus, err := database.GetTeacherPayrollStatus(config.GetDB(), teacherID, firstOfMonth, endOfPeriod)
	// Ignore error for now (or log it), return partial data
	if err != nil {
		// Just log? fmt.Println(err)
	}

	return c.JSON(fiber.Map{
		"salary":  salary,
		"payroll": payrollStatus,
	})
}

func SetTeacherSalaryAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	if teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	var req models.TeacherSalary
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Enforce user ID from URL
	req.UserID = teacherID

	// Enforce Allowance Constraints
	if req.Period == "day" {
		// Daily Salary: No Allowance Allowed
		req.HasAllowance = false
		req.Allowance = 0
		req.AllowancePeriod = ""
	} else if req.Period == "week" {
		// Weekly Salary: Daily Allowance ONLY
		if req.HasAllowance {
			req.AllowancePeriod = "day"
		}
	} else if req.Period == "month" {
		// Monthly Salary: Allow Daily or Weekly
		// Validate/Normalize Allowance Period
		if req.HasAllowance {
			if req.AllowancePeriod != "day" && req.AllowancePeriod != "week" {
				req.AllowancePeriod = "week" // Default
			}
		}
	}

	if err := database.UpsertTeacherSalary(config.GetDB(), &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to set salary",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Salary set successfully",
		"salary":  req,
	})
}
