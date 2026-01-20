package events

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupEventsRoutes sets up events routes
func SetupEventsRoutes(app *fiber.App) {
	db := config.GetDB()
	if err := database.InitEventDatabase(db); err != nil {
		fmt.Printf("Warning: Failed to initialize event database: %v\n", err)
	}

	// Page routes
	app.Get("/events", auth.AuthMiddleware, renderEventsPage)

	// API routes
	api := app.Group("/api/events")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetEventsAPI)
	api.Get("/categories", GetEventCategoriesAPI)
	api.Post("/categories", CreateEventCategoryAPI)
	api.Put("/categories/:id", UpdateEventCategoryAPI)
	api.Delete("/categories/:id", DeleteEventCategoryAPI)
	api.Post("/", CreateEventAPI)
	api.Put("/:id", UpdateEventAPI)
	api.Delete("/:id", DeleteEventAPI)
}

type EventGroup struct {
	Month  string
	Events []models.Event
}

func renderEventsPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	db := config.GetDB()
	events, err := database.GetEvents(db)

	errorMsg := ""
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		errorMsg = err.Error()
	}

	// Group events by Month Year
	var eventGroups []EventGroup
	if len(events) > 0 {
		currentMonth := ""
		var currentGroup *EventGroup

		for _, event := range events {
			monthYear := event.StartDate.Format("January 2006")
			if monthYear != currentMonth {
				if currentGroup != nil {
					eventGroups = append(eventGroups, *currentGroup)
				}
				currentMonth = monthYear
				currentGroup = &EventGroup{
					Month:  monthYear,
					Events: []models.Event{event},
				}
			} else {
				currentGroup.Events = append(currentGroup.Events, event)
			}
		}
		if currentGroup != nil {
			eventGroups = append(eventGroups, *currentGroup)
		}
	}

	// Get categories for the sidebar labels/filter
	categories, _ := database.GetEventCategories(db)

	// Get category counts
	categoryCounts, _ := database.GetEventCategoryCounts(db)

	return c.Render("events/index", fiber.Map{
		"Title":          "Events - Swadiq Schools",
		"CurrentPage":    "events",
		"FirstName":      user.FirstName,
		"LastName":       user.LastName,
		"Email":          user.Email,
		"User":           user,
		"EventGroups":    eventGroups,
		"Events":         events,
		"Categories":     categories,
		"HasEvents":      len(events) > 0,
		"ErrorMessage":   errorMsg,
		"CategoryCounts": categoryCounts,
	})
}
