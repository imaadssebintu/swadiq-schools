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
	teachers.Get("/:id", TeacherViewPage)

	// API routes
	api := app.Group("/api/teachers")
	api.Use(auth.AuthMiddleware)
	api.Get("/", GetTeachersAPI)
	api.Get("/selection", GetTeachersForSelectionAPI) // Fast endpoint for selection
	api.Get("/for-timetable", GetTeachersForTimetableAPI)
	api.Get("/all-for-paper", GetAllTeachersForPaperAPI)
	api.Get("/counts", GetTeacherCountsAPI)
	api.Get("/stats", GetTeacherStatsAPI)
	api.Get("/search", SearchTeachersAPI)
	api.Get("/check-phone", CheckPhoneUniquenessAPI)
	api.Get("/department-overview", GetDepartmentOverviewAPI)
	api.Post("/", CreateTeacherAPI)
	api.Get("/:id", GetTeacherAPI)
	api.Put("/:id", UpdateTeacherAPI)
	api.Delete("/:id", DeleteTeacherAPI)
	api.Get("/:id/subjects", GetTeacherSubjectsAPI)
	api.Post("/:id/subjects", AssignTeacherSubjectsAPI)
	api.Delete("/:id/subjects/:subjectId", RemoveTeacherSubjectAPI)
	api.Get("/:id/availability", GetTeacherAvailabilityAPI)
	api.Post("/:id/availability", UpdateTeacherAvailabilityAPI)
	api.Get("/:id/salary", GetTeacherSalaryAPI)
	api.Post("/:id/salary", SetTeacherSalaryAPI)
	api.Get("/:id/ledger", GetTeacherLedgerAPI)
	api.Get("/:id/ledger/base-salary", GetTeacherBaseSalaryLedgerAPI)
	api.Get("/:id/ledger/allowance", GetTeacherAllowanceLedgerAPI)
	api.Post("/:id/pay", PayTeacherAPI)
	api.Get("/:id/payments", GetTeacherPaymentsAPI)
}

func TeachersPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("teachers/index", fiber.Map{
		"Title":       "Teachers - Swadiq Schools",
		"CurrentPage": "teachers",
		"teachers":    []*models.User{}, // Empty array
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}

func TeacherViewPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	teacherID := c.Params("id")
	db := config.GetDB()

	// Fetch teacher data for initial render
	teacher, err := database.GetTeacherByID(db, teacherID)
	if err != nil {
		return c.Redirect("/teachers")
	}

	// Fetch teacher roles
	var roles []models.Role
	roleQuery := `SELECT r.id, r.name FROM roles r 
				  INNER JOIN user_roles ur ON r.id = ur.role_id 
				  WHERE ur.user_id = $1`
	rows, err := db.Query(roleQuery, teacherID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r models.Role
			if err := rows.Scan(&r.ID, &r.Name); err == nil {
				roles = append(roles, r)
			}
		}
	}

	return c.Render("teachers/view", fiber.Map{
		"Title":       "Teacher Details - Swadiq Schools",
		"CurrentPage": "teachers",
		"teacherID":   teacherID,
		"teacher":     teacher,
		"roles":       roles,
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
	})
}
