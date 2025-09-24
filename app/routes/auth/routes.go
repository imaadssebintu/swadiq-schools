package auth

import (
	"database/sql"
	"strings"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// Public routes
	auth.Get("/login", ShowLoginPage)
	auth.Post("/login", LoginAPI)
	auth.Post("/logout", LogoutAPI)
	auth.Get("/forgot-password", ShowForgotPasswordPage)
	auth.Post("/forgot-password", ForgotPasswordAPI)

	// Protected routes
	auth.Use(AuthMiddleware)
	auth.Get("/profile", ShowProfilePage)
	auth.Post("/change-password", ChangePasswordAPI)
}

func ShowLoginPage(c *fiber.Ctx) error {
	// Check if already logged in
	if sessionID := c.Cookies("session_id"); sessionID != "" {
		if _, err := database.GetSessionByID(config.GetDB(), sessionID); err == nil {
			return c.Redirect("/dashboard")
		}
	}

	return c.Render("auth/login", fiber.Map{
		"Title": "Login - Swadiq Schools",
	}, "")
}

func ShowForgotPasswordPage(c *fiber.Ctx) error {
	return c.Render("auth/forgot-password", fiber.Map{
		"Title": "Forgot Password - Swadiq Schools",
	}, "")
}

func ShowProfilePage(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	return c.Render("auth/profile", fiber.Map{
		"Title":       "Profile - Swadiq Schools",
		"CurrentPage": "profile",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
		"Role":        user.Roles[0].Name, // Assuming the first role is the primary one
	})
}

// AuthMiddleware validates session and sets user context
func AuthMiddleware(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")

	// Check if this is an API request
	isAPIRequest := strings.HasPrefix(c.Path(), "/api/")

	if sessionID == "" {
		if isAPIRequest {
			return c.Status(401).JSON(fiber.Map{"error": "No session found"})
		}
		// For web pages, redirect to login
		return c.Redirect("/auth/login")
	}

	session, err := database.GetSessionByID(config.GetDB(), sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			if isAPIRequest {
				return c.Status(401).JSON(fiber.Map{"error": "Invalid session"})
			}
			// For web pages, redirect to login
			return c.Redirect("/auth/login")
		}
		if isAPIRequest {
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		// For web pages, show error page
		return c.Status(500).Render("error", fiber.Map{
			"Title":        "Error - Swadiq Schools",
			"ErrorCode":    "500",
			"ErrorTitle":   "Database Error",
			"ErrorMessage": "Unable to verify your session. Please try logging in again.",
		})
	}

	// Get user details by ID
	user, err := database.GetUserByID(config.GetDB(), session.UserID)
	if err != nil {
		if isAPIRequest {
			return c.Status(500).JSON(fiber.Map{"error": "User not found"})
		}
		// For web pages, redirect to login
		return c.Redirect("/auth/login")
	}

	// Get user roles
	roles, err := database.GetUserRoles(config.GetDB(), session.UserID)
	if err != nil {
		if isAPIRequest {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get user roles"})
		}
		// For web pages, show error page
		return c.Status(500).Render("error", fiber.Map{
			"Title":        "Error - Swadiq Schools",
			"ErrorCode":    "500",
			"ErrorTitle":   "Authorization Error",
			"ErrorMessage": "Unable to load user permissions. Please try logging in again.",
		})
	}

	// Set user context
	c.Locals("user_id", user.ID)
	c.Locals("user_email", user.Email)
	c.Locals("user_first_name", user.FirstName)
	c.Locals("user_last_name", user.LastName)
	c.Locals("user_roles", roles)
	c.Locals("user", user)

	return c.Next()
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles := c.Locals("user_roles").([]*models.Role)

		for _, userRole := range userRoles {
			for _, allowedRole := range allowedRoles {
				if userRole.Name == allowedRole {
					return c.Next()
				}
			}
		}

		// Check if this is an API request
		isAPIRequest := strings.HasPrefix(c.Path(), "/api/")

		if isAPIRequest {
			return c.Status(403).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		// For web pages, show 403 error page
		return c.Status(403).Render("error", fiber.Map{
			"Title":        "Access Forbidden - Swadiq Schools",
			"CurrentPage":  "",
			"ErrorCode":    "403",
			"ErrorTitle":   "Access Forbidden",
			"ErrorMessage": "You don't have permission to access this resource.",
			"user":         c.Locals("user"),
		})
	}
}
