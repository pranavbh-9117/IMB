// Package repository implements aggregation queries for the dashboard module.
package repository

import (
	"context"
	"fmt"
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
