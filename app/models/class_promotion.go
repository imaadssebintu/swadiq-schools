package models

import "time"

// ClassPromotion represents the promotion settings for a class
type ClassPromotion struct {
	ID                string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	FromClassID       string     `json:"from_class_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	ToClassID         string     `json:"to_class_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	AcademicYearID    *string    `json:"academic_year_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	PromotionCriteria string     `json:"promotion_criteria" gorm:"type:text"` // JSON string with criteria
	IsActive          bool       `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	
	// Relationships
	FromClass     *Class        `json:"from_class,omitempty" gorm:"foreignKey:FromClassID;references:ID"`
	ToClass       *Class        `json:"to_class,omitempty" gorm:"foreignKey:ToClassID;references:ID"`
	AcademicYear  *AcademicYear `json:"academic_year,omitempty" gorm:"foreignKey:AcademicYearID;references:ID"`
}

// PromotionCriteria represents the criteria for promotion
type PromotionCriteria struct {
	MinimumAttendance float64 `json:"minimum_attendance"` // Percentage
	MinimumGrade      string  `json:"minimum_grade"`      // Grade like "C", "B", etc.
	RequiredSubjects  []string `json:"required_subjects"`  // Subject IDs that must be passed
	AutoPromote       bool    `json:"auto_promote"`       // Whether to auto-promote students
}
