// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateEmail    = errors.New("email is already registered")
	ErrUnauthorized      = errors.New("unauthorized action")
	ErrInvalidRole       = errors.New("invalid or unauthorized role assignment")
	ErrInvalidInput      = errors.New("invalid user data")
	ErrRoleImmutable     = errors.New("role cannot be modified after creation")
	ErrSelfManagement    = errors.New("users cannot modify or delete their own accounts via management APIs")
	ErrLockoutPrevention = errors.New("super admin accounts cannot be deleted")
)

type CreateResult struct {
	User         *domain.User
	TempPassword string
}

type LeaveInitializer interface {
	InitializeBalance(ctx context.Context, userID, institutionID uuid.UUID, role domain.Role) error
}

// UserService defines the business logic contract for user management operations.
type UserService interface {
	Create(ctx context.Context, creatorRole domain.Role, creatorInstID *uuid.UUID, user *domain.User) (*CreateResult, error)

	GetByID(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) (*domain.User, error)

	List(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, offset, limit int) ([]domain.User, error)

	Update(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID, updates *domain.User) error

	Delete(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) error
}
