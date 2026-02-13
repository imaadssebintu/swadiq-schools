package models

import "time"

// Event represents a calendar event
type Event struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	Type              string    `json:"type"` // Keep for backward compatibility or display
	CategoryID        string    `json:"category_id"`
	CategoryName      string    `json:"category_name"`
	TermID            string    `json:"term_id"`
	TermName          string    `json:"term_name"`
	Location          string    `json:"location"`
	Color             string    `json:"color"`               // Populated from category
	SuspensionType    string    `json:"suspension_type"`     // NONE, ALL, SPECIFIC
	SuspendedClassIDs []string  `json:"suspended_class_ids"` // List of IDs if SPECIFIC
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
