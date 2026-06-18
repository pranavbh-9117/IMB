// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// ErrNotFound is returned by repository methods when the requested record
// does not exist. The service layer uses errors.Is(err, ErrNotFound) to detect
// this condition without importing GORM.
var ErrNotFound = errors.New("record not found")

// InstitutionRepository defines the data-access contract for the institution
// management module. Implementations must not contain business logic.
type InstitutionRepository interface {
	// Create inserts a new institution record.
	Create(ctx context.Context, inst *domain.Institution) error

	// GetByID retrieves an institution by its primary key UUID. Returns a
	// wrapped ErrNotFound if the record does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error)

	// FindByCode retrieves an institution by its unique code. Returns a
	// wrapped ErrNotFound if the record does not exist.
	FindByCode(ctx context.Context, code string) (*domain.Institution, error)

	// List retrieves a batch of institutions using limit and offset parameters
	// for pagination support.
	List(ctx context.Context, offset, limit int) ([]domain.Institution, error)

	// Update saves modifications to an existing institution record.
	Update(ctx context.Context, inst *domain.Institution) error

	// Delete performs a status toggle (soft delete via IsActive flag) on the
	// institution with the given ID.
	Delete(ctx context.Context, id uuid.UUID) error
}
