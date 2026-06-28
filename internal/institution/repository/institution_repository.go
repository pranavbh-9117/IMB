// Package repository implements data access patterns for institutions.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

type institutionRepository struct {
	db *gorm.DB
}


func NewInstitutionRepository(db *gorm.DB) InstitutionRepository {
	return &institutionRepository{db: db}
}

func (r *institutionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetSession(ctx, r.db)
}

// Create inserts a new institution record into the database.
func (r *institutionRepository) Create(ctx context.Context, inst *domain.Institution) error {
	if err := r.getDB(ctx).Create(inst).Error; err != nil {
		return fmt.Errorf("institution repository: create: %w", err)
	}
	return nil
}

// GetByID retrieves an active institution by ID.
func (r *institutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error) {
	var inst domain.Institution
	err := r.getDB(ctx).Where("id = ? AND is_active = ?", id, true).First(&inst).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("institution repository: get by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("institution repository: get by id: %w", err)
	}
	return &inst, nil
}

// FindByCode looks up an active institution by its unique identifying code,
// allowing deleted codes to be recycled by new institutions.
func (r *institutionRepository) FindByCode(ctx context.Context, code string) (*domain.Institution, error) {
	var inst domain.Institution
	err := r.getDB(ctx).Where("code = ? AND is_active = ?", code, true).First(&inst).Error
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
	query := r.getDB(ctx).Where("is_active = ?", true)

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
	if err := r.getDB(ctx).Save(inst).Error; err != nil {
		return fmt.Errorf("institution repository: update: %w", err)
	}
	return nil
}

// Delete performs a soft-delete by toggling the is_active flag to false.
func (r *institutionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.getDB(ctx).
		Model(&domain.Institution{}).
		Where("id = ?", id).
		Update("is_active", false).Error
	if err != nil {
		return fmt.Errorf("institution repository: delete: %w", err)
	}
	return nil
}
