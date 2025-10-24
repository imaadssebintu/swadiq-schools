package models

import "time"

// FeeType represents a type of fee that can be assigned to students
type FeeType struct {
	ID               string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name             string     `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	Code             string     `json:"code" gorm:"uniqueIndex;not null" validate:"required"`
	Description      *string    `json:"description,omitempty" gorm:"type:text"`
	PaymentFrequency string     `json:"payment_frequency" gorm:"not null;check:payment_frequency IN ('once','per_term','per_year','on_demand')" validate:"required,oneof=once per_term per_year on_demand"`
	IsActive         bool       `json:"is_active" gorm:"default:true;index"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Fees             []*Fee     `json:"fees,omitempty" gorm:"foreignKey:FeeTypeID;references:ID"`
}
