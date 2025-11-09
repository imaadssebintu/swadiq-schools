package auth

import (
	"strings"
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
	if tokenString := c.Cookies("jwt_token"); tokenString != "" {
		if _, err := ValidateJWT(tokenString); err == nil {
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
	userRoles := c.Locals("user_roles").([]*models.Role)
	
	// Handle case where user has no roles
	roleName := ""
	if len(userRoles) > 0 {
		roleName = userRoles[0].Name
	}
	
	return c.Render("auth/profile", fiber.Map{
		"Title":       "Profile - Swadiq Schools",
		"CurrentPage": "profile",
		"user":        user,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
		"Role":        roleName,
	})
}

// AuthMiddleware validates JWT and sets user context
func AuthMiddleware(c *fiber.Ctx) error {
	// Get JWT token from cookie or Authorization header
	var tokenString string
	
	// First try cookie
	tokenString = c.Cookies("jwt_token")
	
	// If no cookie, try Authorization header
	if tokenString == "" {
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenString = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check if this is an API request
	isAPIRequest := strings.HasPrefix(c.Path(), "/api/")

	if tokenString == "" {
		if isAPIRequest {
			return c.Status(401).JSON(fiber.Map{"error": "No token found"})
		}
		// For web pages, redirect to login
		return c.Redirect("/auth/login")
	}

	// Validate JWT token
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		if isAPIRequest {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}
		// For web pages, redirect to login
		return c.Redirect("/auth/login")
	}

	// Create user object from claims
	user := &models.User{
		ID:        claims.UserID,
		Email:     claims.Email,
		FirstName: claims.FirstName,
		LastName:  claims.LastName,
		IsActive:  true,
	}

	// Convert role names to role objects
	roles := make([]*models.Role, len(claims.Roles))
	for i, roleName := range claims.Roles {
		roles[i] = &models.Role{Name: roleName}
	}
	user.Roles = roles

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
