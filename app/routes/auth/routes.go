package auth

import (
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
	auth.Get("/reset-password", ShowResetPasswordPage)
	auth.Post("/reset-password", ResetPasswordAPI)

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

func ShowResetPasswordPage(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).Render("error", fiber.Map{
			"Title":        "Invalid Reset Link - Swadiq Schools",
			"ErrorMessage": "Invalid reset link. Please request a new password reset.",
		})
	}

	return c.Render("auth/reset-password", fiber.Map{
		"Title": "Reset Password - Swadiq Schools",
		"Token": token,
	}, "")
}

func ShowProfilePage(c *fiber.Ctx) error {
	// First check if user_id exists in locals
	userIDLocal := c.Locals("user_id")
	if userIDLocal == nil {
		return c.Redirect("/auth/login")
	}
	userID := userIDLocal.(string)
	db := config.GetDB()

	// Fetch full user data from database to get phone, created_at etc.
	user, err := database.GetUserByID(db, userID)
	if err != nil || user == nil {
		return c.Status(500).Render("error", fiber.Map{
			"Title":        "Error - Swadiq Schools",
			"ErrorCode":    "500",
			"ErrorTitle":   "Database Error",
			"ErrorMessage": "Could not fetch user profile details. Please try logging in again.",
		})
	}

	// Fetch all roles
	roles, err := database.GetUserRoles(db, userID)
	if err != nil {
		roles = []*models.Role{}
	}

	// Fetch departments
	departments, err := database.GetUserDepartments(db, userID)
	if err != nil {
		departments = []*models.Department{}
	}

	// Fetch teacher specific data if applicable
	var subjects []*models.Subject
	var classes []*models.Class
	isTeacher := false

	for _, role := range roles {
		if role != nil && (role.Name == "class_teacher" || role.Name == "subject_teacher") {
			isTeacher = true
			break
		}
	}

	if isTeacher {
		subjects, _ = database.GetTeacherSubjects(db, userID)
		classes, _ = database.GetTeacherClasses(db, userID)
	}

	if subjects == nil {
		subjects = []*models.Subject{}
	}
	if classes == nil {
		classes = []*models.Class{}
	}

	// Handle case where user has no roles (for Role label)
	roleName := "Member"
	if len(roles) > 0 && roles[0] != nil {
		roleName = roles[0].Name
	}

	return c.Render("auth/profile", fiber.Map{
		"Title":       "Profile - Swadiq Schools",
		"CurrentPage": "profile",
		"user":        user,
		"Roles":       roles,
		"Departments": departments,
		"Subjects":    subjects,
		"Classes":     classes,
		"Role":        roleName,
		"FirstName":   user.FirstName,
		"LastName":    user.LastName,
		"Email":       user.Email,
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
