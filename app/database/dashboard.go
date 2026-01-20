package database

import (
	"database/sql"
	"swadiq-schools/app/models"
)

// GetDashboardStats returns statistics for the admin dashboard
func GetDashboardStats(db *sql.DB) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// 1. Total Students
	err := db.QueryRow("SELECT COUNT(*) FROM students WHERE is_active = true").Scan(&stats.TotalStudents)
	if err != nil {
		return nil, err
	}

	// 2. Total Teachers
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT u.id) 
		FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		JOIN roles r ON ur.role_id = r.id 
		WHERE r.name IN ('admin', 'head_teacher', 'class_teacher', 'subject_teacher') 
		AND u.is_active = true
	`).Scan(&stats.TotalTeachers)
	if err != nil {
		return nil, err
	}

	// 3. Total Active Classes
	err = db.QueryRow("SELECT COUNT(*) FROM classes WHERE is_active = true").Scan(&stats.TotalClasses)
	if err != nil {
		return nil, err
	}

	// Mock other stats for now
	stats.MonthlyRevenue = 45250.00
	stats.StudentAttendance = 94.5
	stats.TeacherAttendance = 98.2
	stats.FeeCollectionRate = 87.3

	stats.RecentActivities = []models.Activity{
		{
			Type:        "attendance",
			Title:       "Attendance marked for Class 10-A",
			Description: "Mathematics - 28/30 students present",
			TimeAgo:     "15 minutes ago",
			Icon:        "check",
			Color:       "green",
		},
		{
			Type:        "enrollment",
			Title:       "New student enrolled",
			Description: "Sarah Johnson - Grade 9",
			TimeAgo:     "1 hour ago",
			Icon:        "user-plus",
			Color:       "blue",
		},
		{
			Type:        "exam",
			Title:       "Exam results uploaded",
			Description: "Physics Mid-term - Grade 11",
			TimeAgo:     "2 hours ago",
			Icon:        "clipboard-list",
			Color:       "purple",
		},
		{
			Type:        "finance",
			Title:       "Fee payment received",
			Description: "$450 - John Smith (Grade 8)",
			TimeAgo:     "3 hours ago",
			Icon:        "money-bill",
			Color:       "orange",
		},
	}

	return stats, nil
}
