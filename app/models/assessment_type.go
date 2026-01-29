package models

import "time"

// AssessmentType represents a category of student assessment (e.g., Exam, Test, Project)
type AssessmentType struct {
	ID           string              `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name         string              `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	Code         string              `json:"code" gorm:"uniqueIndex;not null" validate:"required"`
	CategoryID   string              `json:"category_id" gorm:"index;type:uuid;not null"`
	CategoryName string              `json:"category_name" gorm:"-"` // Not stored in DB
	Weight       float64             `json:"weight" gorm:"default:1.0"`
	Color        string              `json:"color" gorm:"default:'indigo'"`
	AllClasses   bool                `json:"all_classes" gorm:"default:true"`
	IsActive     bool                `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time          `json:"deleted_at,omitempty" gorm:"index"`
	Category     *AssessmentCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Classes      []*Class            `json:"classes,omitempty" gorm:"many2many:assessment_type_classes;joinForeignKey:assessment_type_id;joinReferences:class_id"`
}
