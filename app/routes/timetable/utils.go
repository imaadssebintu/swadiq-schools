package timetable

import (
	"database/sql"
	"strings"
)

// ValidateTimeFormat validates time format (HH:MM)
func ValidateTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	return len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 2
}

// ValidateDayOfWeek validates day of week
func ValidateDayOfWeek(day string) bool {
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	day = strings.ToLower(day)
	for _, validDay := range validDays {
		if day == validDay {
			return true
		}
	}
	return false
}

// CheckTimeConflict checks if there's a time conflict for a teacher or class
func CheckTimeConflict(db *sql.DB, teacherID, classID, dayOfWeek, startTime, endTime string, excludeID string) (bool, error) {
	query := `SELECT COUNT(*) FROM timetable_entries 
			  WHERE (teacher_id = $1 OR class_id = $2) 
			  AND day_of_week = $3 
			  AND is_active = true
			  AND (
				  (start_time <= $4 AND end_time > $4) OR
				  (start_time < $5 AND end_time >= $5) OR
				  (start_time >= $4 AND end_time <= $5)
			  )`
	
	args := []interface{}{teacherID, classID, dayOfWeek, startTime, endTime}
	
	if excludeID != "" {
		query += " AND id != $6"
		args = append(args, excludeID)
	}
	
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}