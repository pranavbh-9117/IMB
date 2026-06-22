// Package repository provides data-access implementations for the authentication module.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var ErrNotFound = errors.New("record not found")

// UserRepository contracts
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error

	UpdateGoogleID(ctx context.Context, userID uuid.UUID, googleID string) error
}

// RefreshTokenRepository
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error

	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)

	RevokeByHash(ctx context.Context, hash string) error

	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
