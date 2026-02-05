package database

import (
	"database/sql"
	"log"
	"strings"
	"swadiq-schools/app/models"
	"time"
)

// CreateOrUpdateTeacherAttendance saves a teacher's attendance record
func CreateOrUpdateTeacherAttendance(db *sql.DB, attendance *models.TeacherAttendance) error {
	query := `INSERT INTO teacher_attendances (id, teacher_id, date, status, remarks, created_at, updated_at)
			  VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
			  ON CONFLICT (teacher_id, date) 
			  DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks, updated_at = NOW()`

	_, err := db.Exec(query, attendance.TeacherID, attendance.Date, attendance.Status, attendance.Remarks)
	return err
}

// GetTeacherAttendanceByDate retrieves all teacher attendance records for a specific date
func GetTeacherAttendanceByDate(db *sql.DB, date time.Time) ([]*models.TeacherAttendance, error) {
	query := `SELECT ta.id, ta.teacher_id, ta.date, ta.status, ta.remarks, ta.created_at, ta.updated_at,
			  u.first_name, u.last_name
			  FROM teacher_attendances ta
			  JOIN users u ON ta.teacher_id = u.id
			  WHERE ta.date = $1`

	rows, err := db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.TeacherAttendance
	for rows.Next() {
		record := &models.TeacherAttendance{}
		var firstName, lastName string
		err := rows.Scan(
			&record.ID, &record.TeacherID, &record.Date, &record.Status, &record.Remarks, &record.CreatedAt, &record.UpdatedAt,
			&firstName, &lastName,
		)
		if err != nil {
			continue
		}

		record.Teacher = &models.User{
			FirstName: firstName,
			LastName:  lastName,
		}
		records = append(records, record)
	}

	return records, nil
}

// GetTeacherAttendanceByTeacherAndDate retrieves a specific teacher's attendance for a date
func GetTeacherAttendanceByTeacherAndDate(db *sql.DB, teacherID string, date time.Time) (*models.TeacherAttendance, error) {
	query := `SELECT id, teacher_id, date, status, remarks, created_at, updated_at
			  FROM teacher_attendances
			  WHERE teacher_id = $1 AND date = $2`

	record := &models.TeacherAttendance{}
	err := db.QueryRow(query, teacherID, date).Scan(
		&record.ID, &record.TeacherID, &record.Date, &record.Status, &record.Remarks, &record.CreatedAt, &record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return record, err
}

// DailyStaffSummary represents a teacher's daily status including lesson counts
type DailyStaffSummary struct {
	TeacherID      string `json:"teacher_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	AttendanceID   string `json:"attendance_id,omitempty"`
	Status         string `json:"status"` // present, absent, etc., or "unmarked"
	Remarks        string `json:"remarks"`
	ScheduledCount int    `json:"scheduled_count"`
	ConductedCount int    `json:"conducted_count"`
}

// GetDailyStaffAttendanceSummary retrieves a summary of all teachers for a specific date
// GetDailyStaffAttendanceSummary retrieves a summary of all teachers for a specific date
func GetDailyStaffAttendanceSummary(db *sql.DB, date time.Time) ([]*DailyStaffSummary, error) {
	weekday := strings.ToLower(date.Weekday().String())

	query := `
		WITH scheduled_counts AS (
			SELECT teacher_id, COUNT(*) as scheduled_count
			FROM timetable_entries
			WHERE day_of_week = $2
			GROUP BY teacher_id
		),
		conducted_counts AS (
			SELECT teacher_id, COUNT(*) as conducted_count
			FROM conducted_lessons
			WHERE date = $1
			GROUP BY teacher_id
		)
		SELECT DISTINCT ON (u.id)
			u.id, 
			u.first_name, 
			u.last_name,
			COALESCE(ta.id::text, ''),
			COALESCE(ta.status, 'unmarked'),
			COALESCE(ta.remarks, ''),
			COALESCE(sc.scheduled_count, 0),
			COALESCE(cc.conducted_count, 0)
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		LEFT JOIN teacher_attendances ta ON u.id = ta.teacher_id AND ta.date = $1
		LEFT JOIN scheduled_counts sc ON u.id = sc.teacher_id
		LEFT JOIN conducted_counts cc ON u.id = cc.teacher_id
		WHERE u.is_active = true 
		AND r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher')
		ORDER BY u.id, u.first_name, u.last_name
	`

	rows, err := db.Query(query, date, weekday)
	if err != nil {
		log.Printf("GetDailyStaffAttendanceSummary Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var summaries []*DailyStaffSummary
	for rows.Next() {
		s := &DailyStaffSummary{}
		err := rows.Scan(
			&s.TeacherID, &s.FirstName, &s.LastName,
			&s.AttendanceID, &s.Status, &s.Remarks,
			&s.ScheduledCount, &s.ConductedCount,
		)
		if err != nil {
			log.Printf("GetDailyStaffAttendanceSummary Scan Error: %v", err)
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}
