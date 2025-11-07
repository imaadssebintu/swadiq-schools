package models

import "time"

// FeeTypeAssignment links a fee type to a class or individual students.
type FeeTypeAssignment struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FeeTypeID string     `json:"fee_type_id" gorm:"not null;index"`
	StudentID *string    `json:"student_id,omitempty" gorm:"index"`
	ClassID   *string    `json:"class_id,omitempty" gorm:"index"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	FeeType *FeeType `json:"fee_type,omitempty" gorm:"foreignKey:FeeTypeID;references:ID"`
	Student *Student `json:"student,omitempty" gorm:"foreignKey:StudentID;references:ID"`
	Class   *Class   `json:"class,omitempty" gorm:"foreignKey:ClassID;references:ID"`
}
