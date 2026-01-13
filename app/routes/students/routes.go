package students

import (
	"fmt"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupStudentsRoutes(app *fiber.App) {
	students := app.Group("/students")
	students.Use(auth.AuthMiddleware)

	// Routes
	students.Get("/", StudentsPage)
	students.Get("/:id", StudentViewPage)

	// API routes
	api := app.Group("/api/students")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetStudentsAPI)             // Get all students
	api.Get("/search", SearchStudentsAPI)    // Search students
	api.Get("/stats", GetStudentsStatsAPI)   // Get students statistics
	api.Get("/table", GetStudentsTableAPI)   // Get students formatted for table
	api.Get("/year", GetStudentsByYearAPI)   // Get students by year (?year=2025)
	api.Get("/class", GetStudentsByClassAPI) // Get students by class (?class_id=uuid)
	api.Get("/:id", GetStudentByIDAPI)       // Get single student by ID
	api.Post("/", CreateStudentAPI)          // Create new student
	api.Put("/:id", UpdateStudentAPI)        // Update existing student
	api.Delete("/:id", DeleteStudentAPI)     // Delete student

	// Parent selection API
	app.Get("/api/parents", GetParentsAPI) // Get parents for selection
}

func StudentsPage(c *fiber.Ctx) error {
	students, err := database.GetAllStudents(config.GetDB())
	if err != nil {
		return c.Status(500).Render("error", fiber.Map{
			"Title":        "Error - Swadiq Schools",
			"CurrentPage":  "students",
			"ErrorCode":    "500",
			"ErrorTitle":   "Database Error",
			"ErrorMessage": "Failed to load students. Please try again later.",
			"ShowRetry":    true,
			"user":         c.Locals("user"),
		})
	}

	user := c.Locals("user").(*models.User)
	return c.Render("students/index", fiber.Map{
		"Title":       "Students - Swadiq Schools",
		"CurrentPage": "students",
		"students":    students,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}

func StudentViewPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	studentID := c.Params("id")

	// Get student details to show name in title if possible
	student, _ := database.GetStudentByID(config.GetDB(), studentID)

	title := "Student Profile - Swadiq Schools"
	if student != nil {
		title = fmt.Sprintf("%s %s - Profile", student.FirstName, student.LastName)
	}

	return c.Render("students/view", fiber.Map{
		"Title":       title,
		"CurrentPage": "students",
		"studentID":   studentID,
		"student":     student,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
