package models

import "time"

// Fee represents an actual charge for a specific student within an academic year.
type Fee struct {
	ID             string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	StudentID      string     `json:"student_id" gorm:"not null;index;type:uuid"`
	FeeTypeID      string     `json:"fee_type_id" gorm:"not null;index;type:uuid"`
	AcademicYearID *string    `json:"academic_year_id,omitempty" gorm:"index;type:uuid"`
	TermID         *string    `json:"term_id,omitempty" gorm:"index;type:uuid"`
	Title          string     `json:"title" gorm:"not null"`
	Amount         float64    `json:"amount" gorm:"not null;type:numeric"`
	Balance        float64    `json:"balance" gorm:"type:numeric;default:0"`
	Currency       string     `json:"currency" gorm:"not null;default:'USD'"`
	Paid           bool       `json:"paid" gorm:"default:false"`
	DueDate        time.Time  `json:"due_date" gorm:"not null;type:date"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"default:now()"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Student      *Student      `json:"student,omitempty" gorm:"foreignKey:StudentID;references:ID"`
	FeeType      *FeeType      `json:"fee_type,omitempty" gorm:"foreignKey:FeeTypeID;references:ID"`
	AcademicYear *AcademicYear `json:"academic_year,omitempty" gorm:"foreignKey:AcademicYearID;references:ID"`
	Term         *Term         `json:"term,omitempty" gorm:"foreignKey:TermID;references:ID"`
}

// IsFullyPaid returns true if the fee is marked as paid.
func (f *Fee) IsFullyPaid() bool {
	return f.Paid
}

// MarkAsPaid marks the fee as fully paid.
func (f *Fee) MarkAsPaid() {
	f.Balance = 0
	f.Paid = true
	now := time.Now()
	f.PaidAt = &now
}
