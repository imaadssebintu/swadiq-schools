package models

import "time"

// Grade represents a grading rule, e.g., A, B, C
type Grade struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name      string     `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	MinMarks  float64    `json:"min_marks" gorm:"not null;type:decimal(5,2)" validate:"gte=0"`
	MaxMarks  float64    `json:"max_marks" gorm:"not null;type:decimal(5,2)" validate:"gte=0"`
	GradeValue float64    `json:"grade_value" gorm:"default:0;type:decimal(5,2)" validate:"gte=0"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Results   []*Result  `json:"results,omitempty" gorm:"foreignKey:GradeID;references:ID"`
}
