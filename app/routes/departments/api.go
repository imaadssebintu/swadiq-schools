package departments

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetDepartmentsAPI(c *fiber.Ctx) error {
	departments, err := database.GetAllDepartments(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch departments"})
	}

	return c.JSON(fiber.Map{
		"departments": departments,
		"count":       len(departments),
	})
}

func CreateDepartmentAPI(c *fiber.Ctx) error {
	type CreateDepartmentRequest struct {
		Name                string `json:"name"`
		Code                string `json:"code"`
		HeadOfDepartmentID *string `json:"head_of_department_id"`
	}

	var req CreateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" || req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and code are required"})
	}

	department := &models.Department{
		Name: req.Name,
		Code: req.Code,
		HeadOfDepartmentID: req.HeadOfDepartmentID,
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
		Name                string `json:"name"`
		Code                string `json:"code"`
		HeadOfDepartmentID *string `json:"head_of_department_id"`
	}

	var req UpdateDepartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" || req.Code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name and code are required"})
	}

	department := &models.Department{
		ID:   departmentID,
		Name: req.Name,
		Code: req.Code,
		HeadOfDepartmentID: req.HeadOfDepartmentID,
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
