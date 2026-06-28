// Package repository provides data access for leave requests and balances.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

type txKey struct{}

type leaveRepository struct {
	db *gorm.DB
}


func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db: db}
}

// getDB extracts the transaction session from the infrastructure registry if active.
func (r *leaveRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return database.GetSession(ctx, r.db)
}

// DoInTransaction executes the provided function within a database transaction.
func (r *leaveRepository) DoInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	// If already in a transaction, just execute the function
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// CreateRequest inserts a new leave request into the database.
func (r *leaveRepository) CreateRequest(ctx context.Context, request *domain.LeaveRequest) error {
	if err := r.getDB(ctx).Create(request).Error; err != nil {
		return fmt.Errorf("leave repository: create request: %w", err)
	}
	return nil
}

// GetRequestByID retrieves a specific leave request by its UUID.
func (r *leaveRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error) {
	var request domain.LeaveRequest
	err := r.getDB(ctx).
		Preload("User").
		First(&request, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("leave repository: get request by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("leave repository: get request by id: %w", err)
	}
	return &request, nil
}

// ListRequests retrieves a list of leave requests based on dynamic filters.
func (r *leaveRepository) ListRequests(ctx context.Context, filter RequestFilter, offset, limit int) ([]domain.LeaveRequest, error) {
	var requests []domain.LeaveRequest
	query := r.getDB(ctx).Model(&domain.LeaveRequest{})

	if filter.InstitutionID != nil {
		query = query.Where("institution_id = ?", *filter.InstitutionID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.ReviewerID != nil {
		query = query.Where("reviewed_by = ?", *filter.ReviewerID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("leave repository: list requests: %w", err)
	}
	return requests, nil
}

// UpdateRequestStatus updates the approval status, reviewer ID, and note.
func (r *leaveRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status domain.LeaveStatus, reviewerID *uuid.UUID, note string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      status,
		"reviewed_by": reviewerID,
		"reviewed_at": &now,
		"review_note": note,
	}

	res := r.getDB(ctx).Model(&domain.LeaveRequest{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("leave repository: update request status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("leave repository: update request status: %w", ErrNotFound)
	}

	return nil
}

// CreateBalance creates an initial leave balance record for a user.
func (r *leaveRepository) CreateBalance(ctx context.Context, balance *domain.LeaveBalance) error {
	if err := r.getDB(ctx).Create(balance).Error; err != nil {
		return fmt.Errorf("leave repository: create balance: %w", err)
	}
	return nil
}

// GetBalanceByUserID retrieves the leave balance associated with a user.
func (r *leaveRepository) GetBalanceByUserID(ctx context.Context, userID uuid.UUID) (*domain.LeaveBalance, error) {
	var balance domain.LeaveBalance
	err := r.getDB(ctx).
		First(&balance, "user_id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("leave repository: get balance: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("leave repository: get balance: %w", err)
	}
	return &balance, nil
}

// UpdateBalance persists modifications to a user's total or used leave days.
func (r *leaveRepository) UpdateBalance(ctx context.Context, balance *domain.LeaveBalance) error {
	if err := r.getDB(ctx).Save(balance).Error; err != nil {
		return fmt.Errorf("leave repository: update balance: %w", err)
	}
	return nil
}

func (r *leaveRepository) GetInstitutionLeaveStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) (int, int, int, error) {
	var approved int64
	var rejected int64
	var pending int64

	db := r.getDB(ctx)

	if err := db.Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ? AND reviewed_at >= ? AND reviewed_at < ?", institutionID, domain.LeaveStatusApproved, startTime, endTime).
		Count(&approved).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("leave repository: count approved leaves: %w", err)
	}

	if err := db.Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ? AND reviewed_at >= ? AND reviewed_at < ?", institutionID, domain.LeaveStatusRejected, startTime, endTime).
		Count(&rejected).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("leave repository: count rejected leaves: %w", err)
	}

	if err := db.Model(&domain.LeaveRequest{}).
		Where("institution_id = ? AND status = ? AND created_at < ?", institutionID, domain.LeaveStatusPending, endTime).
		Count(&pending).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("leave repository: count pending leaves: %w", err)
	}

	return int(approved), int(rejected), int(pending), nil
}

func (r *leaveRepository) GetFacultyLeaveStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) ([]domain.FacultyLeaveEntry, error) {
	var entries []domain.FacultyLeaveEntry

	query := `
		SELECT 
			CAST(u.id AS VARCHAR) AS faculty_id,
			u.name AS name,
			COUNT(CASE WHEN lr.status = 'pending' AND lr.created_at < ? THEN 1 END) AS pending,
			COUNT(CASE WHEN lr.status = 'approved' AND lr.reviewed_at >= ? AND lr.reviewed_at < ? THEN 1 END) AS approved,
			COUNT(CASE WHEN lr.status = 'rejected' AND lr.reviewed_at >= ? AND lr.reviewed_at < ? THEN 1 END) AS rejected
		FROM leave_requests lr
		JOIN users u ON u.id = lr.user_id
		WHERE lr.institution_id = ? AND u.role = 'faculty'
		GROUP BY u.id, u.name
		HAVING COUNT(CASE WHEN lr.status = 'pending' AND lr.created_at < ? THEN 1 END) > 0 
		    OR COUNT(CASE WHEN lr.status = 'approved' AND lr.reviewed_at >= ? AND lr.reviewed_at < ? THEN 1 END) > 0 
		    OR COUNT(CASE WHEN lr.status = 'rejected' AND lr.reviewed_at >= ? AND lr.reviewed_at < ? THEN 1 END) > 0
	`

	if err := r.getDB(ctx).Raw(query, endTime, startTime, endTime, startTime, endTime, institutionID, endTime, startTime, endTime, startTime, endTime).Scan(&entries).Error; err != nil {
		return nil, fmt.Errorf("leave repository: get faculty leave stats: %w", err)
	}

	return entries, nil
}

func (r *leaveRepository) GetPendingLeavesWithUser(ctx context.Context) ([]domain.LeaveRequest, error) {
	var requests []domain.LeaveRequest
	err := r.getDB(ctx).
		Preload("User").
		Preload("Institution").
		Where("status = ?", domain.LeaveStatusPending).
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("leave repository: get pending leaves with user: %w", err)
	}
	return requests, nil
}

