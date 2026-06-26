// Package repository provides repository functionality for the IMB platform.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// passwordResetTokenRepository implements PasswordResetTokenRepository interface
type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

// Create inserts a new password reset token record.
func (r *passwordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	err := r.db.WithContext(ctx).Create(token).Error
	if err != nil {
		return fmt.Errorf("password reset token repository: create: %w", err)
	}

	return nil
}

// FindByHash retrieves a password reset token record
func (r *passwordResetTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	var token domain.PasswordResetToken

	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("password reset token repository: find by hash: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("password reset token repository: find by hash: %w", err)
	}

	return &token, nil
}

// MarkAsUsed marks a PasswordResetToken as used
func (r *passwordResetTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&domain.PasswordResetToken{}).
		Where("id = ?", id).
		Update("is_used", true).Error
	if err != nil {
		return fmt.Errorf("password reset token repository: mark as used: %w", err)
	}

	return nil
}
