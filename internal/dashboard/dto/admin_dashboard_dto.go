// Package dto provides request and response payload structures for dashboard.
package dto

// LeaveStatistics represents leave approval metrics for the dashboard.
type LeaveStatistics struct {
	Pending           int64 `json:"pending"`
	ApprovedThisMonth int64 `json:"approved_this_month"`
	RejectedThisMonth int64 `json:"rejected_this_month"`
}

// AdminDashboardData encapsulates the aggregated metrics for the admin dashboard.
type AdminDashboardData struct {
	TotalStudents int64           `json:"total_students"`
	TotalFaculty  int64           `json:"total_faculty"`
	QuizCount     int64           `json:"quiz_count"`
	LeaveStats    LeaveStatistics `json:"leave_statistics"`
}
