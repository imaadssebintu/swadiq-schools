package models

import "time"

// ClassSubject represents the relationship between a class and a subject, forming a many-to-many join table.
type ClassSubject struct {
	ID           string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	ClassID      string     `json:"class_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	SubjectID    string     `json:"subject_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	IsCompulsory bool       `json:"is_compulsory" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Class        *Class     `json:"class,omitempty" gorm:"foreignKey:ClassID;references:ID"`
	Subject      *Subject   `json:"subject,omitempty" gorm:"foreignKey:SubjectID;references:ID"`
}