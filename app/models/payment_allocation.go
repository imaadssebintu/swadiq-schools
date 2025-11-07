package models

import "time"

// PaymentAllocation represents the allocation of a payment to specific fees.
type PaymentAllocation struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	PaymentID   string     `json:"payment_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	FeeID       string     `json:"fee_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	FeeTypeID   string     `json:"fee_type_id" gorm:"not null;index;type:uuid" validate:"required,uuid"`
	Amount      float64    `json:"amount" gorm:"not null;type:decimal(10,2)" validate:"required,gt=0"`
	Balance     float64    `json:"balance" gorm:"type:decimal(10,2);default:0" validate:"gte=0"`
	IsFullyPaid bool       `json:"is_fully_paid" gorm:"default:false;index"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	Payment *Payment `json:"payment,omitempty" gorm:"foreignKey:PaymentID;references:ID"`
	Fee     *Fee     `json:"fee,omitempty" gorm:"foreignKey:FeeID;references:ID"`
	FeeType *FeeType `json:"fee_type,omitempty" gorm:"foreignKey:FeeTypeID;references:ID"`
}
