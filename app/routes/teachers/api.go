package teachers

import (
	"database/sql"
	"fmt"
	"log"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

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

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
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
	query := `SELECT c.id, c.name, 
			  (SELECT COUNT(*) FROM students s WHERE s.class_id = c.id AND s.is_active = true) as student_count
			  FROM classes c WHERE c.teacher_id = $1 AND c.is_active = true`

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
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Role      string `json:"role"`
	}

	var req UpdateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
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

	if err := database.UpdateTeacher(config.GetDB(), existingTeacher); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update teacher"})
	}

	// Update role if provided
	if req.Role != "" {
		// Remove existing teacher roles first
		db := config.GetDB()
		_, err := db.Exec("DELETE FROM user_roles WHERE user_id = $1 AND role_id IN (SELECT id FROM roles WHERE name IN ('class_teacher', 'subject_teacher', 'head_teacher', 'admin'))", teacherID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update role"})
		}

		// Assign new role
		if err := database.AssignTeacherRole(config.GetDB(), teacherID, req.Role); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to assign new role"})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Teacher updated successfully",
		"teacher": existingTeacher,
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

	type AvailabilityRequestItem struct {
		DayOfWeek   int     `json:"day_of_week"`
		IsAvailable bool    `json:"is_available"`
		StartTime   *string `json:"start_time"`
		EndTime     *string `json:"end_time"`
	}

	type UpdateAvailabilityRequest struct {
		Availability []*AvailabilityRequestItem `json:"availability"`
	}

	var req UpdateAvailabilityRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("BodyParser error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Convert to []*models.TeacherAvailability
	availabilityToUpdate := make([]*models.TeacherAvailability, len(req.Availability))
	for i, item := range req.Availability {
		availabilityToUpdate[i] = &models.TeacherAvailability{
			DayOfWeek:   item.DayOfWeek,
			IsAvailable: item.IsAvailable,
			StartTime:   sql.NullString{String: "", Valid: false},
			EndTime:     sql.NullString{String: "", Valid: false},
		}
		if item.StartTime != nil {
			availabilityToUpdate[i].StartTime.String = *item.StartTime
			availabilityToUpdate[i].StartTime.Valid = true
		}
		if item.EndTime != nil {
			availabilityToUpdate[i].EndTime.String = *item.EndTime
			availabilityToUpdate[i].EndTime.Valid = true
		}
	}

	err := database.UpdateTeacherAvailability(config.GetDB(), teacherID, availabilityToUpdate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update teacher availability"})
	}

	return c.JSON(fiber.Map{
		"message": "Teacher availability updated successfully",
	})
}

func RemoveTeacherSubjectAPI(c *fiber.Ctx) error {
	teacherID := c.Params("id")
	subjectID := c.Params("subjectId")
	paperID := c.Query("paper_id")

	db := config.GetDB()

	if paperID != "" {
		// Remove specific paper
		_, err := db.Exec(`DELETE FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2 AND paper_id = $3`, teacherID, subjectID, paperID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to remove paper"})
		}

		// Check if this was the last paper for this subject
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2`, teacherID, subjectID).Scan(&count)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check remaining papers"})
		}

		// If no papers left, remove the subject entirely
		if count == 0 {
			_, err = db.Exec(`DELETE FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2`, teacherID, subjectID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to remove subject"})
			}
		}
	} else {
		// Remove entire subject (all papers)
		_, err := db.Exec(`DELETE FROM teacher_subjects WHERE teacher_id = $1 AND subject_id = $2`, teacherID, subjectID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to remove subject"})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subject/paper removed successfully",
	})
}
