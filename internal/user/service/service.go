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

// CreateResult contains the newly created user and their temporary password.
type CreateResult struct {
	User         *domain.User
	TempPassword string
}

// UserService defines the business logic contract for user management operations.
type UserService interface {
	// Create generates a user enforcing ADR-005 and ADR-006 hierarchy. Returns a temp password.
	Create(ctx context.Context, creatorRole domain.Role, creatorInstID *uuid.UUID, user *domain.User) (*CreateResult, error)

	// GetByID retrieves a user, enforcing ADR-009 tenant isolation.
	GetByID(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) (*domain.User, error)

	// List retrieves users based on the requester's role isolation boundaries.
	List(ctx context.Context, requesterRole domain.Role, requesterInstID *uuid.UUID, offset, limit int) ([]domain.User, error)

	// Update modifies an existing user while enforcing immutability and isolation rules.
	Update(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID, updates *domain.User) error

	// Delete deactivates a user account, preventing self-deletion and super admin lockouts (ADR-010).
	Delete(ctx context.Context, requesterID uuid.UUID, requesterRole domain.Role, requesterInstID *uuid.UUID, targetID uuid.UUID) error
}
