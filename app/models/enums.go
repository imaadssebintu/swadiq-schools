package models

// AttendanceStatus defines the possible status values for attendance.
type AttendanceStatus string

const (
	Present AttendanceStatus = "present"
	Absent  AttendanceStatus = "absent"
	Late    AttendanceStatus = "late"
	Excused AttendanceStatus = "excused"
)

// RecipientType defines the possible recipient types for notifications.
type RecipientType string

const (
	StudentRecipient RecipientType = "student"
	ParentRecipient  RecipientType = "parent"
	TeacherRecipient RecipientType = "teacher"
)

// DayOfWeek defines the days of the week for schedules.
type DayOfWeek string

const (
	Monday    DayOfWeek = "monday"
	Tuesday   DayOfWeek = "tuesday"
	Wednesday DayOfWeek = "wednesday"
	Thursday  DayOfWeek = "thursday"
	Friday    DayOfWeek = "friday"
	Saturday  DayOfWeek = "saturday"
	Sunday    DayOfWeek = "sunday"
)

// Gender defines the possible gender values for a student.
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
	Other  Gender = "other"
)

// RelationshipType defines the relationship of a parent/guardian to a student.
type RelationshipType string

const (
	Father   RelationshipType = "father"
	Mother   RelationshipType = "mother"
	Guardian RelationshipType = "guardian"
	Brother  RelationshipType = "brother"
	Sister   RelationshipType = "sister"
	OtherRel RelationshipType = "other"
)

// PaymentStatus defines the status of a payment
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

// SalaryPeriod defines the period for a teacher's salary.
type SalaryPeriod string

const (
	SalaryDay   SalaryPeriod = "day"
	SalaryWeek  SalaryPeriod = "week"
	SalaryMonth SalaryPeriod = "month"
)
