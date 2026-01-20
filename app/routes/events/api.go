package events

import (
	"strconv"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetEventsAPI returns a list of events
func GetEventsAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	events, err := database.GetEvents(db)
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
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid event ID",
		})
	}

	event := new(models.Event)
	if err := c.BodyParser(event); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	event.ID = id

	db := config.GetDB()
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
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid event ID",
		})
	}

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
