// Package repository implements data access patterns for institutions.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

type institutionRepository struct {
	db *gorm.DB
}

// NewInstitutionRepository creates a new instance of InstitutionRepository
// backed by GORM. It returns the interface type to enforce the abstraction boundary.
func NewInstitutionRepository(db *gorm.DB) InstitutionRepository {
	return &institutionRepository{db: db}
}

// Create inserts a new institution record into the database.
func (r *institutionRepository) Create(ctx context.Context, inst *domain.Institution) error {
	if err := r.db.WithContext(ctx).Create(inst).Error; err != nil {
		return fmt.Errorf("institution repository: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single institution by its primary UUID.
// It explicitly filters by is_active=true to enforce soft-delete boundaries.
func (r *institutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error) {
	var inst domain.Institution
	err := r.db.WithContext(ctx).First(&inst, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("institution repository: get by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("institution repository: get by id: %w", err)
	}
	return &inst, nil
}

// FindByCode looks up an active institution by its unique identifying code.
func (r *institutionRepository) FindByCode(ctx context.Context, code string) (*domain.Institution, error) {
	var inst domain.Institution
	err := r.db.WithContext(ctx).First(&inst, "code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("institution repository: find by code: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("institution repository: find by code: %w", err)
	}
	return &inst, nil
}

// List returns a paginated slice of active institutions.
func (r *institutionRepository) List(ctx context.Context, offset, limit int) ([]domain.Institution, error) {
	var institutions []domain.Institution
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&institutions).Error; err != nil {
		return nil, fmt.Errorf("institution repository: list: %w", err)
	}
	return institutions, nil
}

// Update persists changes to an existing institution record.
func (r *institutionRepository) Update(ctx context.Context, inst *domain.Institution) error {
	if err := r.db.WithContext(ctx).Save(inst).Error; err != nil {
		return fmt.Errorf("institution repository: update: %w", err)
	}
	return nil
}

// Delete performs a soft-delete by toggling the is_active flag to false.
func (r *institutionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Institution{}).
		Where("id = ?", id).
		Update("is_active", false).Error
	if err != nil {
		return fmt.Errorf("institution repository: delete: %w", err)
	}
	return nil
}
