package teachers

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetTeachersAPI(c *fiber.Ctx) error {
	teachers, err := database.GetAllTeachers(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch teachers"})
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
		"count":    len(teachers),
	})
}

func CreateTeacherAPI(c *fiber.Ctx) error {
	type CreateTeacherRequest struct {
		FirstName    string   `json:"first_name"`
		LastName     string   `json:"last_name"`
		Email        string   `json:"email"`
		Password     string   `json:"password"`
		Phone        string   `json:"phone"`
		DepartmentID string   `json:"department_id"`
		SubjectIDs   []string `json:"subject_ids"`
	}

	var req CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name, last name, email, and password are required"})
	}

	// Create user account for teacher
	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password, // This will be hashed in the database function
		Phone:     req.Phone,
	}

	if err := database.CreateTeacher(config.GetDB(), user); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create teacher",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":       "Teacher created successfully",
		"teacher":       user,
		"department_id": req.DepartmentID,
		"subject_ids":   req.SubjectIDs,
	})
}


func SearchTeachersAPI(c *fiber.Ctx) error {
	query := c.Query("q", "")
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	teachers, total, err := SearchTeachersWithPagination(config.GetDB(), query, limit, offset)
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

	teacher, err := GetTeacherByID(config.GetDB(), teacherID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Teacher not found"})
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
		Phone     string `json:"phone"`
	}

	var req UpdateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name, last name, and email are required"})
	}

	// Check if teacher exists
	existingTeacher, err := GetTeacherByID(config.GetDB(), teacherID)
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
	_, err := GetTeacherByID(config.GetDB(), teacherID)
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

func GetSubjectsAPI(c *fiber.Ctx) error {
	departmentID := c.Query("department_id")

	var subjects []*models.Subject
	var err error

	if departmentID != "" {
		subjects, err = database.GetSubjectsByDepartment(config.GetDB(), departmentID)
	} else {
		subjects, err = database.GetAllSubjects(config.GetDB())
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch subjects"})
	}

	return c.JSON(fiber.Map{
		"subjects": subjects,
		"count":    len(subjects),
	})
}
