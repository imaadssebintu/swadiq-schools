package models

import "time"

// TimetableEntryResponse extends the base TimetableEntry with additional details for display
// such as the teacher's full name.
type TimetableEntryResponse struct {
	TimetableEntry
	TeacherName  string `json:"teacher_name"`
	StudentCount int    `json:"student_count"`
	ClassName    string `json:"class_name"`
	SubjectName  string `json:"subject_name"`
	PaperCode    string `json:"paper_code"`
}

type DashboardStats struct {
	TotalStudents     int        `json:"total_students"`
	TotalTeachers     int        `json:"total_teachers"`
	TotalClasses      int        `json:"total_classes"`
	MonthlyRevenue    float64    `json:"monthly_revenue"`
	StudentAttendance float64    `json:"student_attendance"`
	TeacherAttendance float64    `json:"teacher_attendance"`
	FeeCollectionRate float64    `json:"fee_collection_rate"`
	RecentActivities  []Activity `json:"recent_activities"`
}

type Activity struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TimeAgo     string    `json:"time_ago"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	RawTime     time.Time `json:"-"`
}

// Event represents a calendar event
type Event struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Type         string    `json:"type"` // Keep for backward compatibility or display
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	TermID       string    `json:"term_id"`
	TermName     string    `json:"term_name"`
	Location     string    `json:"location"`
	Color        string    `json:"color"` // Populated from category
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
