package models

import "time"

// Result stores a student's marks for a paper in an exam
type Result struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	ExamID    string     `json:"exam_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	StudentID string     `json:"student_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	PaperID   string     `json:"paper_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	TermID    *string    `json:"term_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	Marks     float64    `json:"marks" gorm:"not null;type:decimal(5,2)" validate:"gte=0"`
	GradeID   *string    `json:"grade_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Grade     *Grade     `json:"grade,omitempty" gorm:"foreignKey:GradeID;references:ID"` // optional for JSON responses
	Exam      *Exam      `json:"exam,omitempty" gorm:"foreignKey:ExamID;references:ID"`
	Student   *Student   `json:"student,omitempty" gorm:"foreignKey:StudentID;references:ID"`
	Paper     *Paper     `json:"paper,omitempty" gorm:"foreignKey:PaperID;references:ID"`
}
