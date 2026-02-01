package students

import (
	"strings"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetStudentsAPI(c *fiber.Ctx) error {
	classID := c.Query("class_id")
	limit := c.QueryInt("limit", 0)
	offset := c.QueryInt("offset", 0)

	// If class_id is provided, use filtered query
	if classID != "" {
		filters := database.StudentFilters{
			ClassID: classID,
			Limit:   limit,
			Offset:  offset,
		}
		students, totalCount, err := database.GetStudentsWithFiltersAndPagination(config.GetDB(), filters)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
		}
		return c.JSON(fiber.Map{
			"students":    students,
			"count":       len(students),
			"total_count": totalCount,
		})
	}

	// Default behavior - get all students
	students, err := database.GetStudentsWithDetails(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}

	return c.JSON(fiber.Map{
		"students": students,
		"count":    len(students),
	})
}

// GetStudentsStatsAPI returns students statistics for the students page
func GetStudentsStatsAPI(c *fiber.Ctx) error {
	// Get database connection
	db := config.GetDB()

	// Get students statistics
	stats, err := database.GetStudentsStats(db)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch students statistics",
			"details": err.Error(),
		})
	}

	// Return statistics as JSON
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetStudentsTableAPI returns students formatted for table display with filtering support and pagination
func GetStudentsTableAPI(c *fiber.Ctx) error {
	// Get query parameters for filtering
	search := c.Query("search")
	status := c.Query("status")
	classID := c.Query("class_id")
	classIDs := c.Query("class_ids") // Support multiple class IDs
	gender := c.Query("gender")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	sortBy := c.Query("sort_by", "student_id") // default to student_id
	sortOrder := c.Query("sort_order", "asc")  // default to ascending

	// Get pagination parameters
	limit := c.QueryInt("limit", 10)  // default to 10 students per page
	offset := c.QueryInt("offset", 0) // default to start from beginning

	// Create filter parameters
	filters := database.StudentFilters{
		Search:    search,
		Status:    status,
		ClassID:   classID,
		ClassIDs:  classIDs, // Add support for multiple class IDs
		Gender:    gender,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Offset:    offset,
	}

	// Use GetStudentsWithDetails to get parent data, then apply filters
	allStudents, err := database.GetStudentsWithDetails(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students: " + err.Error()})
	}

	// Apply filters manually
	var filteredStudents []*models.Student
	for _, student := range allStudents {
		// Apply search filter
		if filters.Search != "" && len(filters.Search) >= 3 {
			searchLower := strings.ToLower(filters.Search)
			fullName := strings.ToLower(student.FirstName + " " + student.LastName)
			if !strings.Contains(strings.ToLower(student.FirstName), searchLower) &&
				!strings.Contains(strings.ToLower(student.LastName), searchLower) &&
				!strings.Contains(fullName, searchLower) &&
				!strings.Contains(strings.ToLower(student.StudentID), searchLower) {
				continue
			}
		}

		// Apply status filter
		if filters.Status != "" {
			if (filters.Status == "active" && !student.IsActive) ||
				(filters.Status == "inactive" && student.IsActive) {
				continue
			}
		}

		// Apply class filter
		if filters.ClassID != "" {
			if student.ClassID == nil || *student.ClassID != filters.ClassID {
				continue
			}
		}

		// Apply gender filter
		if filters.Gender != "" {
			if student.Gender == nil || string(*student.Gender) != filters.Gender {
				continue
			}
		}

		filteredStudents = append(filteredStudents, student)
	}

	totalCount := len(filteredStudents)

	// Apply pagination
	students := filteredStudents
	if limit > 0 {
		start := offset
		end := offset + limit
		if start > len(filteredStudents) {
			students = []*models.Student{}
		} else {
			if end > len(filteredStudents) {
				end = len(filteredStudents)
			}
			students = filteredStudents[start:end]
		}
	}

	// Format students for table display
	type StudentTableData struct {
		ID          string `json:"id"`
		StudentID   string `json:"student_id"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		FullName    string `json:"full_name"`
		ClassID     string `json:"class_id,omitempty"`
		ClassName   string `json:"class_name,omitempty"`
		ClassCode   string `json:"class_code,omitempty"`
		ParentName  string `json:"parent_name,omitempty"`
		ParentPhone string `json:"parent_phone,omitempty"`
		ParentEmail string `json:"parent_email,omitempty"`
		Status      string `json:"status"`
		Initials    string `json:"initials"`
		DateOfBirth string `json:"date_of_birth,omitempty"`
		Gender      string `json:"gender,omitempty"`
		Address     string `json:"address,omitempty"`
	}

	var tableData []StudentTableData
	for _, student := range students {
		// Create initials from first and last name
		initials := "??"
		if len(student.FirstName) > 0 && len(student.LastName) > 0 {
			initials = string(student.FirstName[0]) + string(student.LastName[0])
		} else if len(student.FirstName) > 0 {
			initials = string(student.FirstName[0]) + "?"
		} else if len(student.LastName) > 0 {
			initials = "?" + string(student.LastName[0])
		}

		// Get primary parent (first one in the list)
		parentName := ""
		parentPhone := ""
		parentEmail := ""
		if len(student.Parents) > 0 {
			parent := student.Parents[0]
			parentName = parent.FirstName + " " + parent.LastName
			if parent.Phone != nil {
				parentPhone = *parent.Phone
			}
			if parent.Email != nil {
				parentEmail = *parent.Email
			}
		}

		// Get class name, code and ID
		className := ""
		classCode := ""
		classID := ""
		if student.Class != nil {
			className = student.Class.Name
			if student.Class.Code != nil {
				classCode = *student.Class.Code
			}
			classID = student.Class.ID
		} else if student.ClassID != nil {
			classID = *student.ClassID
		}

		// Format date of birth
		dateOfBirth := ""
		if student.DateOfBirth != nil {
			dateOfBirth = student.DateOfBirth.Format("2006-01-02")
		}

		// Format gender
		gender := ""
		if student.Gender != nil {
			gender = string(*student.Gender)
		}

		// Format address
		address := ""
		if student.Address != nil {
			address = *student.Address
		}

		// Determine status
		status := "Active"
		if !student.IsActive {
			status = "Inactive"
		}

		tableData = append(tableData, StudentTableData{
			ID:          student.ID,
			StudentID:   student.StudentID,
			FirstName:   student.FirstName,
			LastName:    student.LastName,
			FullName:    student.FirstName + " " + student.LastName,
			ClassID:     classID,
			ClassName:   className,
			ClassCode:   classCode,
			ParentName:  parentName,
			ParentPhone: parentPhone,
			ParentEmail: parentEmail,
			Status:      status,
			Initials:    initials,
			DateOfBirth: dateOfBirth,
			Gender:      gender,
			Address:     address,
		})
	}

	return c.JSON(fiber.Map{
		"students":    tableData,
		"count":       len(tableData),
		"total_count": totalCount,
		"has_more":    offset+limit < totalCount,
		"next_offset": offset + limit,
	})
}

// GetStudentByIDAPI returns a single student by ID
func GetStudentByIDAPI(c *fiber.Ctx) error {
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
	}

	student, err := database.GetStudentByID(config.GetDB(), studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Extract class name if available
	className := ""
	if student.Class != nil {
		className = student.Class.Name
	}

	// Format response for edit modal
	response := fiber.Map{
		"student": fiber.Map{
			"id":         student.ID,
			"student_id": student.StudentID,
			"first_name": student.FirstName,
			"last_name":  student.LastName,
			"date_of_birth": func() string {
				if student.DateOfBirth != nil {
					return student.DateOfBirth.Format("2006-01-02")
				}
				return ""
			}(),
			"gender": func() string {
				if student.Gender != nil {
					return string(*student.Gender)
				}
				return ""
			}(),
			"address": func() string {
				if student.Address != nil {
					return *student.Address
				}
				return ""
			}(),
			"class_id": func() string {
				if student.ClassID != nil {
					return *student.ClassID
				}
				return ""
			}(),
			"class_name": className,
			"is_active":  student.IsActive,
		},
	}

	// Add parent information if available
	if len(student.Parents) > 0 {
		parent := student.Parents[0] // Get primary parent

		// Get relationship information
		relationship, err := database.GetStudentParentRelationship(config.GetDB(), studentID, parent.ID)
		if err != nil {
			relationship = "guardian" // Default if not found
		}

		response["parent"] = fiber.Map{
			"id":           parent.ID,
			"first_name":   parent.FirstName,
			"last_name":    parent.LastName,
			"email":        parent.Email,
			"phone":        parent.Phone,
			"address":      parent.Address,
			"relationship": relationship,
		}
	}

	return c.JSON(response)
}

// GetStudentsByYearAPI returns students for a specific year
func GetStudentsByYearAPI(c *fiber.Ctx) error {
	year := c.QueryInt("year")
	if year == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Year parameter is required"})
	}

	students, err := database.GetStudentsByYear(config.GetDB(), year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}

	return c.JSON(fiber.Map{
		"students": students,
		"count":    len(students),
		"year":     year,
	})
}

// GetStudentsByClassAPI returns students for a specific class
func GetStudentsByClassAPI(c *fiber.Ctx) error {
	classID := c.Query("class_id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID parameter is required"})
	}

	students, err := database.GetStudentsByClass(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}

	return c.JSON(fiber.Map{
		"students": students,
		"count":    len(students),
		"class_id": classID,
	})
}

func CreateStudentAPI(c *fiber.Ctx) error {
	type CreateStudentRequest struct {
		StudentID          string `json:"student_id"` // Optional - will be auto-generated if empty
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		DateOfBirth        string `json:"date_of_birth"`
		Gender             string `json:"gender"`
		Address            string `json:"address"`
		ClassID            string `json:"class_id"`
		ParentID           string `json:"parent_id"`
		ParentRelationship string `json:"parent_relationship"`
	}

	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Only first name and last name are required now
	if req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "First name and last name are required"})
	}

	// Always auto-generate student ID
	studentID, err := GenerateStudentID(config.GetDB())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate student ID"})
	}

	// Create student
	student := &models.Student{
		StudentID: studentID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if req.DateOfBirth != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.DateOfBirth); err == nil {
			student.DateOfBirth = &parsedDate
		}
	}
	if req.Gender != "" {
		gender := models.Gender(req.Gender)
		student.Gender = &gender
	}
	if req.Address != "" {
		student.Address = &req.Address
	}
	if req.ClassID != "" {
		student.ClassID = &req.ClassID
	}

	if err := database.CreateStudent(config.GetDB(), student); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create student"})
	}

	// Link student to parent if provided
	if req.ParentID != "" {
		relationship := req.ParentRelationship
		if relationship == "" {
			relationship = "guardian" // Default relationship
		}
		if err := database.LinkStudentToParent(config.GetDB(), student.ID, req.ParentID, relationship); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to link student to parent"})
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Student created successfully",
		"student": student,
	})
}

// UpdateStudentAPI updates an existing student
func UpdateStudentAPI(c *fiber.Ctx) error {
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
	}

	type ParentInfo struct {
		ID           string `json:"id,omitempty"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		Relationship string `json:"relationship"`
	}

	type UpdateStudentRequest struct {
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		DateOfBirth        string `json:"date_of_birth"`
		Gender             string `json:"gender"`
		Address            string `json:"address"`
		ClassID            string `json:"class_id"`
		ParentID           string `json:"parent_id"`
		ParentRelationship string `json:"parent_relationship"`
	}

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get existing student
	student, err := database.GetStudentByID(config.GetDB(), studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Update student fields
	if req.FirstName != "" {
		student.FirstName = req.FirstName
	}
	if req.LastName != "" {
		student.LastName = req.LastName
	}
	if req.DateOfBirth != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.DateOfBirth); err == nil {
			student.DateOfBirth = &parsedDate
		}
	}
	if req.Gender != "" {
		gender := models.Gender(req.Gender)
		student.Gender = &gender
	}
	if req.Address != "" {
		student.Address = &req.Address
	}
	if req.ClassID != "" {
		student.ClassID = &req.ClassID
	}

	// Update student in database
	if err := database.UpdateStudent(config.GetDB(), student); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update student"})
	}

	// Handle parent changes
	if req.ParentID != "" {
		// Change the parent for this student
		relationship := req.ParentRelationship
		if relationship == "" {
			relationship = "guardian" // Default relationship
		}
		if err := database.ChangeStudentParent(config.GetDB(), studentID, req.ParentID, relationship); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update parent relationship"})
		}
	} else if req.ParentID == "" && req.ParentRelationship == "" {
		// Remove parent if both are empty (this handles the remove parent case)
		if err := database.ChangeStudentParent(config.GetDB(), studentID, "", ""); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to remove parent"})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Student updated successfully",
		"student": student,
	})
}

// DeleteStudentAPI deletes a student
func DeleteStudentAPI(c *fiber.Ctx) error {
	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
	}

	// Check if student exists
	_, err := database.GetStudentByID(config.GetDB(), studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Delete student
	if err := database.DeleteStudent(config.GetDB(), studentID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete student"})
	}

	return c.JSON(fiber.Map{
		"message": "Student deleted successfully",
	})
}

// GetParentsAPI returns all parents for selection
func GetParentsAPI(c *fiber.Ctx) error {
	search := c.Query("search", "")
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	parents, err := database.GetParentsForSelection(config.GetDB(), search, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch parents"})
	}

	return c.JSON(fiber.Map{
		"parents": parents,
		"count":   len(parents),
	})
}

// SearchStudentsAPI searches for students by name or student ID
func SearchStudentsAPI(c *fiber.Ctx) error {
	query := c.Query("q", "")

	if query == "" {
		return c.JSON(fiber.Map{
			"students": []interface{}{},
			"count":    0,
		})
	}

	students, err := database.SearchStudents(config.GetDB(), query)
	if err != nil {
		return c.JSON(fiber.Map{
			"students": []interface{}{},
			"count":    0,
		})
	}

	// Format for search results
	type SearchResult struct {
		ID        string `json:"id"`
		StudentID string `json:"student_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		FullName  string `json:"full_name"`
		Gender    string `json:"gender"`
		ClassName string `json:"class_name"`
	}

	var results []SearchResult
	for _, student := range students {
		className := "Not assigned"
		if student.Class != nil {
			className = student.Class.Name
		}

		results = append(results, SearchResult{
			ID:        student.ID,
			StudentID: student.StudentID,
			FirstName: student.FirstName,
			LastName:  student.LastName,
			FullName:  student.FirstName + " " + student.LastName,
			Gender:    string(*student.Gender),
			ClassName: className,
		})
	}

	return c.JSON(fiber.Map{
		"students": results,
		"count":    len(results),
	})
}
