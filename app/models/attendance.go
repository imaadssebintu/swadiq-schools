package models

import "time"

// Attendance represents a student's attendance for a class or timetable entry
type Attendance struct {
	ID                string           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	StudentID         string           `json:"student_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	ClassID           *string          `json:"class_id,omitempty" gorm:"index;type:uuid"`
	TimetableEntryID  *string          `json:"timetable_entry_id,omitempty" gorm:"index;type:uuid"`
	PaperID           *string          `json:"paper_id,omitempty" gorm:"index;type:uuid"`
	TermID            *string          `json:"term_id,omitempty" gorm:"index;type:uuid"`
	Date              time.Time        `json:"date" gorm:"not null;index;type:date" validate:"required"`
	Status            AttendanceStatus `json:"status" gorm:"not null;type:varchar(10)" validate:"required,oneof=present absent late excused"`
	MarkedBy          *string          `json:"marked_by,omitempty" gorm:"type:uuid"`
	CreatedAt         time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time       `json:"deleted_at,omitempty" gorm:"index"`
	Student           *Student         `json:"student,omitempty" gorm:"foreignKey:StudentID;references:ID"`
	Class             *Class           `json:"class,omitempty" gorm:"foreignKey:ClassID;references:ID"`
	TimetableEntry    *TimetableEntry  `json:"timetable_entry,omitempty" gorm:"foreignKey:TimetableEntryID;references:ID"`
	Paper             *Paper           `json:"paper,omitempty" gorm:"foreignKey:PaperID;references:ID"`
	Term              *Term            `json:"term,omitempty" gorm:"foreignKey:TermID;references:ID"`
	IsLessonConducted bool             `json:"is_lesson_conducted" gorm:"default:false"`
	MarkedByUser      *User            `json:"marked_by_user,omitempty" gorm:"foreignKey:MarkedBy;references:ID"`
}

// ConductedLesson represents a record that a specific lesson was taught
type ConductedLesson struct {
	ID               string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TimetableEntryID string    `json:"timetable_entry_id" gorm:"not null;index;type:uuid"`
	TermID           *string   `json:"term_id" gorm:"index;type:uuid"`
	Date             time.Time `json:"date" gorm:"not null;index"`
	TeacherID        string    `json:"teacher_id" gorm:"not null;index;type:uuid"`
	Topic            string    `json:"topic"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"default:now()"`
}
