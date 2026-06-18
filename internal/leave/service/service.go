// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/leave/repository"
)

// Sentinel errors for leave business logic
var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrInsufficientBalance = errors.New("insufficient leave balance")
	ErrLeaveNotPending     = errors.New("leave request is not in a pending state")
	ErrBalanceNotFound     = errors.New("leave balance not found")
	ErrRequestNotFound     = errors.New("leave request not found")
	ErrUnauthorized        = errors.New("unauthorized action")
	ErrOverlap             = errors.New("leave request dates overlap with an existing request")
)

// LeaveService defines the business logic operations for leave management.
type LeaveService interface {
	// InitializeBalance creates the initial LeaveBalance for a newly provisioned user.
	InitializeBalance(ctx context.Context, userID, institutionID uuid.UUID, role domain.Role) error

	// GetBalance retrieves the current balance for a user.
	GetBalance(ctx context.Context, userID uuid.UUID) (*domain.LeaveBalance, error)

	// ApplyLeave creates a new pending leave request. Enforces overlap validation and restricts to Students/Faculty.
	ApplyLeave(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, req *domain.LeaveRequest) (*domain.LeaveRequest, error)

	// ProcessLeaveApproval transitions a leave request status enforcing strict hierarchical RBAC.
	ProcessLeaveApproval(ctx context.Context, requestID, reviewerID uuid.UUID, reviewerRole domain.Role, reviewerInstID uuid.UUID, newStatus domain.LeaveStatus, note string) error

	// GetLeaveDetails fetches a specific request, enforcing tenant isolation.
	GetLeaveDetails(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, requestID uuid.UUID) (*domain.LeaveRequest, error)

	// ListLeaves fetches leave requests dynamically.
	ListLeaves(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, filter repository.RequestFilter, offset, limit int) ([]domain.LeaveRequest, error)

	// CancelLeave allows a user to cancel their own pending leave request.
	CancelLeave(ctx context.Context, requesterID uuid.UUID, requestID uuid.UUID) error
}
