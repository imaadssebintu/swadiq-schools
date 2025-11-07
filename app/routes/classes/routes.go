package classes

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupClassesRoutes(app *fiber.App) {
	classes := app.Group("/classes")
	classes.Use(auth.AuthMiddleware)

	// Routes
	classes.Get("/", ClassesPage)
	classes.Get("/:id", ClassDetailPage) // Individual class detail page

	// API routes (these were already set up in main.go, but let's make them explicit here too)
	api := app.Group("/api/classes")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetClassesAPI)
	api.Get("/stats", GetClassesStatsAPI) // Get classes statistics
	api.Get("/table", GetClassesTableAPI) // Get classes formatted for table
	api.Post("/", CreateClassAPI)
	api.Post("/:id/students", AddStudentToClassAPI) // Add student to class
	api.Post("/:id/papers/:paperId/teacher", AssignTeacherToPaperAPI) // Assign teacher to specific paper
}

func ClassesPage(c *fiber.Ctx) error {
	classes, err := database.GetAllClasses(config.GetDB())
	if err != nil {
		// Log the error for debugging
		println("Error getting classes:", err.Error())
		// Initialize empty slice if there's an error
		classes = []*models.Class{}
	}

	// Ensure classes is never nil
	if classes == nil {
		classes = []*models.Class{}
	}

	user := c.Locals("user").(*models.User)
	return c.Render("classes/index", fiber.Map{
		"Title":       "Classes Management - Swadiq Schools",
		"CurrentPage": "classes",
		"classes":     classes,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}

// ClassDetailPage renders the individual class detail page
func ClassDetailPage(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).SendString("Class ID is required")
	}

	// Get class basic info
	class, err := GetClassByID(config.GetDB(), classID)
	if err != nil {
		return c.Status(404).SendString("Class not found")
	}

	user := c.Locals("user").(*models.User)
	return c.Render("classes/detail", fiber.Map{
		"Title":       fmt.Sprintf("%s - Class Details", class.Name),
		"CurrentPage": "classes",
		"class":       class,
		"classID":     classID,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
