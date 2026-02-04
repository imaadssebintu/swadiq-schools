package models

import "time"

// PaperWeight represents the weight of a paper for a specific class, subject, and term.
type PaperWeight struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	ClassID   string    `json:"class_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	SubjectID string    `json:"subject_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	PaperID   string    `json:"paper_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	TermID    string    `json:"term_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	Weight    int       `json:"weight" gorm:"not null;type:integer" validate:"required,min=0,max=100"` // Percentage, e.g., 40
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Class   *Class   `json:"class,omitempty" gorm:"foreignKey:ClassID;references:ID"`
	Subject *Subject `json:"subject,omitempty" gorm:"foreignKey:SubjectID;references:ID"`
	Paper   *Paper   `json:"paper,omitempty" gorm:"foreignKey:PaperID;references:ID"`
	Term    *Term    `json:"term,omitempty" gorm:"foreignKey:TermID;references:ID"`
}
