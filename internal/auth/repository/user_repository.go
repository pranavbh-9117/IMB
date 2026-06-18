package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// userRepository is the GORM-backed implementation of UserRepository.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a UserRepository backed by the provided *gorm.DB.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// FindByEmail retrieves a user by exact email match.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user repository: find by email: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("user repository: find by email: %w", err)
	}

	return &user, nil
}

// FindByID retrieves a user by primary key UUID.
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User

	err := r.db.WithContext(ctx).
		First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user repository: find by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("user repository: find by id: %w", err)
	}

	return &user, nil
}

// UpdatePasswordHash replaces the bcrypt hash stored for the given user.
func (r *userRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error {
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password_hash", newHash).Error
	if err != nil {
		return fmt.Errorf("user repository: update password hash: %w", err)
	}

	return nil
}
