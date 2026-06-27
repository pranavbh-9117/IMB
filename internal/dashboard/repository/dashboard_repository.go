// Package repository implements aggregation queries for the dashboard module.
package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/dashboard/dto"
	"github.com/pranavbh-9117/IMB/internal/domain"
)

// DashboardRepository defines aggregation-only data access for dashboards.
type DashboardRepository interface {
	CountUsersByRole(ctx context.Context, institutionID uuid.UUID, role domain.Role) (int64, error)
	CountQuizzes(ctx context.Context, institutionID uuid.UUID) (int64, error)
	GetLeaveStats(ctx context.Context, institutionID uuid.UUID) (dto.LeaveStatistics, error)
	GetQuizAnalyticsForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID) (*dto.QuizAnalyticsDTO, error)
	GetStudentPerformanceForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID, limit int) ([]dto.StudentPerformanceDTO, error)
	GetLeaveStatsForFaculty(ctx context.Context, institutionID uuid.UUID) (*dto.FacultyLeaveStatsDTO, error)
	GetRecentActivitiesForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID, limit int) ([]dto.RecentActivityDTO, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository initializes an aggregation-only DashboardRepository.
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) CountUsersByRole(ctx context.Context, institutionID uuid.UUID, role domain.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("institution_id = ? AND role = ? AND is_active = ?", institutionID, role, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("dashboard repository: count users by role: %w", err)
	}
	return count, nil
}

func (r *dashboardRepository) CountQuizzes(ctx context.Context, institutionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Quiz{}).
		Where("institution_id = ?", institutionID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("dashboard repository: count quizzes: %w", err)
	}
	return count, nil
}

func (r *dashboardRepository) GetLeaveStats(ctx context.Context, institutionID uuid.UUID) (dto.LeaveStatistics, error) {
	var stats dto.LeaveStatistics

	// Pending
	err := r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ?", institutionID, domain.LeaveStatusPending).
		Count(&stats.Pending).Error
	if err != nil {
		return stats, fmt.Errorf("dashboard repository: count pending leaves: %w", err)
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Approved this month
	err = r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ? AND created_at >= ?", institutionID, domain.LeaveStatusApproved, startOfMonth).
		Count(&stats.ApprovedThisMonth).Error
	if err != nil {
		return stats, fmt.Errorf("dashboard repository: count approved leaves: %w", err)
	}

	// Rejected this month
	err = r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ? AND created_at >= ?", institutionID, domain.LeaveStatusRejected, startOfMonth).
		Count(&stats.RejectedThisMonth).Error
	if err != nil {
		return stats, fmt.Errorf("dashboard repository: count rejected leaves: %w", err)
	}

	return stats, nil
}

func (r *dashboardRepository) GetQuizAnalyticsForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID) (*dto.QuizAnalyticsDTO, error) {
	var created int64
	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).
		Where("institution_id = ? AND created_by = ?", institutionID, facultyID).
		Count(&created).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: count quizzes created: %w", err)
	}

	var published int64
	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).
		Where("institution_id = ? AND created_by = ? AND is_published = ?", institutionID, facultyID, true).
		Count(&published).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: count quizzes published: %w", err)
	}

	var attemptStats struct {
		Attempts int64
		AvgScore *float64
	}
	query := `
		SELECT COUNT(a.id) as attempts, 
		       AVG(CASE WHEN a.total_marks > 0 THEN (CAST(a.score AS FLOAT) / CAST(a.total_marks AS FLOAT)) * 100.0 ELSE 0.0 END) as avg_score
		FROM quiz_attempts a
		JOIN quizzes q ON a.quiz_id = q.id
		WHERE q.institution_id = ? AND q.created_by = ? AND q.deleted_at IS NULL
	`
	if err := r.db.WithContext(ctx).Raw(query, institutionID, facultyID).Scan(&attemptStats).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: get quiz attempts stats: %w", err)
	}

	avg := 0.0
	if attemptStats.AvgScore != nil {
		avg = *attemptStats.AvgScore
	}

	return &dto.QuizAnalyticsDTO{
		TotalQuizzesCreated:   created,
		TotalPublished:        published,
		TotalAttemptsReceived: attemptStats.Attempts,
		AvgScorePercentage:    avg,
	}, nil
}

func (r *dashboardRepository) GetStudentPerformanceForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID, limit int) ([]dto.StudentPerformanceDTO, error) {
	var results []dto.StudentPerformanceDTO
	query := `
		SELECT 
			u.id as student_id,
			u.name as name,
			COUNT(a.id) as total_attempts,
			AVG(CASE WHEN a.total_marks > 0 THEN (CAST(a.score AS FLOAT) / CAST(a.total_marks AS FLOAT)) * 100.0 ELSE 0.0 END) as avg_score_percentage
		FROM quiz_attempts a
		JOIN quizzes q ON a.quiz_id = q.id
		JOIN users u ON a.student_id = u.id
		WHERE q.institution_id = ? AND q.created_by = ? AND q.deleted_at IS NULL
		GROUP BY u.id, u.name
		ORDER BY avg_score_percentage DESC
		LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(query, institutionID, facultyID, limit).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: get student performance: %w", err)
	}
	if results == nil {
		results = []dto.StudentPerformanceDTO{}
	}
	return results, nil
}

func (r *dashboardRepository) GetLeaveStatsForFaculty(ctx context.Context, institutionID uuid.UUID) (*dto.FacultyLeaveStatsDTO, error) {
	var stats dto.FacultyLeaveStatsDTO

	err := r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Joins("JOIN users ON leave_requests.user_id = users.id").
		Where("leave_requests.institution_id = ? AND leave_requests.status = ? AND users.role = ?", institutionID, domain.LeaveStatusPending, domain.RoleStudent).
		Count(&stats.PendingApprovals).Error
	if err != nil {
		return nil, fmt.Errorf("dashboard repository: count faculty pending leaves: %w", err)
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err = r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Joins("JOIN users ON leave_requests.user_id = users.id").
		Where("leave_requests.institution_id = ? AND leave_requests.status = ? AND users.role = ? AND leave_requests.created_at >= ?", institutionID, domain.LeaveStatusApproved, domain.RoleStudent, startOfMonth).
		Count(&stats.ApprovedThisMonth).Error
	if err != nil {
		return nil, fmt.Errorf("dashboard repository: count faculty approved leaves: %w", err)
	}

	err = r.db.WithContext(ctx).
		Model(&domain.LeaveRequest{}).
		Joins("JOIN users ON leave_requests.user_id = users.id").
		Where("leave_requests.institution_id = ? AND leave_requests.status = ? AND users.role = ? AND leave_requests.created_at >= ?", institutionID, domain.LeaveStatusRejected, domain.RoleStudent, startOfMonth).
		Count(&stats.RejectedThisMonth).Error
	if err != nil {
		return nil, fmt.Errorf("dashboard repository: count faculty rejected leaves: %w", err)
	}

	return &stats, nil
}

func (r *dashboardRepository) GetRecentActivitiesForFaculty(ctx context.Context, facultyID, institutionID uuid.UUID, limit int) ([]dto.RecentActivityDTO, error) {
	type quizActivityRow struct {
		StudentName string
		QuizTitle   string
		Timestamp   time.Time
	}
	var quizRows []quizActivityRow
	queryQuiz := `
		SELECT u.name as student_name, q.title as quiz_title, a.started_at as timestamp
		FROM quiz_attempts a
		JOIN quizzes q ON a.quiz_id = q.id
		JOIN users u ON a.student_id = u.id
		WHERE q.institution_id = ? AND q.created_by = ? AND q.deleted_at IS NULL
		ORDER BY a.started_at DESC
		LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(queryQuiz, institutionID, facultyID, limit).Scan(&quizRows).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: get recent quiz activities: %w", err)
	}

	type leaveActivityRow struct {
		StudentName string
		Timestamp   time.Time
	}
	var leaveRows []leaveActivityRow
	queryLeave := `
		SELECT u.name as student_name, lr.created_at as timestamp
		FROM leave_requests lr
		JOIN users u ON lr.user_id = u.id
		WHERE lr.institution_id = ? AND u.role = ?
		ORDER BY lr.created_at DESC
		LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(queryLeave, institutionID, domain.RoleStudent, limit).Scan(&leaveRows).Error; err != nil {
		return nil, fmt.Errorf("dashboard repository: get recent leave activities: %w", err)
	}

	var activities []dto.RecentActivityDTO
	for _, qr := range quizRows {
		activities = append(activities, dto.RecentActivityDTO{
			Type:        "quiz_attempt",
			Description: fmt.Sprintf("Student %s submitted '%s'", qr.StudentName, qr.QuizTitle),
			Timestamp:   qr.Timestamp,
		})
	}
	for _, lr := range leaveRows {
		activities = append(activities, dto.RecentActivityDTO{
			Type:        "leave_request",
			Description: fmt.Sprintf("Student %s applied for leave", lr.StudentName),
			Timestamp:   lr.Timestamp,
		})
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	if len(activities) > limit {
		activities = activities[:limit]
	}
	if activities == nil {
		activities = []dto.RecentActivityDTO{}
	}

	return activities, nil
}
