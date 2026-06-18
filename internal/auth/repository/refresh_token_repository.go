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

// refreshTokenRepository is the GORM-backed implementation of
// RefreshTokenRepository.
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository returns a RefreshTokenRepository backed by the
// provided *gorm.DB.
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create inserts a new refresh token record.
func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	err := r.db.WithContext(ctx).Create(token).Error
	if err != nil {
		return fmt.Errorf("refresh token repository: create: %w", err)
	}

	return nil
}

// FindByHash retrieves a refresh token record by its SHA-256 hex digest.
// The returned record may be revoked or expired; validity checks are the
// responsibility of the service layer.
func (r *refreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken

	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("refresh token repository: find by hash: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("refresh token repository: find by hash: %w", err)
	}

	return &token, nil
}

// RevokeByHash marks a single refresh token as revoked by its hash digest.
func (r *refreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	err := r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("token_hash = ?", hash).
		Update("is_revoked", true).Error
	if err != nil {
		return fmt.Errorf("refresh token repository: revoke by hash: %w", err)
	}

	return nil
}

// RevokeAllByUserID marks every refresh token belonging to the given user as
// revoked. Used for logout-everywhere and password change flows.
func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
	if err != nil {
		return fmt.Errorf("refresh token repository: revoke all by user id: %w", err)
	}

	return nil
}
