package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/user/repository"
	"github.com/pranavbh-9117/IMB/pkg/password"
)

type userService struct {
	repo      repository.UserRepository
	leaveInit LeaveInitializer
}

// NewUserService creates a new UserService to enforce user management business rules.
func NewUserService(repo repository.UserRepository, leaveInit LeaveInitializer) UserService {
	return &userService{repo: repo, leaveInit: leaveInit}
}

func (s *userService) Create(ctx context.Context, creatorRole domain.Role, creatorInstID *uuid.UUID, user *domain.User) (*CreateResult, error) {
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))

	if user.Name == "" || user.Email == "" || user.Role == "" {
		return nil, ErrInvalidInput
	}

	// ADR-006: Enforce Role Hierarchy
	if creatorRole == domain.RoleSuperAdmin {
		if user.Role != domain.RoleInstituteAdmin {
			return nil, fmt.Errorf("%w: super admin can only create institute admins", ErrInvalidRole)
		}
	} else if creatorRole == domain.RoleInstituteAdmin {
		if user.Role != domain.RoleFaculty && user.Role != domain.RoleStudent {
			return nil, fmt.Errorf("%w: institute admin can only create faculty or student", ErrInvalidRole)
		}
	} else {
		return nil, fmt.Errorf("%w: you are not authorized to create users", ErrUnauthorized)
	}

	// ADR-005: Enforce Institution Assignment
	if creatorRole == domain.RoleSuperAdmin {
		if user.InstitutionID == nil {
			return nil, fmt.Errorf("%w: institution_id is required for super admin creating an institute admin", ErrInvalidInput)
		}
	} else if creatorRole == domain.RoleInstituteAdmin {
		// Forcefully derive from logged-in admin
		user.InstitutionID = creatorInstID
	}

	// ADR-008: Global Email Uniqueness
	exists, err := s.repo.EmailExists(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("user service: email check: %w", err)
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	// ADR-004: Generate Temporary Password
	tempPass, err := password.GenerateTemp(12)
	if err != nil {
		return nil, fmt.Errorf("user service: temp password: %w", err)
	}

	hash, err := password.Hash(tempPass)
	if err != nil {
		return nil, fmt.Errorf("user service: hash password: %w", err)
	}

	user.PasswordHash = hash
	user.MustChangePassword = true
	user.IsActive = true

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user service: create: %w", err)
	}

	// Trigger Leave Balance Initialization if the dependency is provided
	if s.leaveInit != nil && user.InstitutionID != nil {
		if err := s.leaveInit.InitializeBalance(ctx, user.ID, *user.InstitutionID, user.Role); err != nil {
			// In Phase 5, we log the error but do not fail the user creation.
			fmt.Printf("Warning: failed to initialize leave balance for user %s: %v\n", user.ID, err)
		}
	}

	return &CreateResult{
		User:         user,
		TempPassword: tempPass,
	}, nil
}

func (s *userService) GetByID(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) (*domain.User, error) {
	if requesterRole != domain.RoleSuperAdmin && requesterRole != domain.RoleInstituteAdmin {
		return nil, ErrUnauthorized
	}

	user, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user service: get by id: %w", err)
	}

	// ADR-009: Tenant Leakage Prevention
	if requesterRole == domain.RoleInstituteAdmin {
		if user.InstitutionID == nil || *user.InstitutionID != *requesterInstID {
			return nil, ErrUserNotFound // Mask as 404
		}
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, offset, limit int) ([]domain.User, error) {
	if requesterRole != domain.RoleSuperAdmin && requesterRole != domain.RoleInstituteAdmin {
		return nil, ErrUnauthorized
	}

	var filterInstID *uuid.UUID
	if requesterRole == domain.RoleInstituteAdmin {
		filterInstID = requesterInstID
	}

	users, err := s.repo.List(ctx, filterInstID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("user service: list: %w", err)
	}
	return users, nil
}

func (s *userService) Update(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID, updates *domain.User) error {
	// ADR-010: Prevent self management via this API
	if requesterID == targetID {
		return ErrSelfManagement
	}

	if requesterRole != domain.RoleSuperAdmin && requesterRole != domain.RoleInstituteAdmin {
		return ErrUnauthorized
	}

	existing, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("user service: update get target: %w", err)
	}

	// ADR-009: Tenant Leakage Prevention
	if requesterRole == domain.RoleInstituteAdmin {
		if existing.InstitutionID == nil || *existing.InstitutionID != *requesterInstID {
			return ErrUserNotFound
		}
	}

	// ADR-006: Role Immutability
	if updates.Role != "" && updates.Role != existing.Role {
		return ErrRoleImmutable
	}

	// Role boundary checking just in case
	if requesterRole == domain.RoleSuperAdmin && existing.Role != domain.RoleInstituteAdmin {
		return fmt.Errorf("%w: super admin can only update institute admins", ErrInvalidRole)
	}
	if requesterRole == domain.RoleInstituteAdmin && (existing.Role != domain.RoleFaculty && existing.Role != domain.RoleStudent) {
		return fmt.Errorf("%w: institute admin can only update faculty or students", ErrInvalidRole)
	}

	newName := strings.TrimSpace(updates.Name)
	if newName != "" {
		existing.Name = newName
	}

	newEmail := strings.TrimSpace(strings.ToLower(updates.Email))
	if newEmail != "" && newEmail != existing.Email {
		exists, err := s.repo.EmailExists(ctx, newEmail)
		if err != nil {
			return fmt.Errorf("user service: email check: %w", err)
		}
		if exists {
			return ErrDuplicateEmail
		}
		existing.Email = newEmail
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("user service: update execute: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) error {
	// ADR-010: Prevent self management
	if requesterID == targetID {
		return ErrSelfManagement
	}

	if requesterRole != domain.RoleSuperAdmin && requesterRole != domain.RoleInstituteAdmin {
		return ErrUnauthorized
	}

	existing, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("user service: delete get target: %w", err)
	}

	// ADR-010: Prevent SUPER_ADMIN lockout
	if existing.Role == domain.RoleSuperAdmin {
		return ErrLockoutPrevention
	}

	// ADR-009: Tenant Leakage Prevention
	if requesterRole == domain.RoleInstituteAdmin {
		if existing.InstitutionID == nil || *existing.InstitutionID != *requesterInstID {
			return ErrUserNotFound
		}
	}

	// Role boundary check
	if requesterRole == domain.RoleSuperAdmin && existing.Role != domain.RoleInstituteAdmin {
		return fmt.Errorf("%w: super admin can only delete institute admins", ErrInvalidRole)
	}
	if requesterRole == domain.RoleInstituteAdmin && (existing.Role != domain.RoleFaculty && existing.Role != domain.RoleStudent) {
		return fmt.Errorf("%w: institute admin can only delete faculty or students", ErrInvalidRole)
	}

	if err := s.repo.Delete(ctx, targetID); err != nil {
		return fmt.Errorf("user service: delete execute: %w", err)
	}

	return nil
}
