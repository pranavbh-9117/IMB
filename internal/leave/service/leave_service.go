// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/leave/repository"
)

type leaveService struct {
	repo repository.LeaveRepository
}

// NewLeaveService creates a new LeaveService.
func NewLeaveService(repo repository.LeaveRepository) LeaveService {
	return &leaveService{repo: repo}
}

// InitializeBalance implements the corresponding interface or provides the named functionality.
func (s *leaveService) InitializeBalance(ctx context.Context, userID, institutionID uuid.UUID, role domain.Role) error {
	totalDays := 0
	switch role {
	case domain.RoleSuperAdmin:
		totalDays = 30
	case domain.RoleInstituteAdmin:
		totalDays = 30
	case domain.RoleFaculty:
		totalDays = 20
	case domain.RoleStudent:
		totalDays = 10
	default:
		return fmt.Errorf("%w: unrecognized role for leave balance", ErrInvalidInput)
	}

	balance := &domain.LeaveBalance{
		UserID:        userID,
		InstitutionID: institutionID,
		TotalDays:     totalDays,
		UsedDays:      0,
	}

	if err := s.repo.CreateBalance(ctx, balance); err != nil {
		return fmt.Errorf("leave service: initialize balance: %w", err)
	}
	return nil
}

// GetBalance implements the corresponding interface or provides the named functionality.
func (s *leaveService) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.LeaveBalance, error) {
	balance, err := s.repo.GetBalanceByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBalanceNotFound
		}
		return nil, fmt.Errorf("leave service: get balance: %w", err)
	}
	return balance, nil
}

// ApplyLeave implements the corresponding interface or provides the named functionality.
func (s *leaveService) ApplyLeave(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, req *domain.LeaveRequest) (*domain.LeaveRequest, error) {
	// 1. Only Students and Faculty can apply
	if requesterRole != domain.RoleStudent && requesterRole != domain.RoleFaculty {
		return nil, fmt.Errorf("%w: only students and faculty can apply for leave", ErrUnauthorized)
	}

	// 2. Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("%w: end date cannot be before start date", ErrInvalidInput)
	}

	// 3. Overlap check (fetch all active leaves for user)
	existingLeaves, err := s.repo.ListRequests(ctx, repository.RequestFilter{UserID: &requesterID}, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("leave service: list existing requests: %w", err)
	}

	for _, existing := range existingLeaves {
		if existing.Status == domain.LeaveStatusRejected || existing.Status == domain.LeaveStatusCancelled {
			continue
		}
		// Overlap logic: StartA <= EndB AND EndA >= StartB
		if !req.StartDate.After(existing.EndDate) && !req.EndDate.Before(existing.StartDate) {
			return nil, ErrOverlap
		}
	}

	// 4. Balance check
	requestedDays := calculateRequestedDays(req.StartDate, req.EndDate)
	balance, err := s.repo.GetBalanceByUserID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBalanceNotFound
		}
		return nil, fmt.Errorf("leave service: fetch balance: %w", err)
	}

	if balance.UsedDays+requestedDays > balance.TotalDays {
		return nil, ErrInsufficientBalance
	}

	// 5. Populate secure fields
	req.UserID = requesterID
	req.InstitutionID = requesterInstID
	req.Status = domain.LeaveStatusPending
	req.ReviewedBy = nil
	req.ReviewedAt = nil

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("leave service: create request: %w", err)
	}

	return req, nil
}

// ProcessLeaveApproval implements the corresponding interface or provides the named functionality.
func (s *leaveService) ProcessLeaveApproval(ctx context.Context, requestID, reviewerID uuid.UUID, reviewerRole domain.Role, reviewerInstID uuid.UUID, newStatus domain.LeaveStatus, note string) error {
	if newStatus != domain.LeaveStatusApproved && newStatus != domain.LeaveStatusRejected {
		return fmt.Errorf("%w: invalid target status", ErrInvalidInput)
	}

	return s.repo.DoInTransaction(ctx, func(txCtx context.Context) error {
		req, err := s.repo.GetRequestByID(txCtx, requestID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrRequestNotFound
			}
			return fmt.Errorf("leave service: fetch request: %w", err)
		}

		if req.InstitutionID != reviewerInstID {
			return ErrRequestNotFound // Tenant isolation (mask existence)
		}

		if req.Status != domain.LeaveStatusPending {
			return ErrLeaveNotPending
		}

		// Hierarchical RBAC Check based on explicit business rules
		if req.User.Role == domain.RoleStudent {
			if reviewerRole != domain.RoleFaculty {
				return fmt.Errorf("%w: student leaves can only be approved by faculty", ErrUnauthorized)
			}
		} else if req.User.Role == domain.RoleFaculty {
			if reviewerRole != domain.RoleInstituteAdmin {
				return fmt.Errorf("%w: faculty leaves can only be approved by institute admins", ErrUnauthorized)
			}
		} else {
			return fmt.Errorf("%w: unrecognized role for leave request owner", ErrUnauthorized)
		}

		if newStatus == domain.LeaveStatusApproved {
			requestedDays := calculateRequestedDays(req.StartDate, req.EndDate)
			balance, err := s.repo.GetBalanceByUserID(txCtx, req.UserID)
			if err != nil {
				return fmt.Errorf("leave service: fetch balance: %w", err)
			}

			if balance.UsedDays+requestedDays > balance.TotalDays {
				return ErrInsufficientBalance
			}

			balance.UsedDays += requestedDays
			if err := s.repo.UpdateBalance(txCtx, balance); err != nil {
				return fmt.Errorf("leave service: update balance: %w", err)
			}
		}

		if err := s.repo.UpdateRequestStatus(txCtx, requestID, newStatus, &reviewerID, note); err != nil {
			return fmt.Errorf("leave service: update request status: %w", err)
		}

		return nil
	})
}

// GetLeaveDetails implements the corresponding interface or provides the named functionality.
func (s *leaveService) GetLeaveDetails(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, requestID uuid.UUID) (*domain.LeaveRequest, error) {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequestNotFound
		}
		return nil, fmt.Errorf("leave service: get request: %w", err)
	}

	if req.InstitutionID != requesterInstID {
		return nil, ErrRequestNotFound // Tenant isolation
	}

	if requesterRole == domain.RoleStudent && req.UserID != requesterID {
		return nil, ErrRequestNotFound // Students can only see their own
	}

	return req, nil
}

// ListLeaves implements the corresponding interface or provides the named functionality.
func (s *leaveService) ListLeaves(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, filter repository.RequestFilter, offset, limit int) ([]domain.LeaveRequest, error) {
	// Tenant isolation is absolute
	filter.InstitutionID = &requesterInstID

	// Role-based visibility
	if requesterRole == domain.RoleStudent {
		filter.UserID = &requesterID // Students can only see their own leaves
	}

	requests, err := s.repo.ListRequests(ctx, filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("leave service: list requests: %w", err)
	}
	return requests, nil
}

// CancelLeave implements the corresponding interface or provides the named functionality.
func (s *leaveService) CancelLeave(ctx context.Context, requesterID uuid.UUID, requestID uuid.UUID) error {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRequestNotFound
		}
		return fmt.Errorf("leave service: fetch request: %w", err)
	}

	if req.UserID != requesterID {
		return ErrRequestNotFound
	}

	if req.Status != domain.LeaveStatusPending {
		return ErrLeaveNotPending
	}

	if err := s.repo.UpdateRequestStatus(ctx, requestID, domain.LeaveStatusCancelled, nil, "cancelled by user"); err != nil {
		return fmt.Errorf("leave service: cancel request: %w", err)
	}
	return nil
}

// calculateRequestedDays calculates the inclusive number of calendar days between start and end.
func calculateRequestedDays(start, end time.Time) int {
	// Normalize to start of day to avoid time-of-day offsets
	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	if e.Before(s) {
		return 0
	}

	duration := e.Sub(s)
	days := int(duration.Hours() / 24)
	return days + 1 // Inclusive of the start day
}
