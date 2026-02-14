package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// CustomTime allows parsing dates in YYYY-MM-DD format
type CustomTime struct {
	time.Time
}

// UnmarshalJSON parses dates in YYYY-MM-DD format
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Handle null or empty
	s := string(data)
	if s == "null" || s == "" || s == `""` {
		ct.Time = time.Time{}
		return nil
	}

	// Remove quotes
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Parse the date
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	ct.Time = t
	return nil
}

// MarshalJSON formats dates in YYYY-MM-DD format
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, ct.Time.Format("2006-01-02"))), nil
}

// Scan implements the Scanner interface for database reading
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	if t, ok := value.(time.Time); ok {
		ct.Time = t
		return nil
	}

	return fmt.Errorf("cannot scan %T into CustomTime", value)
}

// Value implements the Valuer interface for database writing
func (ct CustomTime) Value() (driver.Value, error) {
	return ct.Time, nil
}

// AcademicYear represents an academic year/term in the school
type AcademicYear struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()" validate:"required,uuid"`
	Name      string     `json:"name" gorm:"uniqueIndex;not null" validate:"required"`
	StartDate CustomTime `json:"start_date" gorm:"not null;index" validate:"required"`
	EndDate   CustomTime `json:"end_date" gorm:"not null;index" validate:"required"`
	IsCurrent bool       `json:"is_current" gorm:"default:false;index"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Terms     []*Term    `json:"terms" gorm:"foreignKey:AcademicYearID;references:ID"`
}

// IsCurrent checks if the academic year is current based on today's date
func (ay *AcademicYear) IsCurrentByDate() bool {
	now := time.Now()
	return now.After(ay.StartDate.Time) && now.Before(ay.EndDate.Time)
}
