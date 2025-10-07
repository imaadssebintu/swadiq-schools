package main

import (
	"log"
	"swadiq-schools/app/config"
	"swadiq-schools/app/routes/academic"
	"swadiq-schools/app/routes/attendance"
	"swadiq-schools/app/routes/auth"
	"swadiq-schools/app/routes/classes"
	"swadiq-schools/app/routes/dashboard"
	"swadiq-schools/app/routes/departments"
	"swadiq-schools/app/routes/fees"
	"swadiq-schools/app/routes/papers"
	"swadiq-schools/app/routes/parents"
	"swadiq-schools/app/routes/students"
	"swadiq-schools/app/routes/subjects"
	"swadiq-schools/app/routes/teachers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

// customErrorHandler handles HTTP errors with custom templates
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Handle different error codes
	switch code {
	case 404:
		return c.Status(404).Render("404", fiber.Map{
			"Title":       "Page Not Found - Swadiq Schools",
			"CurrentPage": "",
		})
	case 403:
		return c.Status(403).Render("error", fiber.Map{
			"Title":        "Access Forbidden - Swadiq Schools",
			"CurrentPage":  "",
			"ErrorCode":    "403",
			"ErrorTitle":   "Access Forbidden",
			"ErrorMessage": "You don't have permission to access this resource.",
		})
	case 401:
		return c.Status(401).Render("error", fiber.Map{
			"Title":        "Unauthorized - Swadiq Schools",
			"CurrentPage":  "",
			"ErrorCode":    "401",
			"ErrorTitle":   "Unauthorized",
			"ErrorMessage": "Please log in to access this resource.",
		})
	case 500:
		return c.Status(500).Render("error", fiber.Map{
			"Title":        "Server Error - Swadiq Schools",
			"CurrentPage":  "",
			"ErrorCode":    "500",
			"ErrorTitle":   "Internal Server Error",
			"ErrorMessage": "We're experiencing technical difficulties. Please try again later.",
			"ShowRetry":    true,
		})
	default:
		return c.Status(code).Render("error", fiber.Map{
			"Title":        "Error - Swadiq Schools",
			"CurrentPage":  "",
			"ErrorCode":    code,
			"ErrorTitle":   "An Error Occurred",
			"ErrorMessage": err.Error(),
		})
	}
}

func main() {
	// Initialize database
	config.InitDB()
	defer config.GetDB().Close()

	// Initialize template engine
	engine := html.New("./app/templates", ".html")
	engine.Reload(false) // Disable template reloading
	engine.Debug(false)  // Disable debug mode

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views:             engine,
		ViewsLayout:       "layouts/main",
		PassLocalsToViews: true,
		ErrorHandler:      customErrorHandler,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Static files
	app.Static("/static", "./static")
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("./static/favicon.ico")
	})

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/auth/login")
	})

	// Setup auth routes
	auth.SetupAuthRoutes(app)

	// Setup dashboard routes
	dashboard.SetupDashboardRoutes(app)

	// Setup students routes
	students.SetupStudentsRoutes(app)

	// Setup teachers routes
	teachers.SetupTeachersRoutes(app)

	// Setup classes routes
	classes.SetupClassesRoutes(app)

	// Setup subjects routes
	subjects.SetupSubjectsRoutes(app)

	// Setup departments routes
	departments.SetupDepartmentsRoutes(app)

	// Setup attendance routes
	attendance.SetupAttendanceRoutes(app)

	// Setup fees routes
	fees.SetupFeesRoutes(app)

	// Setup papers routes
	papers.SetupPapersRoutes(app)

	// Setup academic routes
	academic.RegisterRoutes(app, config.GetDB())

	// Setup parents API routes
	api := app.Group("/api/parents")
	api.Use(auth.AuthMiddleware)
	api.Get("/", parents.GetParentsAPI)
	api.Post("/", parents.CreateParentAPI)
	api.Get("/search", parents.SearchParentsAPI)

	// Setup classes API routes
	classesAPI := app.Group("/api/classes")
	classesAPI.Use(auth.AuthMiddleware)
	classesAPI.Get("/", classes.GetClassesAPI)
	classesAPI.Post("/", classes.CreateClassAPI)

	// Setup roles API routes
	rolesAPI := app.Group("/api/roles")
	rolesAPI.Use(auth.AuthMiddleware)
	rolesAPI.Get("/", teachers.GetRolesAPI)
	rolesAPI.Post("/", teachers.CreateRoleAPI)

	// Setup papers API routes
	papersAPI := app.Group("/api/papers")
	papersAPI.Use(auth.AuthMiddleware)
	papersAPI.Get("/", papers.GetPapersAPI)
	papersAPI.Get("/:id", papers.GetPaperAPI)
	papersAPI.Post("/", papers.CreatePaperAPI)
	papersAPI.Put("/:id", papers.UpdatePaperAPI)
	papersAPI.Delete("/:id", papers.DeletePaperAPI)

	// Catch-all route for 404 errors (must be last)
	app.Use("*", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "Page not found")
	})

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(app.Listen(":8080"))
}
