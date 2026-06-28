// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type RequestFilter struct {
	InstitutionID *uuid.UUID
	UserID        *uuid.UUID
	ReviewerID    *uuid.UUID
	Status        *domain.LeaveStatus
}

// LeaveRepository defines the data-access contract for the leave management module.

type LeaveRepository interface {
	DoInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error

	CreateRequest(ctx context.Context, request *domain.LeaveRequest) error

	GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)

	ListRequests(ctx context.Context, filter RequestFilter, offset, limit int) ([]domain.LeaveRequest, error)

	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status domain.LeaveStatus, reviewerID *uuid.UUID, note string) error

	CreateBalance(ctx context.Context, balance *domain.LeaveBalance) error

	GetBalanceByUserID(ctx context.Context, userID uuid.UUID) (*domain.LeaveBalance, error)

	UpdateBalance(ctx context.Context, balance *domain.LeaveBalance) error

	GetInstitutionLeaveStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) (int, int, int, error)

	GetFacultyLeaveStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) ([]domain.FacultyLeaveEntry, error)

	GetPendingLeavesWithUser(ctx context.Context) ([]domain.LeaveRequest, error)
}
