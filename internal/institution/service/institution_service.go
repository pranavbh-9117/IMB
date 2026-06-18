// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/institution/repository"
)

type institutionService struct {
	repo repository.InstitutionRepository
}

// NewInstitutionService creates a new instance of InstitutionService.
func NewInstitutionService(repo repository.InstitutionRepository) InstitutionService {
	return &institutionService{repo: repo}
}

// Create implements the corresponding interface or provides the named functionality.
func (s *institutionService) Create(ctx context.Context, inst *domain.Institution) error {
	inst.Name = strings.TrimSpace(inst.Name)
	inst.Code = strings.TrimSpace(inst.Code)

	if inst.Name == "" || inst.Code == "" {
		return ErrInvalidInput
	}

	// Enforce Code uniqueness
	existing, err := s.repo.FindByCode(ctx, inst.Code)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("institution service: create: %w", err)
	}
	if existing != nil {
		return ErrDuplicateCode
	}

	// Active by default
	inst.IsActive = true

	if err := s.repo.Create(ctx, inst); err != nil {
		return fmt.Errorf("institution service: create: %w", err)
	}

	return nil
}

// GetByID implements the corresponding interface or provides the named functionality.
func (s *institutionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error) {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInstitutionNotFound
		}
		return nil, fmt.Errorf("institution service: get by id: %w", err)
	}
	return inst, nil
}

// List implements the corresponding interface or provides the named functionality.
func (s *institutionService) List(ctx context.Context, offset, limit int) ([]domain.Institution, error) {
	institutions, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("institution service: list: %w", err)
	}
	return institutions, nil
}

// Update implements the corresponding interface or provides the named functionality.
func (s *institutionService) Update(ctx context.Context, id uuid.UUID, updates *domain.Institution) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInstitutionNotFound
		}
		return fmt.Errorf("institution service: update: %w", err)
	}

	newName := strings.TrimSpace(updates.Name)
	if newName == "" {
		return ErrInvalidInput
	}

	// Apply allowed updates. Code is immutable.
	existing.Name = newName
	existing.Address = strings.TrimSpace(updates.Address)
	existing.Phone = strings.TrimSpace(updates.Phone)
	existing.Email = strings.TrimSpace(updates.Email)

	// Note: We don't toggle IsActive here; we leave that to Delete/Restore flows.
	// But if needed in the future, it can be added.

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("institution service: update: %w", err)
	}

	return nil
}

// Delete implements the corresponding interface or provides the named functionality.
func (s *institutionService) Delete(ctx context.Context, id uuid.UUID) error {
	// First check if it exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInstitutionNotFound
		}
		return fmt.Errorf("institution service: delete: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("institution service: delete: %w", err)
	}

	return nil
}
