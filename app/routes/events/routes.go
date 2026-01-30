package events

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"time"

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
	// Get ALL events for the calendar widget
	allEvents, _ := database.GetEvents(db, false)

	// Determine current time for filtering
	now := time.Now()

	// Filter and group ONLY upcoming events for the main list
	var eventGroups []EventGroup
	var upcomingEvents []models.Event

	for _, event := range allEvents {
		// Event is upcoming if its end date is now or in the future
		if event.EndDate.After(now) || event.EndDate.Equal(now) {
			upcomingEvents = append(upcomingEvents, event)
		}
	}

	if len(upcomingEvents) > 0 {
		currentMonth := ""
		var currentGroup *EventGroup

		for _, event := range upcomingEvents {
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

	// Get category counts - show count for upcoming events only to match the list
	categoryCounts, _ := database.GetEventCategoryCounts(db, true)

	return c.Render("events/index", fiber.Map{
		"Title":          "Events - Swadiq Schools",
		"CurrentPage":    "events",
		"FirstName":      user.FirstName,
		"LastName":       user.LastName,
		"Email":          user.Email,
		"User":           user,
		"EventGroups":    eventGroups,
		"Events":         allEvents, // Pass all events for the calendar
		"HasEvents":      len(upcomingEvents) > 0,
		"Categories":     categories,
		"CategoryCounts": categoryCounts,
	})
}
