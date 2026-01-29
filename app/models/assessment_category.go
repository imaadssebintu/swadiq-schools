package models

import "time"

// AssessmentCategory represents a major group of assessments (e.g., Exam, Test, Project)
type AssessmentCategory struct {
	ID        string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name      string            `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	Code      string            `json:"code" gorm:"uniqueIndex;not null" validate:"required"`
	Color     string            `json:"color" gorm:"default:'indigo'"`
	IsActive  bool              `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
	Types     []*AssessmentType `json:"types,omitempty" gorm:"foreignKey:CategoryID"`
}
