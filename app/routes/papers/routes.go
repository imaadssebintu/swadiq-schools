package papers

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupPapersRoutes(app *fiber.App) {
	papers := app.Group("/papers")
	papers.Use(auth.AuthMiddleware)

	// Routes
	papers.Get("/", PapersPage)

	// API routes are already defined in api.go
}

func PapersPage(c *fiber.Ctx) error {
	papers, err := database.GetAllPapers(config.GetDB())
	if err != nil {
		// Log the error for debugging
		println("Error getting papers:", err.Error())
		// Initialize empty slice if there's an error
		papers = []*models.Paper{}
	}

	// Ensure papers is never nil
	if papers == nil {
		papers = []*models.Paper{}
	}

	// Get subjects for count
	subjects, err := database.GetAllSubjects(config.GetDB())
	if err != nil {
		// Initialize empty slice if there's an error
		subjects = []*models.Subject{}
	}

	// Ensure subjects is never nil
	if subjects == nil {
		subjects = []*models.Subject{}
	}

	// Get teachers for count
	teachers, err := database.GetAllTeachers(config.GetDB())
	if err != nil {
		// Initialize empty slice if there's an error
		teachers = []*models.User{}
	}

	// Ensure teachers is never nil
	if teachers == nil {
		teachers = []*models.User{}
	}

	user := c.Locals("user").(*models.User)
	return c.Render("papers/index", fiber.Map{
		"Title":         "Papers Management - Swadiq Schools",
		"CurrentPage":   "papers",
		"papers":        papers,
		"subjectsCount": len(subjects),
		"teachersCount": len(teachers),
		"user":          user,
		"FirstName":     user.FirstName,
		"LastName":      user.LastName,
		"Email":         user.Email,
	})
}