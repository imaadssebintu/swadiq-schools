package parents

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func GetParentsAPI(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	parents, err := database.GetParentsForSelection(config.GetDB(), "", limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch parents"})
	}

	return c.JSON(fiber.Map{
		"parents": parents,
		"count":   len(parents),
	})
}

func CreateParentAPI(c *fiber.Ctx) error {
	type CreateParentRequest struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		Address   string `json:"address"`
	}

	var req CreateParentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name and last name are required"})
	}

	parent := &models.Parent{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if req.Phone != "" {
		parent.Phone = &req.Phone
	}
	if req.Email != "" {
		parent.Email = &req.Email
	}
	if req.Address != "" {
		parent.Address = &req.Address
	}

	if err := database.CreateParent(config.GetDB(), parent); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create parent"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Parent created successfully",
		"parent":  parent,
	})
}

// SearchParentsAPI handles searching for parents
func SearchParentsAPI(c *fiber.Ctx) error {
	query := c.Query("q", "")
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	db := config.GetDB()
	parents, err := database.GetParentsForSelection(db, query, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search parents",
		})
	}

	return c.JSON(fiber.Map{
		"parents": parents,
		"count":   len(parents),
	})
}
