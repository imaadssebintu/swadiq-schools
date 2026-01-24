package academic

import (
	"database/sql"
	"swadiq-schools/app/config"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// AcademicDashboardHandler serves the comprehensive performance dashboard
func AcademicDashboardHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		// Fetch summary metrics (Simulated for initial UI, should be calculated from results)
		// We can get real data here like average marks, total students, etc.

		data := fiber.Map{
			"Title":       "Academic Performance",
			"CurrentPage": "academic",
			"FirstName":   user.FirstName,
			"LastName":    user.LastName,
			"Email":       user.Email,
			"user":        user,
		}

		return c.Render("academic/index", data)
	}
}

// SetupAcademicRoutes registers the new performance-focused routes
func SetupAcademicRoutes(app *fiber.App) {
	db := config.GetDB()

	// Web Routes
	app.Get("/academic", auth.AuthMiddleware, AcademicDashboardHandler(db))
}
