// Package service implements dashboard business logic and aggregations.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/dashboard/dto"
	"github.com/pranavbh-9117/IMB/internal/dashboard/repository"
	"github.com/pranavbh-9117/IMB/internal/workerpool"
	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// FacultyDashboardService defines the contract for fetching faculty dashboard metrics.
type FacultyDashboardService interface {
	GetFacultyDashboard(ctx context.Context, facultyID, institutionID uuid.UUID) (*dto.FacultyDashboardData, error)
}

type facultyDashboardService struct {
	repo repository.DashboardRepository
	pool workerpool.Pool
}

// NewFacultyDashboardService initializes a FacultyDashboardService.
func NewFacultyDashboardService(repo repository.DashboardRepository, pool workerpool.Pool) FacultyDashboardService {
	return &facultyDashboardService{
		repo: repo,
		pool: pool,
	}
}

type facultyResultHolder struct {
	mu                 sync.Mutex
	quizAnalytics      *dto.QuizAnalyticsDTO
	studentPerformance []dto.StudentPerformanceDTO
	leaveStats         *dto.FacultyLeaveStatsDTO
	recentActivities   []dto.RecentActivityDTO
}

func (s *facultyDashboardService) GetFacultyDashboard(ctx context.Context, facultyID, institutionID uuid.UUID) (*dto.FacultyDashboardData, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	holder := &facultyResultHolder{}
	group := s.pool.NewGroup(ctx)

	// Task 1: Quiz Analytics
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		start := time.Now()
		res, err := s.repo.GetQuizAnalyticsForFaculty(jobCtx, facultyID, institutionID)
		if err != nil {
			return fmt.Errorf("fetch quiz analytics: %w", err)
		}
		logger.Info(jobCtx, "dashboard: task completed", "task", "quiz_analytics", "duration_ms", time.Since(start).Milliseconds())
		holder.mu.Lock()
		holder.quizAnalytics = res
		holder.mu.Unlock()
		return nil
	}))

	// Task 2: Student Performance
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		start := time.Now()
		res, err := s.repo.GetStudentPerformanceForFaculty(jobCtx, facultyID, institutionID, 10)
		if err != nil {
			return fmt.Errorf("fetch student performance: %w", err)
		}
		logger.Info(jobCtx, "dashboard: task completed", "task", "student_performance", "duration_ms", time.Since(start).Milliseconds())
		holder.mu.Lock()
		holder.studentPerformance = res
		holder.mu.Unlock()
		return nil
	}))

	// Task 3: Leave Statistics
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		start := time.Now()
		res, err := s.repo.GetLeaveStatsForFaculty(jobCtx, institutionID)
		if err != nil {
			return fmt.Errorf("fetch leave stats: %w", err)
		}
		logger.Info(jobCtx, "dashboard: task completed", "task", "leave_statistics", "duration_ms", time.Since(start).Milliseconds())
		holder.mu.Lock()
		holder.leaveStats = res
		holder.mu.Unlock()
		return nil
	}))

	// Task 4: Recent Activities
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		start := time.Now()
		res, err := s.repo.GetRecentActivitiesForFaculty(jobCtx, facultyID, institutionID, 10)
		if err != nil {
			return fmt.Errorf("fetch recent activities: %w", err)
		}
		logger.Info(jobCtx, "dashboard: task completed", "task", "recent_activities", "duration_ms", time.Since(start).Milliseconds())
		holder.mu.Lock()
		holder.recentActivities = res
		holder.mu.Unlock()
		return nil
	}))

	if err := group.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("dashboard: timeout fetching data")
		}
		return nil, fmt.Errorf("faculty dashboard aggregation failed: %w", err)
	}

	return &dto.FacultyDashboardData{
		QuizAnalytics:      *holder.quizAnalytics,
		StudentPerformance: holder.studentPerformance,
		LeaveStatistics:    *holder.leaveStats,
		RecentActivities:   holder.recentActivities,
	}, nil
}
