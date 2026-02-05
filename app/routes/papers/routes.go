package papers

import (
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupPapersRoutes(app *fiber.App) {
	papers := app.Group("/subjects/papers")
	papers.Use(auth.AuthMiddleware)

	// HTML Routes
	papers.Get("/", PapersPage)

	// API Routes
	api := app.Group("/api/subjects/papers")
	api.Get("/", GetPapersAPI)
	api.Get("/table", GetPapersTableAPI)
	api.Get("/stats", GetPapersStatsAPI)
	api.Get("/subject/:subjectId", GetPapersBySubjectAPI)
	api.Get("/by-subject/:subjectId", GetPapersBySubjectAPI)
	api.Get("/:id", GetPaperAPI)
	api.Post("/", CreatePaperAPI)
	api.Put("/:id", UpdatePaperAPI)
	api.Delete("/:id", DeletePaperAPI)
	api.Get("/weights", GetPaperWeightsAPI)
	api.Post("/weights", SavePaperWeightsAPI)
}

func PapersPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("papers/index", fiber.Map{
		"Title":       "Papers Management - Swadiq Schools",
		"CurrentPage": "subjects",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
