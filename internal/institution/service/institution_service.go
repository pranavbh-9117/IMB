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

func NewInstitutionService(repo repository.InstitutionRepository) InstitutionService {
	return &institutionService{repo: repo}
}

// Create Institution Service
func (s *institutionService) Create(ctx context.Context, inst *domain.Institution) error {
	inst.Name = strings.TrimSpace(inst.Name)
	inst.Code = strings.TrimSpace(inst.Code)

	if inst.Name == "" || inst.Code == "" {
		return ErrInvalidInput
	}

	existing, err := s.repo.FindByCode(ctx, inst.Code)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("institution service: create: %w", err)
	}
	if existing != nil {
		return ErrDuplicateCode
	}

	inst.IsActive = true

	if err := s.repo.Create(ctx, inst); err != nil {
		return fmt.Errorf("institution service: create: %w", err)
	}

	return nil
}

// Get Institution By ID
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

// List of institutions with offest and limit
func (s *institutionService) List(ctx context.Context, offset, limit int) ([]domain.Institution, error) {
	institutions, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("institution service: list: %w", err)
	}
	return institutions, nil
}

// Update Institution details
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

	existing.Name = newName
	existing.Address = strings.TrimSpace(updates.Address)
	existing.Phone = strings.TrimSpace(updates.Phone)
	existing.Email = strings.TrimSpace(updates.Email)

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
