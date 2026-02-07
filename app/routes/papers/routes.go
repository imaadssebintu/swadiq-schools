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
	papi := app.Group("/api/subjects/papers", auth.AuthMiddleware)
	papi.Get("/", GetPapersAPI)
	papi.Get("/table", GetPapersTableAPI)
	papi.Get("/stats", GetPapersStatsAPI)
	papi.Get("/subject/:subjectId", GetPapersBySubjectAPI)
	papi.Get("/by-subject/:subjectId", GetPapersBySubjectAPI)
	papi.Get("/weights", GetPaperWeightsAPI)
	papi.Post("/weights", SavePaperWeightsAPI)
	papi.Get("/:id", GetPaperAPI)
	papi.Post("/", CreatePaperAPI)
	papi.Put("/:id", UpdatePaperAPI)
	papi.Delete("/:id", DeletePaperAPI)
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
