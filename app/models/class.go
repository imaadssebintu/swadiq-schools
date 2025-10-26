package models

import "time"

type Class struct {
	ID           string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name         string     `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	Code         string     `json:"code" gorm:"uniqueIndex;not null" validate:"required"`
	TeacherID    *string    `json:"teacher_id,omitempty" gorm:"index;type:uuid" validate:"omitempty,uuid"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	StudentCount int        `json:"student_count" gorm:"-"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Teacher      *User      `json:"teacher,omitempty" gorm:"foreignKey:TeacherID;references:ID"`
	Subjects     []*Subject `json:"subjects,omitempty" gorm:"many2many:class_subjects;"`
	Students     []*Student `json:"students,omitempty" gorm:"foreignKey:ClassID;references:ID"`
}
