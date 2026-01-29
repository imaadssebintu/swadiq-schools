package models

import "time"

// AssessmentType represents a category of student assessment (e.g., Exam, Test, Project)
type AssessmentType struct {
	ID        string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name      string            `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	Code      string            `json:"code" gorm:"uniqueIndex;not null" validate:"required"`
	ParentID  *string           `json:"parent_id,omitempty" gorm:"index;type:uuid"`
	Category  string            `json:"category" gorm:"default:'other'"`
	Weight    float64           `json:"weight" gorm:"default:1.0"`
	Color     string            `json:"color" gorm:"default:'indigo'"`
	IsActive  bool              `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
	Children  []*AssessmentType `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}
