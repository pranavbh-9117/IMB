// Package dto provides request and response payload structures for dashboard.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// QuizAnalyticsDTO represents aggregated quiz metrics for a faculty member.
type QuizAnalyticsDTO struct {
	TotalQuizzesCreated   int64   `json:"total_quizzes_created"`
	TotalPublished        int64   `json:"total_published"`
	TotalAttemptsReceived int64   `json:"total_attempts_received"`
	AvgScorePercentage    float64 `json:"avg_score_percentage"`
}

// StudentPerformanceDTO represents individual student performance metrics on faculty quizzes.
type StudentPerformanceDTO struct {
	StudentID          uuid.UUID `json:"student_id"`
	Name               string    `json:"name"`
	TotalAttempts      int64     `json:"total_attempts"`
	AvgScorePercentage float64   `json:"avg_score_percentage"`
}

// FacultyLeaveStatsDTO represents leave statistics relevant to faculty.
type FacultyLeaveStatsDTO struct {
	PendingApprovals  int64 `json:"pending_approvals"`
	ApprovedThisMonth int64 `json:"approved_this_month"`
	RejectedThisMonth int64 `json:"rejected_this_month"`
}

// RecentActivityDTO represents a chronological activity entry (quiz attempt or leave request).
type RecentActivityDTO struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// FacultyDashboardData encapsulates all four data categories returned in the faculty dashboard payload.
type FacultyDashboardData struct {
	QuizAnalytics      QuizAnalyticsDTO        `json:"quiz_analytics"`
	StudentPerformance []StudentPerformanceDTO `json:"student_performance"`
	LeaveStatistics    FacultyLeaveStatsDTO    `json:"leave_statistics"`
	RecentActivities   []RecentActivityDTO     `json:"recent_activities"`
}
