// Package repository provides data-access implementations for the authentication module.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// ErrNotFound is returned by repository methods when the requested record
// does not exist. Service layers use errors.Is(err, ErrNotFound) to detect
// this condition without importing GORM.
var ErrNotFound = errors.New("record not found")

// UserRepository defines the data-access contract for user lookups required
// by the auth module. Implementations must not apply business-rule filters
// such as IsActive; those decisions belong to the service layer.
type UserRepository interface {
	// FindByEmail retrieves a user by exact email match. Returns a wrapped
	// ErrNotFound when no record exists.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByID retrieves a user by primary key UUID. Returns a wrapped
	// ErrNotFound when no record exists.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// UpdatePasswordHash replaces the stored bcrypt hash for the given user.
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error
}

// RefreshTokenRepository defines the data-access contract for refresh token
// persistence required by the auth module. Implementations store and manage
// token hashes only; raw tokens are never passed through this interface.
type RefreshTokenRepository interface {
	// Create inserts a new refresh token record into the database.
	Create(ctx context.Context, token *domain.RefreshToken) error

	// FindByHash retrieves a refresh token record whose token_hash column
	// matches the provided SHA-256 hex digest. Returns the row as-is,
	// including revoked or expired records; the service layer applies
	// validity checks. Returns nil and an error wrapping
	// gorm.ErrRecordNotFound when no match exists.
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)

	// RevokeByHash sets is_revoked to true on the token whose token_hash
	// matches the provided digest. Used by single-session logout.
	RevokeByHash(ctx context.Context, hash string) error

	// RevokeAllByUserID sets is_revoked to true on every refresh token
	// belonging to the given user. Used by logout-everywhere and password
	// change flows.
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
