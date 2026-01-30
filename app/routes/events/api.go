package events

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetEventsAPI returns a list of events
func GetEventsAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	events, err := database.GetEvents(db, false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch events",
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"events":  events,
	})
}

// CreateEventAPI creates a new event
func CreateEventAPI(c *fiber.Ctx) error {
	event := new(models.Event)
	if err := c.BodyParser(event); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	db := config.GetDB()

	// Automatically assign current term ID if not provided
	if event.TermID == "" {
		if term, err := database.GetCurrentTerm(db); err == nil {
			event.TermID = term.ID
		}
	}

	if err := database.CreateEvent(db, event); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create event",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"event":   event,
	})
}

// UpdateEventAPI updates an existing event
func UpdateEventAPI(c *fiber.Ctx) error {
	id := c.Params("id")

	event := new(models.Event)
	if err := c.BodyParser(event); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	event.ID = id

	db := config.GetDB()

	// Automatically assign current term ID if not provided
	if event.TermID == "" {
		if term, err := database.GetCurrentTerm(db); err == nil {
			event.TermID = term.ID
		}
	}

	if err := database.UpdateEvent(db, event); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update event",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event updated successfully",
	})
}

// DeleteEventAPI deletes an event
func DeleteEventAPI(c *fiber.Ctx) error {
	id := c.Params("id")

	db := config.GetDB()
	if err := database.DeleteEvent(db, id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete event",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event deleted successfully",
	})
}

// GetEventCategoriesAPI returns all event categories
func GetEventCategoriesAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	categories, err := database.GetEventCategories(db)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch categories",
		})
	}
	return c.JSON(fiber.Map{
		"success":    true,
		"categories": categories,
	})
}

// CreateEventCategoryAPI creates a new event category
func CreateEventCategoryAPI(c *fiber.Ctx) error {
	category := new(models.EventCategory)
	if err := c.BodyParser(category); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	db := config.GetDB()
	if err := database.CreateEventCategory(db, category); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create category",
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"category": category,
	})
}

// UpdateEventCategoryAPI updates an existing event category
func UpdateEventCategoryAPI(c *fiber.Ctx) error {
	id := c.Params("id")

	category := new(models.EventCategory)
	if err := c.BodyParser(category); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	category.ID = id

	db := config.GetDB()
	if err := database.UpdateEventCategory(db, category); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update category",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Category updated successfully",
	})
}

// DeleteEventCategoryAPI deletes an event category
func DeleteEventCategoryAPI(c *fiber.Ctx) error {
	id := c.Params("id")

	db := config.GetDB()
	if err := database.DeleteEventCategory(db, id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete category",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Category deleted successfully",
	})
}
