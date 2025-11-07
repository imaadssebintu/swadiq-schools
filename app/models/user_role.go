package models

import "time"

type UserRole struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    string     `json:"user_id" gorm:"not null;index;type:uuid"`
	RoleID    string     `json:"role_id" gorm:"not null;index;type:uuid"`
	CreatedAt time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:now()"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Role      *Role      `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID"`
}
