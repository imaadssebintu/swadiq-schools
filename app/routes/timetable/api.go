package timetable

import (
	"encoding/json"
	"fmt"
	"log"
	"swadiq-schools/app/config"

	"github.com/gofiber/fiber/v2"
)

func GetTimetableAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	
	query := `SELECT id, class_id, subject_id, teacher_id, day_of_week, start_time, end_time, room
			  FROM timetable_entries 
			  WHERE is_active = true 
			  ORDER BY day_of_week, start_time`
	
	rows, err := db.Query(query)
	if err != nil {
		return c.JSON(fiber.Map{
			"entries": []interface{}{},
			"count":   0,
		})
	}
	defer rows.Close()

	var entries []fiber.Map
	for rows.Next() {
		var id, classID, subjectID, teacherID, dayOfWeek, startTime, endTime, room string
		if err := rows.Scan(&id, &classID, &subjectID, &teacherID, &dayOfWeek, &startTime, &endTime, &room); err != nil {
			continue
		}
		entries = append(entries, fiber.Map{
			"id":          id,
			"class_id":    classID,
			"subject_id":  subjectID,
			"teacher_id":  teacherID,
			"day_of_week": dayOfWeek,
			"start_time":  startTime,
			"end_time":    endTime,
			"room":        room,
		})
	}

	return c.JSON(fiber.Map{
		"entries": entries,
		"count":   len(entries),
	})
}

func CreateTimetableEntryAPI(c *fiber.Ctx) error {
	type CreateEntryRequest struct {
		ClassID    string `json:"class_id"`
		SubjectID  string `json:"subject_id"`
		TeacherID  string `json:"teacher_id"`
		DayOfWeek  string `json:"day_of_week"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
		Room       string `json:"room"`
	}

	var req CreateEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.ClassID == "" || req.SubjectID == "" || req.TeacherID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class, subject, and teacher are required"})
	}

	db := config.GetDB()
	query := `INSERT INTO timetable_entries (class_id, subject_id, teacher_id, day_of_week, start_time, end_time, room, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW(), NOW())`
	
	_, err := db.Exec(query, req.ClassID, req.SubjectID, req.TeacherID, req.DayOfWeek, req.StartTime, req.EndTime, req.Room)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create timetable entry"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Timetable entry created successfully"})
}

func UpdateTimetableEntryAPI(c *fiber.Ctx) error {
	entryID := c.Params("id")
	if entryID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Entry ID is required"})
	}

	type UpdateEntryRequest struct {
		ClassID    string `json:"class_id"`
		SubjectID  string `json:"subject_id"`
		TeacherID  string `json:"teacher_id"`
		DayOfWeek  string `json:"day_of_week"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
		Room       string `json:"room"`
	}

	var req UpdateEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	db := config.GetDB()
	query := `UPDATE timetable_entries 
			  SET class_id = $1, subject_id = $2, teacher_id = $3, day_of_week = $4, 
				  start_time = $5, end_time = $6, room = $7, updated_at = NOW()
			  WHERE id = $8`
	
	_, err := db.Exec(query, req.ClassID, req.SubjectID, req.TeacherID, req.DayOfWeek, req.StartTime, req.EndTime, req.Room, entryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update timetable entry"})
	}

	return c.JSON(fiber.Map{"message": "Timetable entry updated successfully"})
}

func DeleteTimetableEntryAPI(c *fiber.Ctx) error {
	entryID := c.Params("id")
	if entryID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Entry ID is required"})
	}

	db := config.GetDB()
	query := `UPDATE timetable_entries SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	_, err := db.Exec(query, entryID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete timetable entry"})
	}

	return c.JSON(fiber.Map{"message": "Timetable entry deleted successfully"})
}

func GetClassTimetableAPI(c *fiber.Ctx) error {
	classID := c.Params("id")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	db := config.GetDB()
	query := `
		SELECT te.id, te.day_of_week, te.start_time, te.end_time,
			   s.name as subject_name, p.name as paper_name,
			   t.first_name || ' ' || t.last_name as teacher_name
		FROM timetable_entries te
		LEFT JOIN papers p ON te.paper_id = p.id
		LEFT JOIN subjects s ON te.subject_id = s.id
		LEFT JOIN users t ON te.teacher_id = t.id
		WHERE te.class_id = $1 AND te.is_active = true
		ORDER BY te.start_time, te.day_of_week
	`

	rows, err := db.Query(query, classID)
	if err != nil {
		log.Printf("Error fetching timetable: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch timetable", "details": err.Error()})
	}
	defer rows.Close()

	timetable := make([]fiber.Map, 0)
	for rows.Next() {
		var id, day, startTime, endTime, subjectName, paperName, teacherName string
		if err := rows.Scan(&id, &day, &startTime, &endTime, &subjectName, &paperName, &teacherName); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to scan timetable entry", "details": err.Error()})
		}
		timetable = append(timetable, fiber.Map{
			"id":           id,
			"day":          day,
			"time_slot":    fmt.Sprintf("%s - %s", startTime, endTime),
			"subject_name": subjectName,
			"paper_name":   paperName,
			"teacher_name": teacherName,
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"timetable": timetable,
	})
}

func SaveClassTimetableAPI(c *fiber.Ctx) error {
	classID := c.Params("classId")
	if classID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class ID is required"})
	}

	type TimetableEntry struct {
		TimeSlot  string `json:"time_slot"`
		Day       string `json:"day"`
		PaperID   string `json:"paper_id"`
		TeacherID string `json:"teacher_id"`
	}

	type TimetableRequest struct {
		Timetable []TimetableEntry `json:"timetable"`
	}

	var req TimetableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	db := config.GetDB()
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Delete existing timetable for the class
	if _, err := tx.Exec("DELETE FROM timetable_entries WHERE class_id = $1", classID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear existing timetable"})
	}

	for _, entry := range req.Timetable {
		if entry.PaperID == "" {
			continue // Skip entries without a paper
		}

		// Get subject_id from paper
		var subjectID string
		err := tx.QueryRow("SELECT subject_id FROM papers WHERE id = $1", entry.PaperID).Scan(&subjectID)
		if err != nil {
			// If paper is not found, we can either skip or return an error
			// For now, let's skip it
			continue
		}

		query := `INSERT INTO timetable_entries (class_id, subject_id, paper_id, teacher_id, day_of_week, start_time, end_time, is_active, created_at, updated_at)
				  VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW(), NOW())`

		// Extract start and end times from time_slot
		var startTime, endTime string
		fmt.Sscanf(entry.TimeSlot, "%s - %s", &startTime, &endTime)

		_, err = tx.Exec(query, classID, subjectID, entry.PaperID, entry.TeacherID, entry.Day, startTime, endTime)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create timetable entry", "details": err.Error()})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Timetable saved successfully",
	})
}

// Timetable Settings APIs
func GetTimetableSettingsAPI(c *fiber.Ctx) error {
	classID := c.Params("classId")
	db := config.GetDB()
	
	query := `SELECT id, class_id, days, start_time, end_time, lesson_duration, breaks, is_default 
			  FROM timetable_settings WHERE class_id = $1 AND is_default = false`
	
	var settings struct {
		ID             string `db:"id"`
		ClassID        string `db:"class_id"`
		Days           string `db:"days"`
		StartTime      string `db:"start_time"`
		EndTime        string `db:"end_time"`
		LessonDuration int    `db:"lesson_duration"`
		Breaks         string `db:"breaks"`
		IsDefault      bool   `db:"is_default"`
	}
	
	err := db.QueryRow(query, classID).Scan(
		&settings.ID, &settings.ClassID, &settings.Days, &settings.StartTime,
		&settings.EndTime, &settings.LessonDuration, &settings.Breaks, &settings.IsDefault,
	)
	
	if err != nil {
		// Return default settings if none found
		return c.JSON(fiber.Map{
			"success": true,
			"settings": fiber.Map{
				"days":            json.RawMessage(`["monday","tuesday","wednesday","thursday","friday"]`),
				"start_time":      "08:00:00",
				"end_time":        "16:00:00",
				"lesson_duration": 60,
				"breaks":          json.RawMessage(`[{"name":"Breakfast Break","start_time":"10:00","end_time":"10:30"},{"name":"Lunch Break","start_time":"12:30","end_time":"13:30"}]`),
			},
		})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"settings": fiber.Map{
			"days":            json.RawMessage(settings.Days),
			"start_time":      settings.StartTime,
			"end_time":        settings.EndTime,
			"lesson_duration": settings.LessonDuration,
			"breaks":          json.RawMessage(settings.Breaks),
		},
	})
}

func SaveTimetableSettingsAPI(c *fiber.Ctx) error {
	classID := c.Params("classId")
	
	type SettingsRequest struct {
		Days           []string    `json:"days"`
		StartTime      string      `json:"start_time"`
		EndTime        string      `json:"end_time"`
		LessonDuration int         `json:"lesson_duration"`
		Breaks         []fiber.Map `json:"breaks"`
	}
	
	var req SettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	db := config.GetDB()
	
	// Convert arrays to JSON strings safely
	daysJSON, _ := json.Marshal(req.Days)
	breaksJSON, _ := json.Marshal(req.Breaks)
	
	// Check if settings exist for this class
	var existingID string
	err := db.QueryRow("SELECT id FROM timetable_settings WHERE class_id = $1 AND is_default = false", classID).Scan(&existingID)
	
	if err != nil {
		// Create new settings
		query := `INSERT INTO timetable_settings (class_id, days, start_time, end_time, lesson_duration, breaks, is_default, created_at, updated_at)
				  VALUES ($1, $2, $3, $4, $5, $6, false, NOW(), NOW())`
		
		_, err = db.Exec(query, classID, string(daysJSON), req.StartTime, req.EndTime, req.LessonDuration, string(breaksJSON))
	} else {
		// Update existing settings
		query := `UPDATE timetable_settings 
				  SET days = $2, start_time = $3, end_time = $4, lesson_duration = $5, breaks = $6, updated_at = NOW()
				  WHERE class_id = $1 AND is_default = false`
		
		_, err = db.Exec(query, classID, string(daysJSON), req.StartTime, req.EndTime, req.LessonDuration, string(breaksJSON))
	}
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save settings"})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings saved successfully",
	})
}

func GetDefaultTimetableSettingsAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	
	query := `SELECT id, days, start_time, end_time, lesson_duration, breaks 
			  FROM timetable_settings WHERE is_default = true LIMIT 1`
	
	var settings struct {
		ID             string `db:"id"`
		Days           string `db:"days"`
		StartTime      string `db:"start_time"`
		EndTime        string `db:"end_time"`
		LessonDuration int    `db:"lesson_duration"`
		Breaks         string `db:"breaks"`
	}
	
	err := db.QueryRow(query).Scan(
		&settings.ID, &settings.Days, &settings.StartTime,
		&settings.EndTime, &settings.LessonDuration, &settings.Breaks,
	)
	
	if err != nil {
		// Return hardcoded default if none in database
		return c.JSON(fiber.Map{
			"success": true,
			"settings": fiber.Map{
				"days":            json.RawMessage(`["monday","tuesday","wednesday","thursday","friday"]`),
				"start_time":      "08:00",
				"end_time":        "16:00",
				"lesson_duration": 60,
				"breaks":          json.RawMessage(`[{"name":"Breakfast Break","start_time":"10:00","end_time":"10:30"},{"name":"Lunch Break","start_time":"12:30","end_time":"13:30"}]`),
			},
		})
	}
	
	// Return raw JSON strings for days and breaks
	daysJSON := settings.Days
	if daysJSON == "" {
		daysJSON = `["monday","tuesday","wednesday","thursday","friday"]`
	}
	
	breaksJSON := settings.Breaks
	if breaksJSON == "" {
		breaksJSON = `[{"name":"Breakfast Break","start_time":"10:00","end_time":"10:30"},{"name":"Lunch Break","start_time":"12:30","end_time":"13:30"}]`
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"settings": fiber.Map{
			"days":            json.RawMessage(daysJSON),
			"start_time":      settings.StartTime,
			"end_time":        settings.EndTime,
			"lesson_duration": settings.LessonDuration,
			"breaks":          json.RawMessage(breaksJSON),
		},
	})
}

func SaveDefaultTimetableSettingsAPI(c *fiber.Ctx) error {
	type SettingsRequest struct {
		Days           []string    `json:"days"`
		StartTime      string      `json:"start_time"`
		EndTime        string      `json:"end_time"`
		LessonDuration int         `json:"lesson_duration"`
		Breaks         []fiber.Map `json:"breaks"`
	}
	
	var req SettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	db := config.GetDB()
	
	// Convert arrays to JSON strings safely
	daysJSON := `[]`
	if len(req.Days) > 0 {
		daysJSON = `["` + req.Days[0]
		for i := 1; i < len(req.Days); i++ {
			daysJSON += `","` + req.Days[i]
		}
		daysJSON += `"]`
	}
	
	breaksJSON := `[]`
	if len(req.Breaks) > 0 {
		breaksJSON = `[`
		for i, b := range req.Breaks {
			if i > 0 {
				breaksJSON += `,`
			}
			// Safely get values with type assertion and nil checks
			name := ""
			startTime := ""
			endTime := ""
			if nameVal, ok := b["name"]; ok && nameVal != nil {
				if nameStr, ok := nameVal.(string); ok {
					name = nameStr
				}
			}
			if startVal, ok := b["start_time"]; ok && startVal != nil {
				if startStr, ok := startVal.(string); ok {
					startTime = startStr
				}
			}
			if endVal, ok := b["end_time"]; ok && endVal != nil {
				if endStr, ok := endVal.(string); ok {
					endTime = endStr
				}
			}
			breaksJSON += `{"name":"` + name + `","start_time":"` + startTime + `","end_time":"` + endTime + `"}`
		}
		breaksJSON += `]`
	}
	
	// Delete all existing default settings
	_, err := db.Exec("DELETE FROM timetable_settings WHERE is_default = true")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear existing defaults"})
	}
	
	// Insert new default
	query := `INSERT INTO timetable_settings (class_id, days, start_time, end_time, lesson_duration, breaks, is_default, created_at, updated_at)
			  VALUES ('', $1, $2, $3, $4, $5, true, NOW(), NOW())`
	
	_, err = db.Exec(query, daysJSON, req.StartTime, req.EndTime, req.LessonDuration, breaksJSON)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save default settings"})
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Default settings saved successfully",
	})
}

func ApplyDefaultSettingsAPI(c *fiber.Ctx) error {
	db := config.GetDB()
	
	// Get default settings
	query := `SELECT days, start_time, end_time, lesson_duration, breaks 
			  FROM timetable_settings WHERE is_default = true LIMIT 1`
	
	var defaultSettings struct {
		Days           string `db:"days"`
		StartTime      string `db:"start_time"`
		EndTime        string `db:"end_time"`
		LessonDuration int    `db:"lesson_duration"`
		Breaks         string `db:"breaks"`
	}
	
	err := db.QueryRow(query).Scan(
		&defaultSettings.Days, &defaultSettings.StartTime, &defaultSettings.EndTime,
		&defaultSettings.LessonDuration, &defaultSettings.Breaks,
	)
	
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No default settings found"})
	}
	
	// Get all classes without custom timetable settings
	classesQuery := `SELECT c.id FROM classes c 
					 LEFT JOIN timetable_settings ts ON c.id = ts.class_id 
					 WHERE ts.class_id IS NULL`
	
	rows, err := db.Query(classesQuery)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get classes"})
	}
	defer rows.Close()
	
	var classIDs []string
	for rows.Next() {
		var classID string
		if err := rows.Scan(&classID); err != nil {
			continue
		}
		classIDs = append(classIDs, classID)
	}
	
	// Apply default settings to classes without custom settings
	var appliedCount int
	for _, classID := range classIDs {
		insertQuery := `INSERT INTO timetable_settings (class_id, days, start_time, end_time, lesson_duration, breaks, is_default, created_at, updated_at)
						VALUES ($1, $2, $3, $4, $5, $6, false, NOW(), NOW())`
		
		_, err = db.Exec(insertQuery, classID, defaultSettings.Days, defaultSettings.StartTime,
			defaultSettings.EndTime, defaultSettings.LessonDuration, defaultSettings.Breaks)
		
		if err == nil {
			appliedCount++
		}
	}
	
	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Applied default settings to %d classes", appliedCount),
		"applied_count": appliedCount,
	})
}