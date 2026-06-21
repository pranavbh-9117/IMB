// Package repository provides PostgreSQL data access implementations for users.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user record into the database.
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("user repository: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single active user by their UUID.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user repository: get by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("user repository: get by id: %w", err)
	}
	return &user, nil
}

//Return list of active users
func (r *userRepository) List(ctx context.Context, institutionID *uuid.UUID, offset, limit int) ([]domain.User, error) {
	var users []domain.User
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if institutionID != nil {
		query = query.Where("institution_id = ?", *institutionID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("user repository: list: %w", err)
	}
	return users, nil
}

// Update persists modified user attributes to the database.
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("user repository: update: %w", err)
	}
	return nil
}

// Delete performs a soft-delete by toggling the is_active flag to false.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
	if err != nil {
		return fmt.Errorf("user repository: delete: %w", err)
	}
	return nil
}

// EmailExists checks if the given email is currently bound to any user

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("user repository: email exists: %w", err)
	}
	return count > 0, nil
}
