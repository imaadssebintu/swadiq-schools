package expenses

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func ExpensesPageHandler(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("expenses/index", fiber.Map{
		"Title":       "Expenses Management",
		"CurrentPage": "expenses",
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"user":        user,
	})
}

func GetExpensesAPI(c *fiber.Ctx) error {
	expenses, err := GetAllExpenses(config.GetDB())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load expenses",
			"details": err.Error(),
		})
	}
	return c.JSON(expenses)
}

func GetCategoriesAPI(c *fiber.Ctx) error {
	categories, err := GetAllCategories(config.GetDB())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load categories",
			"details": err.Error(),
		})
	}
	return c.JSON(categories)
}

func CreateExpenseAPI(c *fiber.Ctx) error {
	var e models.Expense
	if err := c.BodyParser(&e); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := CreateExpense(config.GetDB(), &e); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create expense"})
	}

	return c.Status(fiber.StatusCreated).JSON(e)
}

func UpdateExpenseAPI(c *fiber.Ctx) error {
	id := c.Params("id")
	var e models.Expense
	if err := c.BodyParser(&e); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	e.ID = id
	if err := UpdateExpense(config.GetDB(), &e); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update expense"})
	}

	return c.JSON(e)
}

func DeleteExpenseAPI(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := DeleteExpense(config.GetDB(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete expense"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func CreateCategoryAPI(c *fiber.Ctx) error {
	var cat models.Category
	if err := c.BodyParser(&cat); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := CreateCategory(config.GetDB(), &cat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create category"})
	}

	return c.Status(fiber.StatusCreated).JSON(cat)
}

func UpdateCategoryAPI(c *fiber.Ctx) error {
	id := c.Params("id")
	var cat models.Category
	if err := c.BodyParser(&cat); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	cat.ID = id
	if err := UpdateCategory(config.GetDB(), &cat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update category"})
	}

	return c.JSON(cat)
}

func DeleteCategoryAPI(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := DeleteCategory(config.GetDB(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete category"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
