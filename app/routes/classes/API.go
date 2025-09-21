package classes

import (
	"strings"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

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
		TeacherID string `json:"teacher_id"`
	}

	var req CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class name is required"})
	}

	class := &models.Class{
		Name: req.Name,
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

// GetClassDetailsAPI returns detailed information about a specific class
func GetClassDetailsAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	db := config.GetDB()

	// Get class basic info
	class, err := GetClassByID(db, classID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Class not found"})
	}

	// Get students in this class
	students, err := database.GetStudentsByClass(db, classID)
	if err != nil {
		students = []*models.Student{} // Default to empty if error
	}

	// Get class statistics
	stats := make(map[string]interface{})

	// Student count by gender
	var maleCount, femaleCount int
	for _, student := range students {
		if student.Gender != nil {
			if *student.Gender == "male" {
				maleCount++
			} else if *student.Gender == "female" {
				femaleCount++
			}
		}
	}

	// Get promotion settings for this class (with safe error handling)
	var promotionSettings *models.ClassPromotion
	if settings, err := GetClassPromotionSettings(db, classID); err == nil {
		promotionSettings = settings
	}

	// Get available classes for promotion (with safe error handling)
	availableClasses := []*models.Class{}
	if classes, err := GetAvailablePromotionClasses(db, classID); err == nil {
		availableClasses = classes
	}

	stats["total_students"] = len(students)
	stats["male_students"] = maleCount
	stats["female_students"] = femaleCount
	stats["active_students"] = len(students) // All returned students are active

	return c.JSON(fiber.Map{
		"success":                     true,
		"class":                       class,
		"students":                    students,
		"statistics":                  stats,
		"promotion_settings":          promotionSettings,
		"available_promotion_classes": availableClasses,
	})
}

func UpdateClassAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type UpdateClassRequest struct {
		Name      string `json:"name"`
		TeacherID string `json:"teacher_id"`
	}

	var req UpdateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class name is required"})
	}

	// Check if class exists
	existingClass, err := GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Class not found"})
	}

	// Update class data
	existingClass.Name = req.Name
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
