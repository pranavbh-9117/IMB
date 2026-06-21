// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var (
	ErrInstitutionNotFound = errors.New("institution not found")

	ErrDuplicateCode = errors.New("institution code already exists")

	ErrInvalidInput = errors.New("invalid institution data")
)

// InstitutionService defines the business logic contract for institution management.
type InstitutionService interface {
	Create(ctx context.Context, inst *domain.Institution) error

	GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error)

	List(ctx context.Context, offset, limit int) ([]domain.Institution, error)

	Update(ctx context.Context, id uuid.UUID, updates *domain.Institution) error

	Delete(ctx context.Context, id uuid.UUID) error
}
