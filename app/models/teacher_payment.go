package models

import "time"

type PaymentType string

const (
	PaymentTypeBaseSalary PaymentType = "base_salary"
	PaymentTypeAllowance  PaymentType = "allowance"
	PaymentTypeCombined   PaymentType = "combined"
)

// TeacherPayment represents a payment made to a teacher
type TeacherPayment struct {
	ID          string      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	TeacherID   string      `json:"teacher_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	Amount      int64       `json:"amount" gorm:"not null;type:bigint" validate:"required,gt=0"`
	Type        PaymentType `json:"type" gorm:"not null;type:varchar(20)" validate:"required"`
	PeriodStart time.Time   `json:"period_start" gorm:"not null;type:date" validate:"required"`
	PeriodEnd   time.Time   `json:"period_end" gorm:"not null;type:date" validate:"required"`
	PaidAt      time.Time   `json:"paid_at" gorm:"autoCreateTime"`
	Reference   string      `json:"reference" gorm:"type:varchar(100)"` // Check number, Transaction ID, etc.
	Notes       string      `json:"notes" gorm:"type:text"`

	Teacher *User `json:"teacher,omitempty" gorm:"foreignKey:TeacherID;references:ID"`
}
