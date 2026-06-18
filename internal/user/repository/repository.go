// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// ErrNotFound is returned by repository methods when the requested user
// does not exist or is inactive (soft deleted).
var ErrNotFound = errors.New("record not found")

// UserRepository defines the data-access contract for the user management module.
type UserRepository interface {
	// Create inserts a new user record.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves an active user by primary key. Returns ErrNotFound if the
	// user does not exist or is deactivated.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail retrieves an active user by email. Returns ErrNotFound if the
	// user does not exist or is deactivated.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// List retrieves active users, optionally filtered by InstitutionID.
	List(ctx context.Context, institutionID *uuid.UUID, offset, limit int) ([]domain.User, error)

	// Update saves modifications to an existing user record.
	Update(ctx context.Context, user *domain.User) error

	// Delete performs a soft delete by setting IsActive = false on the user.
	Delete(ctx context.Context, id uuid.UUID) error

	// EmailExists checks if the given email is taken by any user, including
	// deactivated ones, enforcing global email identity constraints (ADR-008).
	EmailExists(ctx context.Context, email string) (bool, error)
}
