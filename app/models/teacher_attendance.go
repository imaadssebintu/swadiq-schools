package models

import "time"

// TeacherAttendance represents a teacher's attendance record for a specific date
// This table is used to calculate "Duty Days" for allowance payouts.
type TeacherAttendance struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	TeacherID string    `json:"teacher_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	Date      time.Time `json:"date" gorm:"not null;index;type:date" validate:"required"`
	Status    string    `json:"status" gorm:"not null;type:varchar(20)" validate:"required"` // 'present', 'absent', 'leave'
	Remarks   string    `json:"remarks" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Teacher *User `json:"teacher,omitempty" gorm:"foreignKey:TeacherID;references:ID"`
}
