package database

import (
	"database/sql"
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
