package academic

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers the academic year and term routes (Legacy - routes moved to settings)
func RegisterRoutes(app *fiber.App, db *sql.DB) {
	// All routes now live in settings/routes.go under /api/settings
	SetupAcademicRoutes(app)
}
