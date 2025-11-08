package departments

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetDepartmentsAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	query := `SELECT d.id, d.name, d.code, d.description, d.head_of_department_id, d.assistant_head_id, d.is_active, d.created_at, d.updated_at,
		h.first_name as head_first_name, h.last_name as head_last_name, h.email as head_email,
		a.first_name as assistant_first_name, a.last_name as assistant_last_name, a.email as assistant_email,
		COUNT(ud.user_id) as teacher_count
		FROM departments d
		LEFT JOIN users h ON d.head_of_department_id = h.id AND h.is_active = true
		LEFT JOIN users a ON d.assistant_head_id = a.id AND a.is_active = true
		LEFT JOIN user_departments ud ON d.id = ud.department_id
		LEFT JOIN users u ON ud.user_id = u.id AND u.is_active = true
		WHERE d.is_active = true
		GROUP BY d.id, d.name, d.code, d.description, d.head_of_department_id, d.assistant_head_id, d.is_active, d.created_at, d.updated_at, h.first_name, h.last_name, h.email, a.first_name, a.last_name, a.email
		ORDER BY d.name`

	rows, err := db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch departments"})
	}
	defer rows.Close()

	departments := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name, code string
		var description, headID, assistantID *string
		var isActive bool
		var createdAt, updatedAt string
		var headFirstName, headLastName, headEmail, assistantFirstName, assistantLastName, assistantEmail *string
		var teacherCount int

		if err := rows.Scan(&id, &name, &code, &description, &headID, &assistantID, &isActive, &createdAt, &updatedAt, &headFirstName, &headLastName, &headEmail, &assistantFirstName, &assistantLastName, &assistantEmail, &teacherCount); err == nil {
			dept := map[string]interface{}{
				"id":          id,
				"name":        name,
				"code":        code,
				"description": description,
				"is_active":   isActive,
				"created_at":  createdAt,
				"updated_at":  updatedAt,
				"teacher_count": teacherCount,
			}

			if headID != nil && headFirstName != nil && headLastName != nil {
				dept["head_of_department"] = map[string]interface{}{
					"id":         *headID,
					"first_name": *headFirstName,
					"last_name":  *headLastName,
					"email":      headEmail,
				}
			}

			if assistantID != nil && assistantFirstName != nil && assistantLastName != nil {
				dept["assistant_head"] = map[string]interface{}{
					"id":         *assistantID,
					"first_name": *assistantFirstName,
					"last_name":  *assistantLastName,
					"email":      assistantEmail,
				}
			}

			departments = append(departments, dept)
		}
	}

	return c.JSON(fiber.Map{
		"departments": departments,
		"count":       len(departments),
	})
}

func CreateDepartmentAPI(c *fiber.Ctx) error {
	type CreateDepartmentRequest struct {
		Name                string  `json:"name"`
		Code                string  `json:"code"`
		Description         *string `json:"description"`
		HeadOfDepartmentID  *string `json:"head_of_department_id"`
		AssistantHeadID     *string `json:"assistant_head_id"`
	}

	var req CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" || req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and code are required"})
	}

	department := &models.Department{
		Name:               req.Name,
		Code:               req.Code,
		Description:        req.Description,
		HeadOfDepartmentID: req.HeadOfDepartmentID,
		AssistantHeadID:    req.AssistantHeadID,
	}

	if err := database.CreateDepartment(config.GetDB(), department); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create department"})
	}

	return c.Status(201).JSON(department)
}

func UpdateDepartmentAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	if departmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID is required"})
	}

	type UpdateDepartmentRequest struct {
		Name                string  `json:"name"`
		Code                string  `json:"code"`
		Description         *string `json:"description"`
		HeadOfDepartmentID  *string `json:"head_of_department_id"`
		AssistantHeadID     *string `json:"assistant_head_id"`
	}

	var req UpdateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" || req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and code are required"})
	}

	department := &models.Department{
		ID:                 departmentID,
		Name:               req.Name,
		Code:               req.Code,
		Description:        req.Description,
		HeadOfDepartmentID: req.HeadOfDepartmentID,
		AssistantHeadID:    req.AssistantHeadID,
	}

	if err := database.UpdateDepartment(config.GetDB(), department); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update department"})
	}

	return c.Status(200).JSON(department)
}

func DeleteDepartmentAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	if departmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID is required"})
	}

	if err := database.DeleteDepartment(config.GetDB(), departmentID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete department"})
	}

	return c.SendStatus(204)
}

func GetDepartmentTeachersAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	if departmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID is required"})
	}

	db := config.GetDB()
	// Get department info first
	deptQuery := `SELECT head_of_department_id, assistant_head_id FROM departments WHERE id = $1`
	var headID, assistantID *string
	err := db.QueryRow(deptQuery, departmentID).Scan(&headID, &assistantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch department info"})
	}

	// Get all teachers in department
	query := `SELECT u.id, u.first_name, u.last_name, u.email, u.phone
		FROM users u
		INNER JOIN user_departments ud ON u.id = ud.user_id
		WHERE ud.department_id = $1 AND u.is_active = true
		ORDER BY u.first_name, u.last_name`

	rows, err := db.Query(query, departmentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch department teachers"})
	}
	defer rows.Close()

	teachers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, firstName, lastName, email string
		var phone *string
		if err := rows.Scan(&id, &firstName, &lastName, &email, &phone); err == nil {
			role := "Member"
			if headID != nil && *headID == id {
				role = "Head of Department"
			} else if assistantID != nil && *assistantID == id {
				role = "Assistant Head"
			}

			teachers = append(teachers, map[string]interface{}{
				"id":         id,
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"phone":      phone,
				"role":       role,
			})
		}
	}

	return c.JSON(fiber.Map{
		"teachers": teachers,
		"count":    len(teachers),
	})
}

func AddTeacherToDepartmentAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	if departmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID is required"})
	}

	type AddTeacherRequest struct {
		TeacherID string `json:"teacher_id"`
	}

	var req AddTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.TeacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID is required"})
	}

	db := config.GetDB()
	query := `INSERT INTO user_departments (user_id, department_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := db.Exec(query, req.TeacherID, departmentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add teacher to department"})
	}

	return c.JSON(fiber.Map{"message": "Teacher added successfully"})
}

func RemoveTeacherFromDepartmentAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	teacherID := c.Params("teacherId")
	if departmentID == "" || teacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID and Teacher ID are required"})
	}

	db := config.GetDB()
	query := `DELETE FROM user_departments WHERE user_id = $1 AND department_id = $2`
	_, err := db.Exec(query, teacherID, departmentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to remove teacher from department"})
	}

	return c.JSON(fiber.Map{"message": "Teacher removed successfully"})
}

func SetDepartmentLeadershipAPI(c *fiber.Ctx) error {
	departmentID := c.Params("id")
	if departmentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Department ID is required"})
	}

	type SetLeadershipRequest struct {
		TeacherID string `json:"teacher_id"`
		Role      string `json:"role"` // "head", "assistant", or "member"
	}

	var req SetLeadershipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.TeacherID == "" || (req.Role != "head" && req.Role != "assistant" && req.Role != "member") {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID and valid role (head/assistant/member) are required"})
	}

	db := config.GetDB()
	var query string
	var args []interface{}
	
	if req.Role == "head" {
		query = `UPDATE departments SET head_of_department_id = $1 WHERE id = $2`
		args = []interface{}{req.TeacherID, departmentID}
	} else if req.Role == "assistant" {
		query = `UPDATE departments SET assistant_head_id = $1 WHERE id = $2`
		args = []interface{}{req.TeacherID, departmentID}
	} else {
		// Remove from leadership positions
		query = `UPDATE departments SET head_of_department_id = CASE WHEN head_of_department_id = $1 THEN NULL ELSE head_of_department_id END, assistant_head_id = CASE WHEN assistant_head_id = $1 THEN NULL ELSE assistant_head_id END WHERE id = $2`
		args = []interface{}{req.TeacherID, departmentID}
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update department leadership"})
	}

	return c.JSON(fiber.Map{"message": "Leadership updated successfully"})
}



