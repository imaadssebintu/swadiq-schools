package models

import "time"

// Exam represents an exam event for a class
type Exam struct {
	ID               string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name             string          `json:"name" gorm:"not null" validate:"required"`
	ClassID          string          `json:"class_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	AcademicYearID   *string         `json:"academic_year_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	TermID           *string         `json:"term_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	PaperID          string          `json:"paper_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	AssessmentTypeID *string         `json:"assessment_type_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	Type             string          `json:"type" gorm:"not null;default:'exam'" validate:"required"`
	StartTime        time.Time       `json:"start_time" gorm:"not null" validate:"required"`
	EndTime          time.Time       `json:"end_time" gorm:"not null" validate:"required"`
	IsActive         bool            `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
	Class            *Class          `json:"class,omitempty" gorm:"foreignKey:ClassID;references:ID"`
	AcademicYear     *AcademicYear   `json:"academic_year,omitempty" gorm:"foreignKey:AcademicYearID;references:ID"`
	Term             *Term           `json:"term,omitempty" gorm:"foreignKey:TermID;references:ID"`
	Paper            *Paper          `json:"paper,omitempty" gorm:"foreignKey:PaperID;references:ID"`
	AssessmentType   *AssessmentType `json:"assessment_type,omitempty" gorm:"foreignKey:AssessmentTypeID;references:ID"`
}
