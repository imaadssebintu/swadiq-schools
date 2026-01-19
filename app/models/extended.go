package models

// TimetableEntryResponse extends the base TimetableEntry with additional details for display
// such as the teacher's full name.
type TimetableEntryResponse struct {
	TimetableEntry
	TeacherName  string `json:"teacher_name"`
	StudentCount int    `json:"student_count"`
}
