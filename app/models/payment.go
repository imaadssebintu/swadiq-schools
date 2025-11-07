package models

import "time"

// Payment represents a payment made by a student for one or more fees.
type Payment struct {
	ID              string               `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	StudentID       string               `json:"student_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	TotalAmount     float64              `json:"total_amount" gorm:"not null;type:decimal(10,2)" validate:"required,gt=0"`
	PaymentDate     time.Time            `json:"payment_date" gorm:"not null;index" validate:"required"`
	PaymentMethod   string               `json:"payment_method" gorm:"type:varchar(50)" validate:"required"`
	PaidBy          string               `json:"paid_by" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	TransactionID   *string              `json:"transaction_id,omitempty" gorm:"index"`
	Status          PaymentStatus        `json:"status" gorm:"not null;default:'pending';index;type:varchar(20)" validate:"required"`
	PaidAt          *time.Time           `json:"paid_at,omitempty" gorm:"index"`
	CreatedAt       time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time           `json:"deleted_at,omitempty" gorm:"index"`

	Student         *Student             `json:"student,omitempty" gorm:"foreignKey:StudentID;references:ID"`
	ProcessedByUser *User                `json:"processed_by_user,omitempty" gorm:"foreignKey:PaidBy;references:ID"`
	Allocations     []*PaymentAllocation `json:"allocations,omitempty" gorm:"foreignKey:PaymentID;references:ID"`
}
