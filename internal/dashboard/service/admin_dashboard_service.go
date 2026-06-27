// Package service implements dashboard business logic and aggregations.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/dashboard/dto"
	"github.com/pranavbh-9117/IMB/internal/dashboard/repository"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/workerpool"
)

// AdminDashboardService defines the contract for fetching admin dashboard metrics.
type AdminDashboardService interface {
	GetAdminDashboard(ctx context.Context, institutionID uuid.UUID) (*dto.AdminDashboardData, error)
}

type adminDashboardService struct {
	repo repository.DashboardRepository
	pool workerpool.Pool
}

// NewAdminDashboardService initializes a stateless AdminDashboardService.
func NewAdminDashboardService(repo repository.DashboardRepository, pool workerpool.Pool) AdminDashboardService {
	return &adminDashboardService{
		repo: repo,
		pool: pool,
	}
}

// resultHolder encapsulates request-scoped aggregation metrics with thread-safe access.
type resultHolder struct {
	mu            sync.Mutex
	totalStudents int64
	totalFaculty  int64
	quizCount     int64
	leaveStats    dto.LeaveStatistics
}

func (s *adminDashboardService) GetAdminDashboard(ctx context.Context, institutionID uuid.UUID) (*dto.AdminDashboardData, error) {
	holder := &resultHolder{}
	group := s.pool.NewGroup(ctx)

	// Task 1: Count Students
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		count, err := s.repo.CountUsersByRole(jobCtx, institutionID, domain.RoleStudent)
		if err != nil {
			return fmt.Errorf("fetch students count: %w", err)
		}
		holder.mu.Lock()
		holder.totalStudents = count
		holder.mu.Unlock()
		return nil
	}))

	// Task 2: Count Faculty
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		count, err := s.repo.CountUsersByRole(jobCtx, institutionID, domain.RoleFaculty)
		if err != nil {
			return fmt.Errorf("fetch faculty count: %w", err)
		}
		holder.mu.Lock()
		holder.totalFaculty = count
		holder.mu.Unlock()
		return nil
	}))

	// Task 3: Count Quizzes
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		count, err := s.repo.CountQuizzes(jobCtx, institutionID)
		if err != nil {
			return fmt.Errorf("fetch quizzes count: %w", err)
		}
		holder.mu.Lock()
		holder.quizCount = count
		holder.mu.Unlock()
		return nil
	}))

	// Task 4: Leave Stats
	_ = group.Submit(workerpool.JobFunc(func(jobCtx context.Context) error {
		stats, err := s.repo.GetLeaveStats(jobCtx, institutionID)
		if err != nil {
			return fmt.Errorf("fetch leave stats: %w", err)
		}
		holder.mu.Lock()
		holder.leaveStats = stats
		holder.mu.Unlock()
		return nil
	}))

	// Wait for atomic all-or-nothing completion
	if err := group.Wait(); err != nil {
		return nil, fmt.Errorf("admin dashboard aggregation failed: %w", err)
	}

	return &dto.AdminDashboardData{
		TotalStudents: holder.totalStudents,
		TotalFaculty:  holder.totalFaculty,
		QuizCount:     holder.quizCount,
		LeaveStats:    holder.leaveStats,
	}, nil
}
