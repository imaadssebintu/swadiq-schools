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
	// Page routes
	app.Get("/events", auth.AuthMiddleware, renderEventsPage)

	// API routes
	api := app.Group("/api/events")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetEventsAPI)
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
		"HasEvents":      len(events) > 0,
		"ErrorMessage":   errorMsg,
		"CategoryCounts": categoryCounts,
	})
}
