package models

import "time"

type TeacherSalary struct {
	ID               string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	UserID           string       `json:"user_id" gorm:"not null;type:uuid;index" validate:"required,uuid"`
	Amount           int64        `json:"amount" gorm:"not null;type:bigint" validate:"required,gt=0"`
	Allowance        int64        `json:"allowance" gorm:"default:0;type:bigint"`
	HasAllowance     bool         `json:"has_allowance" gorm:"default:false;not null"`
	Period           SalaryPeriod `json:"period" gorm:"not null;type:varchar(20)" validate:"required"`
	AllowancePeriod  SalaryPeriod `json:"allowance_period" gorm:"default:'month';type:varchar(20)"`
	AllowanceTrigger string       `json:"allowance_trigger" gorm:"default:'start_of_duty';type:varchar(20)"`
	EffectiveDate    time.Time    `json:"effective_date" gorm:"not null;type:date" validate:"required"`
	CreatedAt        time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time   `json:"deleted_at,omitempty" gorm:"index"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}
