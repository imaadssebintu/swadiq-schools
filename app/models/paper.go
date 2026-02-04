package models

import "time"

type Paper struct {
	ID           string        `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	SubjectID    string        `json:"subject_id" gorm:"not null;index;type:uuid"`
	Name         string        `json:"name" gorm:"not null"`
	Code         string        `json:"code" gorm:"not null"`
	IsCompulsory bool          `json:"is_compulsory" gorm:"default:true"`
	IsActive     bool          `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time     `json:"created_at" gorm:"default:now()"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"default:now()"`
	DeletedAt    *time.Time    `json:"deleted_at,omitempty" gorm:"index"`
	Subject      *Subject      `json:"subject,omitempty" gorm:"foreignKey:SubjectID;references:ID"`
	ClassPapers  []*ClassPaper `json:"class_papers,omitempty" gorm:"foreignKey:PaperID;references:ID"`
	Results      []*Result     `json:"results,omitempty" gorm:"foreignKey:PaperID;references:ID"`
}
