package models

import (
	"time"
)

// TimetableSettings represents the timetable configuration for a class
type TimetableSettings struct {
	ID             string    `json:"id" db:"id"`
	ClassID        string    `json:"class_id" db:"class_id"`
	Days           []string  `json:"days" db:"days"`
	StartTime      string    `json:"start_time" db:"start_time"`
	EndTime        string    `json:"end_time" db:"end_time"`
	LessonDuration int       `json:"lesson_duration" db:"lesson_duration"`
	Breaks         []Break   `json:"breaks" db:"breaks"`
	IsDefault      bool      `json:"is_default" db:"is_default"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Break represents a break period in the timetable
type Break struct {
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// TimetableEntry represents a single lesson in the timetable
type TimetableEntry struct {
	ID          string    `json:"id" db:"id"`
	ClassID     string    `json:"class_id" db:"class_id"`
	SubjectID   string    `json:"subject_id" db:"subject_id"`
	TeacherID   string    `json:"teacher_id" db:"teacher_id"`
	Day         string    `json:"day" db:"day"`
	TimeSlot    string    `json:"time_slot" db:"time_slot"`
	PaperID     *string   `json:"paper_id" db:"paper_id"`
	SubjectName string    `json:"subject_name"`
	ClassName   string    `json:"class_name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
