package models

import "time"

// Term represents a term/semester within an academic year
type Term struct {
	ID             string        `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AcademicYearID string        `json:"academic_year_id" gorm:"not null;index;type:uuid"`
	Name           string        `json:"name" gorm:"not null"`
	StartDate      CustomTime    `json:"start_date" gorm:"not null;type:date"`
	EndDate        CustomTime    `json:"end_date" gorm:"not null;type:date"`
	IsCurrent      bool          `json:"is_current" gorm:"default:false"`
	IsActive       bool          `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time     `json:"created_at" gorm:"default:now()"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"default:now()"`
	DeletedAt      *time.Time    `json:"deleted_at,omitempty" gorm:"index"`
	AcademicYear   *AcademicYear `json:"academic_year,omitempty" gorm:"foreignKey:AcademicYearID;references:ID"`
}

// IsCurrentByDate checks if the term is current based on today's date
func (t *Term) IsCurrentByDate() bool {
	now := time.Now()
	return now.After(t.StartDate.Time) && now.Before(t.EndDate.Time)
}
