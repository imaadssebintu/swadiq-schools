package attendance

import (
	"fmt"
	"strings"
	"swadiq-schools/app/config"
	"swadiq-schools/app/database"
	"swadiq-schools/app/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetAttendanceByClassAPI(c *fiber.Ctx) error {
	classID := c.Params("classId")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	students, err := database.GetStudentsByClass(config.GetDB(), classID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}

	return c.JSON(fiber.Map{
		"students": students,
		"count":    len(students),
	})
}

func GetAttendanceByClassAndDateAPI(c *fiber.Ctx) error {
	classID := c.Params("classId")
	dateStr := c.Params("date")

	if classID == "" || dateStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID and date are required"})
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	attendanceRecords, err := database.GetAttendanceByClassAndDate(config.GetDB(), classID, date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch attendance records"})
	}

	return c.JSON(fiber.Map{
		"attendance": attendanceRecords,
		"count":      len(attendanceRecords),
		"date":       dateStr,
		"class_id":   classID,
	})
}

func CreateOrUpdateAttendanceAPI(c *fiber.Ctx) error {
	type AttendanceRequest struct {
		StudentID        string  `json:"student_id" validate:"required,uuid"`
		ClassID          *string `json:"class_id,omitempty"`
		TimetableEntryID *string `json:"timetable_entry_id,omitempty"`
		PaperID          *string `json:"paper_id,omitempty"`
		Date             string  `json:"date" validate:"required"`
		Status           string  `json:"status" validate:"required,oneof=present absent late excused"`
	}

	var req AttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate that either class_id or timetable_entry_id is provided
	if req.ClassID == nil && req.TimetableEntryID == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Either class_id or timetable_entry_id must be provided"})
	}

	// If ClassID is missing but TimetableEntryID is present, fetch the ClassID
	// as it is likely required by the database schema (NOT NULL constraint)
	if req.ClassID == nil && req.TimetableEntryID != nil {
		entry, err := database.GetTimetableEntryByID(config.GetDB(), *req.TimetableEntryID)
		if err == nil && entry != nil {
			req.ClassID = &entry.ClassID
		}
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	// Validate status
	var status models.AttendanceStatus
	switch req.Status {
	case "present":
		status = models.Present
	case "absent":
		status = models.Absent
	case "late":
		status = models.Late
	case "excused":
		status = models.Excused
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid status. Must be present, absent, late, or excused"})
	}

	// Get current user ID for marked_by
	user := c.Locals("user").(*models.User)
	markedBy := user.ID

	attendance := &models.Attendance{
		StudentID:        req.StudentID,
		ClassID:          req.ClassID,
		TimetableEntryID: req.TimetableEntryID,
		PaperID:          req.PaperID,
		Date:             date,
		Status:           status,
		MarkedBy:         &markedBy,
	}

	if err := database.CreateOrUpdateAttendance(config.GetDB(), attendance); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save attendance record"})
	}

	return c.JSON(fiber.Map{
		"message":    "Attendance record saved successfully",
		"attendance": attendance,
	})
}

func BatchUpdateAttendanceAPI(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Batch update not implemented"})
}

func GetAttendanceStatsAPI(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"stats": map[string]interface{}{}})
}

func GetStudentsByTimetableEntryAPI(c *fiber.Ctx) error {
	timetableEntryID := c.Params("timetableEntryId")
	if timetableEntryID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Timetable entry ID is required"})
	}

	students, err := database.GetStudentsByTimetableEntry(config.GetDB(), timetableEntryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
	}

	return c.JSON(fiber.Map{
		"students": students,
		"count":    len(students),
	})
}

func GetAttendanceByTimetableEntryAPI(c *fiber.Ctx) error {
	timetableEntryID := c.Params("timetableEntryId")
	dateStr := c.Params("date")

	if timetableEntryID == "" || dateStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Timetable entry ID and date are required"})
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	attendanceRecords, err := database.GetAttendanceByTimetableEntryAndDate(config.GetDB(), timetableEntryID, date)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch attendance records"})
	}

	return c.JSON(fiber.Map{
		"attendance":         attendanceRecords,
		"count":              len(attendanceRecords),
		"date":               dateStr,
		"timetable_entry_id": timetableEntryID,
	})
}

func GetCurrentUserLessonsAPI(c *fiber.Ctx) error {
	dayOfWeek := c.Params("dayOfWeek")

	if dayOfWeek == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Day of week is required"})
	}

	// Get current user
	user := c.Locals("user").(*models.User)
	teacherID := user.ID

	// Ensure day of week is lowercase
	dayOfWeek = strings.ToLower(dayOfWeek)
	fmt.Printf("DEBUG: Day of Week: %s, Teacher ID: %s\n", dayOfWeek, teacherID)

	timetableEntries, err := database.GetTimetableEntriesByTeacherAndDay(config.GetDB(), teacherID, dayOfWeek)
	fmt.Printf("DEBUG: Found %d timetable entries\n", len(timetableEntries))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch timetable entries"})
	}

	return c.JSON(fiber.Map{
		"timetable_entries": timetableEntries,
		"count":             len(timetableEntries),
		"day_of_week":       dayOfWeek,
		"teacher_id":        teacherID,
	})
}

// Get timetable entries for a teacher on a specific date
func GetTimetableEntriesByTeacherAndDateAPI(c *fiber.Ctx) error {
	teacherID := c.Params("teacherId")
	dateStr := c.Params("date")

	if teacherID == "" || dateStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teacher ID and date are required"})
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	// Get day of week
	dayOfWeek := strings.ToLower(date.Weekday().String())

	timetableEntries, err := database.GetTimetableEntriesByTeacherAndDay(config.GetDB(), teacherID, dayOfWeek)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch timetable entries"})
	}

	return c.JSON(fiber.Map{
		"timetable_entries": timetableEntries,
		"count":             len(timetableEntries),
		"date":              dateStr,
		"teacher_id":        teacherID,
	})
}

// GetAllLessonsAPI returns all timetable entries for a day (admin/head teacher only)
func GetAllLessonsAPI(c *fiber.Ctx) error {
	dayOfWeek := c.Params("dayOfWeek")

	if dayOfWeek == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Day of week is required"})
	}

	// Get current user
	user := c.Locals("user").(*models.User)

	// Check if user can access all classes
	if !user.CanAccessAllClasses() {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied. Admin or head teacher role required."})
	}

	// Ensure day of week is lowercase
	dayOfWeek = strings.ToLower(dayOfWeek)
	fmt.Printf("DEBUG: Getting all lessons for day: %s\n", dayOfWeek)

	timetableEntries, err := database.GetAllTimetableEntriesByDay(config.GetDB(), dayOfWeek)
	fmt.Printf("DEBUG: Found %d total timetable entries\n", len(timetableEntries))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch timetable entries"})
	}

	return c.JSON(fiber.Map{
		"timetable_entries": timetableEntries,
		"count":             len(timetableEntries),
		"day_of_week":       dayOfWeek,
		"user_id":           user.ID,
	})
}
