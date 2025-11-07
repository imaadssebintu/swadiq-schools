package models

// TeacherSubject represents the many-to-many relationship between teachers and subjects
type TeacherSubject struct {
	TeacherID string   `json:"teacher_id" gorm:"primaryKey;not null;type:uuid"`
	SubjectID string   `json:"subject_id" gorm:"primaryKey;not null;type:uuid"`
	Teacher   *User    `json:"teacher,omitempty" gorm:"foreignKey:TeacherID;references:ID"`
	Subject   *Subject `json:"subject,omitempty" gorm:"foreignKey:SubjectID;references:ID"`
}