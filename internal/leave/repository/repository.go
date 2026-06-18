// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// ErrNotFound is returned when a requested record does not exist in the database.
var ErrNotFound = errors.New("record not found")

// RequestFilter provides dynamic filtering options for ListRequests.
type RequestFilter struct {
	InstitutionID *uuid.UUID
	UserID        *uuid.UUID
	ReviewerID    *uuid.UUID
	Status        *domain.LeaveStatus
}

// LeaveRepository defines the data-access contract for the leave management module.
// It is strictly responsible for database interactions and enforces no business logic.
type LeaveRepository interface {
	// DoInTransaction executes the provided function within a database transaction.
	// The provided context should be passed to all repository methods called within the function.
	DoInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error

	// --- LeaveRequest Operations ---

	// CreateRequest inserts a new leave request into the database.
	CreateRequest(ctx context.Context, request *domain.LeaveRequest) error

	// GetRequestByID retrieves a specific leave request by its UUID.
	GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.LeaveRequest, error)

	// ListRequests retrieves a list of leave requests based on the provided dynamic filters.
	ListRequests(ctx context.Context, filter RequestFilter, offset, limit int) ([]domain.LeaveRequest, error)

	// UpdateRequestStatus updates the approval status, the reviewer ID, and the review note.
	// It is intended for targeted updates during the review lifecycle.
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status domain.LeaveStatus, reviewerID *uuid.UUID, note string) error

	// --- LeaveBalance Operations ---

	// CreateBalance creates an initial leave balance record for a user.
	CreateBalance(ctx context.Context, balance *domain.LeaveBalance) error

	// GetBalanceByUserID retrieves the leave balance associated with a specific user.
	GetBalanceByUserID(ctx context.Context, userID uuid.UUID) (*domain.LeaveBalance, error)

	// UpdateBalance persists modifications to a user's total or used leave days.
	UpdateBalance(ctx context.Context, balance *domain.LeaveBalance) error
}
