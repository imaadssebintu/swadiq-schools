package models

import "time"

// TeacherBaseSalary represents the core compensation configuration
type TeacherBaseSalary struct {
	ID            string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	UserID        string       `json:"user_id" gorm:"not null;type:uuid;index" validate:"required,uuid"`
	Amount        int64        `json:"amount" gorm:"not null;type:bigint" validate:"required,gt=0"`
	Period        SalaryPeriod `json:"period" gorm:"not null;type:varchar(20)" validate:"required"`
	EffectiveDate time.Time    `json:"effective_date" gorm:"not null;type:date;default:CURRENT_DATE"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time   `json:"deleted_at,omitempty" gorm:"index"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TeacherAllowance represents independent stipend/extra compensation
type TeacherAllowance struct {
	ID            string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	UserID        string       `json:"user_id" gorm:"not null;type:uuid;index" validate:"required,uuid"`
	Amount        int64        `json:"amount" gorm:"default:0;type:bigint"`
	Period        SalaryPeriod `json:"period" gorm:"default:'month';type:varchar(20)"`
	IsActive      bool         `json:"is_active" gorm:"default:true;not null"`
	EffectiveDate time.Time    `json:"effective_date" gorm:"not null;type:date;default:CURRENT_DATE"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time   `json:"deleted_at,omitempty" gorm:"index"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TeacherSalary legacy struct for internal calculations
type TeacherSalary struct {
	ID               string       `json:"id"`
	UserID           string       `json:"user_id"`
	Amount           int64        `json:"amount"`
	Allowance        int64        `json:"allowance"`
	HasAllowance     bool         `json:"has_allowance"`
	Period           SalaryPeriod `json:"period"`
	AllowancePeriod  SalaryPeriod `json:"allowance_period"`
	AllowanceTrigger string       `json:"allowance_trigger"`
	EffectiveDate    time.Time    `json:"effective_date"`
	CreatedAt        time.Time    `json:"created_at"`
}
