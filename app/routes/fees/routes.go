package fees

import (
	"swadiq-schools/app/config"
	"swadiq-schools/app/routes/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupFeesRoutes sets up the fees routes
func SetupFeesRoutes(app *fiber.App) {
	// Group for fees routes with authentication middleware
	fees := app.Group("/fees")
	fees.Use(auth.AuthMiddleware)

	// API routes for fees
	feesAPI := app.Group("/api/fees")
	feesAPI.Use(auth.AuthMiddleware)
	
	// Add apply fees route
	feesAPI.Post("/apply", func(c *fiber.Ctx) error {
		return ApplyFeesAPI(c, config.GetDB())
	})

	// Web routes
	fees.Get("/", func(c *fiber.Ctx) error {
		return c.Render("fees/index", fiber.Map{
			"Title":       "Fees Management - Swadiq Schools",
			"CurrentPage": "fees",
		})
	})

	// API routes
	feesAPI.Get("/", func(c *fiber.Ctx) error {
		return GetFeesAPI(c, config.GetDB())
	})

	feesAPI.Get("/:id", func(c *fiber.Ctx) error {
		return GetFeeByIDAPI(c, config.GetDB())
	})

	feesAPI.Post("/", func(c *fiber.Ctx) error {
		return CreateFeeAPI(c, config.GetDB())
	})

	feesAPI.Put("/:id", func(c *fiber.Ctx) error {
		return UpdateFeeAPI(c, config.GetDB())
	})

	feesAPI.Delete("/:id", func(c *fiber.Ctx) error {
		return DeleteFeeAPI(c, config.GetDB())
	})

	feesAPI.Post("/:id/pay", func(c *fiber.Ctx) error {
		return MarkFeeAsPaidAPI(c, config.GetDB())
	})

	feesAPI.Get("/stats", func(c *fiber.Ctx) error {
		return GetFeeStatsAPI(c, config.GetDB())
	})

	// Fee Types routes
	feeTypes := app.Group("/fee-types")
	feeTypes.Use(auth.AuthMiddleware)

	feeTypesAPI := app.Group("/api/fee-types")
	feeTypesAPI.Use(auth.AuthMiddleware)

	// Fee Types web route
	feeTypes.Get("/", func(c *fiber.Ctx) error {
		return c.Render("fees/fee_types", fiber.Map{
			"Title":       "Fee Types - Swadiq Schools",
			"CurrentPage": "fees",
		})
	})

	// Active Fees web route
	fees.Get("/active", func(c *fiber.Ctx) error {
		return c.Render("fees/active_fees", fiber.Map{
			"Title":       "Active Fees - Swadiq Schools",
			"CurrentPage": "fees",
		})
	})

	// Fee Types API routes
	feeTypesAPI.Get("/", func(c *fiber.Ctx) error {
		return GetFeeTypesAPI(c, config.GetDB())
	})

	feeTypesAPI.Get("/:id", func(c *fiber.Ctx) error {
		return GetFeeTypeAPI(c, config.GetDB())
	})

	feeTypesAPI.Post("/", func(c *fiber.Ctx) error {
		return CreateFeeTypeAPI(c, config.GetDB())
	})

	feeTypesAPI.Put("/:id", func(c *fiber.Ctx) error {
		return UpdateFeeTypeAPI(c, config.GetDB())
	})

	feeTypesAPI.Delete("/:id", func(c *fiber.Ctx) error {
		return DeleteFeeTypeAPI(c, config.GetDB())
	})

	feeTypesAPI.Get("/:id/assignments", func(c *fiber.Ctx) error {
		return GetFeeTypeAssignmentsAPI(c, config.GetDB())
	})

	// Students by classes API route
	feesAPI.Get("/students-by-classes", func(c *fiber.Ctx) error {
		return GetStudentsForClassesAPI(c, config.GetDB())
	})

	// Fee activation routes
	feesAPI.Post("/activate", func(c *fiber.Ctx) error {
		return ActivateFeesAPI(c, config.GetDB())
	})

	feesAPI.Get("/active", func(c *fiber.Ctx) error {
		return GetActiveFeeTypesAPI(c, config.GetDB())
	})
}
