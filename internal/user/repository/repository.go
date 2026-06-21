// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var ErrNotFound = errors.New("record not found")

// UserRepository defines the data-access contract for the user management module.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error

	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	List(ctx context.Context, institutionID *uuid.UUID, offset, limit int) ([]domain.User, error)

	Update(ctx context.Context, user *domain.User) error

	Delete(ctx context.Context, id uuid.UUID) error

	EmailExists(ctx context.Context, email string) (bool, error)
}
