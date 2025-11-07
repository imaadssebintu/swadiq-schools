package classes

import (
	"strings"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetClassesAPI(c *fiber.Ctx) error {
	classes, err := database.GetAllClasses(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch classes"})
	}

	return c.JSON(fiber.Map{
		"classes": classes,
		"count":   len(classes),
	})
}

// GetClassesStatsAPI returns classes statistics for the classes page
func GetClassesStatsAPI(c *fiber.Ctx) error {
	db := config.GetDB()

	// Get classes statistics
	stats := make(map[string]interface{})

	// Total Classes
	var totalClasses int
	err := db.QueryRow("SELECT COUNT(*) FROM classes WHERE is_active = true").Scan(&totalClasses)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch classes statistics",
			"details": err.Error(),
		})
	}

	// Active Classes (same as total for now)
	activeClasses := totalClasses

	// Classes with students
	var classesWithStudents int
	err = db.QueryRow(`SELECT COUNT(DISTINCT c.id) FROM classes c
					   INNER JOIN students s ON c.id = s.class_id
					   WHERE c.is_active = true AND s.is_active = true`).Scan(&classesWithStudents)
	if err != nil {
		classesWithStudents = 0 // Default to 0 if query fails
	}

	// Classes without teachers
	var classesWithoutTeachers int
	err = db.QueryRow("SELECT COUNT(*) FROM classes WHERE is_active = true AND teacher_id IS NULL").Scan(&classesWithoutTeachers)
	if err != nil {
		classesWithoutTeachers = 0
	}

	stats["total_classes"] = totalClasses
	stats["active_classes"] = activeClasses
	stats["classes_with_students"] = classesWithStudents
	stats["classes_without_teachers"] = classesWithoutTeachers

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetClassesTableAPI returns classes formatted for table display
func GetClassesTableAPI(c *fiber.Ctx) error {
	search := c.Query("search", "")

	classes, err := database.GetAllClasses(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch classes"})
	}

	// Filter classes based on search if provided
	if search != "" {
		var filteredClasses []*models.Class
		searchLower := strings.ToLower(search)

		for _, class := range classes {
			if strings.Contains(strings.ToLower(class.Name), searchLower) ||
				(class.Teacher != nil &&
					(strings.Contains(strings.ToLower(class.Teacher.FirstName), searchLower) ||
						strings.Contains(strings.ToLower(class.Teacher.LastName), searchLower))) {
				filteredClasses = append(filteredClasses, class)
			}
		}
		classes = filteredClasses
	}

	return c.JSON(fiber.Map{
		"success": true,
		"classes": classes,
		"count":   len(classes),
	})
}

func CreateClassAPI(c *fiber.Ctx) error {
	type CreateClassRequest struct {
		Name      string `json:"name"`
		Code      string `json:"code"`
		TeacherID string `json:"teacher_id"`
	}

	var req CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class name is required"})
	}

	if req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class code is required"})
	}

	class := &models.Class{
		Name: req.Name,
		Code: &req.Code,
	}

	if req.TeacherID != "" {
		class.TeacherID = &req.TeacherID
	}

	if err := database.CreateClass(config.GetDB(), class); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create class"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Class created successfully",
		"class":   class,
	})
}

func GetClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	class, err := GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Class not found"})
	}

	return c.JSON(fiber.Map{
		"class": class,
	})
}

// GetClassStatisticsAPI returns accurate statistics for a specific class
func GetClassStatisticsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	stats, err := database.GetClassStatistics(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch class statistics"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetClassStudentsAPI returns accurate list of students for a specific class
func GetClassStudentsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	students, err := database.GetClassStudents(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch class students"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"students": students,
		"count":   len(students),
	})
}

// GetClassDetailsAPI returns detailed information about a specific class
func GetClassDetailsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	db := config.GetDB()

	// Get basic student list (minimal data)
	students := []map[string]interface{}{}
	rows, err := db.Query(`
		SELECT id, student_id, first_name, last_name, date_of_birth, gender, address 
		FROM students 
		WHERE class_id = $1 AND is_active = true 
		ORDER BY first_name, last_name
	`, classID)
	
	var totalStudents, maleCount, femaleCount int
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, studentID, firstName, lastName, gender string
			var dateOfBirth *time.Time
			var address *string
			if err := rows.Scan(&id, &studentID, &firstName, &lastName, &dateOfBirth, &gender, &address); err == nil {
				addressStr := ""
				if address != nil {
					addressStr = *address
				}
				
				dateStr := ""
				if dateOfBirth != nil {
					dateStr = dateOfBirth.Format("2006-01-02")
				}
				
				students = append(students, map[string]interface{}{
					"id":            id,
					"student_id":    studentID,
					"first_name":    firstName,
					"last_name":     lastName,
					"date_of_birth": dateStr,
					"gender":        gender,
					"address":       addressStr,
				})
				totalStudents++
				if gender == "male" {
					maleCount++
				} else if gender == "female" {
					femaleCount++
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"class": map[string]interface{}{
			"id":   classID,
			"name": "Class Details",
		},
		"students": students,
		"statistics": map[string]interface{}{
			"total_students":  totalStudents,
			"male_students":   maleCount,
			"female_students": femaleCount,
		},
		"promotion_settings":          nil,
		"available_promotion_classes": nil,
	})
}

func UpdateClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type UpdateClassRequest struct {
		Name      string `json:"name"`
		Code      string `json:"code"`
		TeacherID string `json:"teacher_id"`
	}

	var req UpdateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class name is required"})
	}

	if req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class code is required"})
	}

	// Check if class exists
	existingClass, err := GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Class not found"})
	}

	// Update class data
	existingClass.Name = req.Name
	existingClass.Code = &req.Code
	if req.TeacherID != "" {
		existingClass.TeacherID = &req.TeacherID
	} else {
		existingClass.TeacherID = nil
	}

	if err := UpdateClass(config.GetDB(), existingClass); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update class"})
	}

	return c.JSON(fiber.Map{
		"message": "Class updated successfully",
		"class":   existingClass,
	})
}

func DeleteClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	// Check if class exists
	_, err := GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Class not found"})
	}

	// TODO: Check if class has students before deleting
	// For now, we'll do a soft delete
	if err := DeleteClass(config.GetDB(), classID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete class"})
	}

	return c.JSON(fiber.Map{
		"message": "Class deleted successfully",
	})
}

// UpdateClassPromotionSettingsAPI updates promotion settings for a class
func UpdateClassPromotionSettingsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type PromotionRequest struct {
		ToClassID         string  `json:"to_class_id"`
		AcademicYearID    *string `json:"academic_year_id"`
		PromotionCriteria string  `json:"promotion_criteria"`
	}

	var req PromotionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.ToClassID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Target class ID is required"})
	}

	promotion := &models.ClassPromotion{
		FromClassID:       classID,
		ToClassID:         req.ToClassID,
		AcademicYearID:    req.AcademicYearID,
		PromotionCriteria: req.PromotionCriteria,
	}

	if err := SaveClassPromotionSettings(config.GetDB(), promotion); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save promotion settings"})
	}

	return c.JSON(fiber.Map{
		"success":            true,
		"message":            "Promotion settings updated successfully",
		"promotion_settings": promotion,
	})
}

// GetClassSubjectsAPI returns subjects assigned to a class with papers and teachers
func GetClassSubjectsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	subjects, err := database.GetClassSubjectsWithPapers(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch class subjects"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"subjects": subjects,
	})
}

// AddClassSubjectsAPI adds subjects to a class
func AddClassSubjectsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type AddSubjectsRequest struct {
		Subjects []database.SubjectAssignment `json:"subjects"`
	}

	var req AddSubjectsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if len(req.Subjects) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "At least one subject is required"})
	}

	if err := database.AddSubjectsToClassWithCompulsory(config.GetDB(), classID, req.Subjects); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add subjects to class"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subjects added to class successfully",
	})
}

// AddStudentToClassAPI adds a student to a class
func AddStudentToClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type AddStudentRequest struct {
		StudentID string `json:"student_id"`
	}

	var req AddStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.StudentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
	}

	// Update student's class_id
	db := config.GetDB()
	_, err := db.Exec("UPDATE students SET class_id = $1 WHERE id = $2", classID, req.StudentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add student to class"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student added to class successfully",
	})
}

// GetClassPapersAPI returns papers assigned to a class
func GetClassPapersAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	papers, err := database.GetClassPapers(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch class papers"})
	}

	return c.JSON(papers)
}

// AssignPapersToClassAPI assigns papers to a class using ClassPaper model
func AssignPapersToClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type PaperAssignment struct {
		PaperID   string  `json:"paper_id"`
		TeacherID *string `json:"teacher_id"`
	}

	type AssignPapersRequest struct {
		PaperAssignments []PaperAssignment `json:"paper_assignments"`
	}

	var req AssignPapersRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	if len(req.PaperAssignments) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No paper assignments provided"})
	}

	db := config.GetDB()
	var createdIDs []string

	// Process each paper assignment using ClassPaper model
	for _, pa := range req.PaperAssignments {
		var existingID string
		checkQuery := `SELECT id FROM class_papers WHERE class_id = $1 AND paper_id = $2 AND deleted_at IS NULL`
		err := db.QueryRow(checkQuery, classID, pa.PaperID).Scan(&existingID)

		if err != nil {
			// Create new class paper
			query := `INSERT INTO class_papers (class_id, paper_id, teacher_id, created_at, updated_at)
					  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`
			
			var classPaperID string
			err = db.QueryRow(query, classID, pa.PaperID, pa.TeacherID).Scan(&classPaperID)
			if err != nil {
				continue
			}
			createdIDs = append(createdIDs, classPaperID)
		} else {
			// Update existing class paper
			updateQuery := `UPDATE class_papers SET teacher_id = $1, updated_at = NOW() WHERE id = $2`
			db.Exec(updateQuery, pa.TeacherID, existingID)
			createdIDs = append(createdIDs, existingID)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Papers assigned successfully",
		"ids":     createdIDs,
	})
}

// GetSubjectPapersForClassAPI returns papers for a subject with assignment status
func GetSubjectPapersForClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	subjectID := c.Params("subjectId")
	
	if classID == "" || subjectID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID and Subject ID are required"})
	}

	papers, err := database.GetSubjectPapersForClass(config.GetDB(), classID, subjectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch papers"})
	}

	return c.JSON(papers)
}

// RemoveClassSubjectAPI removes a subject from a class
func RemoveClassSubjectAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	subjectID := c.Params("subjectId")
	
	if classID == "" || subjectID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID and Subject ID are required"})
	}

	if err := database.RemoveSubjectFromClass(config.GetDB(), classID, subjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to remove subject from class"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subject removed from class successfully",
	})
}

// AssignTeacherToPaperAPI assigns a teacher to a specific paper in a class
func AssignTeacherToPaperAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	paperID := c.Params("paperId")
	
	if classID == "" || paperID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID and Paper ID are required"})
	}

	type AssignTeacherRequest struct {
		TeacherID string `json:"teacher_id"`
	}

	var req AssignTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	db := config.GetDB()

	// Check if class paper already exists
	var existingID string
	checkQuery := `SELECT id FROM class_papers WHERE class_id = $1 AND paper_id = $2 AND deleted_at IS NULL`
	err := db.QueryRow(checkQuery, classID, paperID).Scan(&existingID)

	if err != nil {
		// Create new class paper
		query := `INSERT INTO class_papers (class_id, paper_id, teacher_id, created_at, updated_at)
				  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`
		
		var classPaperID string
		err = db.QueryRow(query, classID, paperID, req.TeacherID).Scan(&classPaperID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to assign teacher to paper"})
		}

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"message": "Teacher assigned to paper successfully",
			"id":      classPaperID,
		})
	}

	// Update existing class paper
	updateQuery := `UPDATE class_papers SET teacher_id = $1, updated_at = NOW() WHERE id = $2`
	_, err = db.Exec(updateQuery, req.TeacherID, existingID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update teacher assignment"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Teacher assignment updated successfully",
		"id":      existingID,
	})
}
