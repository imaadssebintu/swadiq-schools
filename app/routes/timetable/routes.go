package timetable

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"
	"swadiq-schools/app/routes/classes"

	"github.com/gofiber/fiber/v2"
)

func SetupTimetableRoutes(app *fiber.App) {
	timetable := app.Group("/timetable")
	timetable.Use(auth.AuthMiddleware)

	// Routes
	timetable.Get("/", TimetableIndexPage)
	timetable.Get("/class/:id", ClassTimetablePage)

	// API routes
	api := app.Group("/api/timetable")
	api.Use(auth.AuthMiddleware)
	api.Get("/class/:id", GetTimetableDataAPI)
	api.Post("/class/:id", SaveTimetableAPI)
	api.Get("/settings/class/:classId", GetTimetableSettingsAPI)
	api.Post("/settings/class/:classId", SaveTimetableSettingsAPI)
	api.Get("/settings/default", GetDefaultTimetableSettingsAPI)
	api.Post("/settings/default", SaveDefaultTimetableSettingsAPI)
	api.Post("/settings/apply-default", ApplyDefaultSettingsAPI)
}

func TimetableIndexPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("timetable/index", fiber.Map{
		"Title":       "Timetable Management - Swadiq Schools",
		"CurrentPage": "timetable",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}

// ClassTimetablePage renders the class timetable page
func ClassTimetablePage(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).SendString("Class ID is required")
	}

	// Get class basic info
	class, err := classes.GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).SendString("Class not found")
	}

	user := c.Locals("user").(*models.User)
	return c.Render("timetable/class", fiber.Map{
		"Title":       fmt.Sprintf("%s - Timetable", class.Name),
		"CurrentPage": "timetable",
		"class":       class,
		"classID":     classID,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
