package teachers

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupTeachersRoutes(app *fiber.App) {
	teachers := app.Group("/teachers")
	teachers.Use(auth.AuthMiddleware)

	// Routes
	teachers.Get("/", TeachersPage)

	// API routes
	api := app.Group("/api/teachers")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetTeachersAPI)
	api.Get("/search", SearchTeachersAPI)
	api.Post("/", CreateTeacherAPI)
	api.Get("/:id", GetTeacherAPI)
	api.Put("/:id", UpdateTeacherAPI)
	api.Delete("/:id", DeleteTeacherAPI)

	subjectsAPI := app.Group("/api/subjects")
	subjectsAPI.Use(auth.AuthMiddleware)
	subjectsAPI.Get("/", GetSubjectsAPI)
}

func TeachersPage(c *fiber.Ctx) error {
	teachers, err := database.GetAllTeachers(config.GetDB())
	if err != nil {
		// Log the error for debugging
		println("Error getting teachers:", err.Error())
		// Initialize empty slice if there's an error
		teachers = []*models.User{}
	}

	// Ensure teachers is never nil
	if teachers == nil {
		teachers = []*models.User{}
	}

	user := c.Locals("user").(*models.User)
	return c.Render("teachers/index", fiber.Map{
		"Title":       "Teachers - Swadiq Schools",
		"CurrentPage": "teachers",
		"teachers":    teachers,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}


