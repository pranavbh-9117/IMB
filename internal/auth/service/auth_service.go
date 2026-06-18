// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/auth/repository"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/jwtutil"
	"github.com/pranavbh-9117/IMB/pkg/password"
	"github.com/pranavbh-9117/IMB/pkg/tokenutil"
)

// authService is the concrete implementation of AuthService.
type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
	cfg       config.JWTConfig
}

// NewAuthService constructs an AuthService with the provided repository
// interfaces and JWT configuration. It returns the interface type to enforce
// the abstraction boundary at the call site.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	cfg config.JWTConfig,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

// Login verifies the provided email and password, checks that the account is
// active, and issues a LoginResult on success.
func (s *authService) Login(ctx context.Context, email, plainPassword string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth service: login: %w", err)
	}

	if user.PasswordHash == "" || !password.Compare(user.PasswordHash, plainPassword) {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User: UserInfo{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}, nil
}

// Refresh validates the incoming raw refresh token, revokes it (rotation),
// re-loads the user from the database to pick up any role or institution
// changes, and issues a new TokenPair.
func (s *authService) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	hash := tokenutil.HashRefreshToken(rawRefreshToken)

	record, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("auth service: refresh: find token: %w", err)
	}

	if record.IsRevoked || time.Now().After(record.ExpiresAt) {
		return nil, ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, record.UserID)
	if err != nil {
		return nil, fmt.Errorf("auth service: refresh: find user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Rotation: revoke the consumed token before issuing a new pair.
	if err := s.tokenRepo.RevokeByHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("auth service: refresh: revoke old token: %w", err)
	}

	return s.issueTokenPair(ctx, user)
}

// Logout revokes the refresh token for the current session.
func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := tokenutil.HashRefreshToken(rawRefreshToken)

	if _, err := s.tokenRepo.FindByHash(ctx, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTokenInvalid
		}
		return fmt.Errorf("auth service: logout: find token: %w", err)
	}

	if err := s.tokenRepo.RevokeByHash(ctx, hash); err != nil {
		return fmt.Errorf("auth service: logout: revoke token: %w", err)
	}

	return nil
}

// ChangePassword verifies the current password, replaces it with a hash of
// the new password, and revokes all active sessions for the user.
func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("auth service: change password: find user: %w", err)
	}

	// Google-only accounts cannot set a password through this flow.
	if user.PasswordHash == "" {
		return ErrInvalidCredentials
	}

	if !password.Compare(user.PasswordHash, oldPassword) {
		return ErrWrongPassword
	}

	newHash, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("auth service: change password: hash: %w", err)
	}

	if err := s.userRepo.UpdatePasswordHash(ctx, userID, newHash); err != nil {
		return fmt.Errorf("auth service: change password: update: %w", err)
	}

	if err := s.tokenRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("auth service: change password: revoke sessions: %w", err)
	}

	return nil
}

// issueTokenPair is a shared helper that generates a signed access token and
// a new raw refresh token, persists the refresh token hash, and returns the
// TokenPair to the caller.
func (s *authService) issueTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	institutionID := ""
	if user.InstitutionID != nil {
		institutionID = user.InstitutionID.String()
	}

	accessToken, err := jwtutil.GenerateAccessToken(
		user.ID.String(),
		string(user.Role),
		institutionID,
		s.cfg.AccessExpiry,
		s.cfg.Secret,
	)
	if err != nil {
		return nil, fmt.Errorf("auth service: issue token pair: access token: %w", err)
	}

	rawRefreshToken, err := tokenutil.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth service: issue token pair: refresh token: %w", err)
	}

	record := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenutil.HashRefreshToken(rawRefreshToken),
		ExpiresAt: time.Now().Add(s.cfg.RefreshExpiry),
	}

	if err := s.tokenRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("auth service: issue token pair: persist token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}
