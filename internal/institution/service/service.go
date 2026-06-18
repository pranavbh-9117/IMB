package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// Sentinel errors returned by InstitutionService methods. Handlers use these
// to map business errors to HTTP status codes.
var (
	// ErrInstitutionNotFound is returned when an institution cannot be found by ID.
	ErrInstitutionNotFound = errors.New("institution not found")

	// ErrDuplicateCode is returned during creation if an institution with the
	// same code already exists.
	ErrDuplicateCode = errors.New("institution code already exists")

	// ErrInvalidInput is returned when domain constraints are violated, such as
	// providing an empty or whitespace-only name or code.
	ErrInvalidInput = errors.New("invalid institution data")
)

// InstitutionService defines the business logic contract for institution management.
type InstitutionService interface {
	// Create validates the input and creates a new institution. It enforces
	// the uniqueness of the institution code.
	Create(ctx context.Context, inst *domain.Institution) error

	// GetByID retrieves a single institution by its UUID. Returns ErrInstitutionNotFound
	// if it does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error)

	// List retrieves institutions, delegating offset and limit to the repository.
	List(ctx context.Context, offset, limit int) ([]domain.Institution, error)

	// Update modifies an existing institution. The 'Code' field cannot be changed.
	// Empty 'Name' updates are rejected.
	Update(ctx context.Context, id uuid.UUID, updates *domain.Institution) error

	// Delete deactivates an institution, preventing future associations.
	// It operates idempotently.
	Delete(ctx context.Context, id uuid.UUID) error
}
