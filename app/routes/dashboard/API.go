package dashboard

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

// GetDashboard handles dashboard page
func GetDashboard(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	user := c.Locals("user").(*models.User)

	// Fetch upcoming events
	db := config.GetDB()
	var upcomingEvents []models.Event

	// Fetch upcoming events only (Efficiency: GetEvents(db, true))
	if allEvents, err := database.GetEvents(db, true); err == nil {
		count := 0
		for _, e := range allEvents {
			upcomingEvents = append(upcomingEvents, e)
			count++
			if count >= 2 {
				break
			}
		}
	} else {
		// Log error if any (can't see stdout, but good practice)
		// fmt.Println("Dashboard Event Fetch Error:", err)
	}

	c.Locals("Title", "Dashboard")
	return c.Render("dashboard/index", fiber.Map{
		"Title":       "Dashboard",
		"CurrentPage": "dashboard",
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
		"user":        user,
		"Events":      upcomingEvents,
	})
}

// GetDashboardStatsAPI returns dashboard statistics as JSON
func GetDashboardStatsAPI(c *fiber.Ctx) error {
	// Get database connection
	db := config.GetDB()

	// Get dashboard statistics
	stats, err := database.GetDashboardStats(db)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch dashboard statistics",
			"details": err.Error(),
		})
	}

	// Return statistics as JSON
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}
