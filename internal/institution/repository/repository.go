// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type InstitutionRepository interface {
	Create(ctx context.Context, inst *domain.Institution) error

	GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error)

	FindByCode(ctx context.Context, code string) (*domain.Institution, error)

	List(ctx context.Context, offset, limit int) ([]domain.Institution, error)

	Update(ctx context.Context, inst *domain.Institution) error

	Delete(ctx context.Context, id uuid.UUID) error
}
