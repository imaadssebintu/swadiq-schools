package parents

import (
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupParentsRoutes(app *fiber.App) {
	api := app.Group("/api/parents")
	api.Use(auth.AuthMiddleware)

	api.Get("/", GetParentsAPI)
	api.Post("/", CreateParentAPI)
	api.Get("/search", SearchParentsAPI)
}
