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

func NewLeaveService(repo repository.LeaveRepository) LeaveService {
	return &leaveService{repo: repo}
}

// InitializeBalance when user is created
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

// Get Leave Balance
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

// ApplyLeave
func (s *leaveService) ApplyLeave(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, req *domain.LeaveRequest) (*domain.LeaveRequest, error) {

	if requesterRole != domain.RoleStudent && requesterRole != domain.RoleFaculty {
		return nil, fmt.Errorf("%w: only students and faculty can apply for leave", ErrUnauthorized)
	}

	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("%w: end date cannot be before start date", ErrInvalidInput)
	}

	existingLeaves, err := s.repo.ListRequests(ctx, repository.RequestFilter{UserID: &requesterID}, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("leave service: list existing requests: %w", err)
	}

	for _, existing := range existingLeaves {
		if existing.Status == domain.LeaveStatusRejected || existing.Status == domain.LeaveStatusCancelled {
			continue
		}

		if !req.StartDate.After(existing.EndDate) && !req.EndDate.Before(existing.StartDate) {
			return nil, ErrOverlap
		}
	}

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

// Update Leave status
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
			return ErrRequestNotFound
		}

		if req.Status != domain.LeaveStatusPending {
			return ErrLeaveNotPending
		}

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

// Return leave details
func (s *leaveService) GetLeaveDetails(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, requestID uuid.UUID) (*domain.LeaveRequest, error) {
	req, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRequestNotFound
		}
		return nil, fmt.Errorf("leave service: get request: %w", err)
	}

	if req.InstitutionID != requesterInstID {
		return nil, ErrRequestNotFound
	}

	if requesterRole == domain.RoleStudent && req.UserID != requesterID {
		return nil, ErrRequestNotFound
	}

	return req, nil
}

// ListLeaves based on filters
func (s *leaveService) ListLeaves(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID uuid.UUID, filter repository.RequestFilter, offset, limit int) ([]domain.LeaveRequest, error) {

	filter.InstitutionID = &requesterInstID

	if requesterRole == domain.RoleStudent {
		filter.UserID = &requesterID
	}

	requests, err := s.repo.ListRequests(ctx, filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("leave service: list requests: %w", err)
	}
	return requests, nil
}

// Cancel Created  Leaves
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

	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	if e.Before(s) {
		return 0
	}

	duration := e.Sub(s)
	days := int(duration.Hours() / 24)
	return days + 1
}
