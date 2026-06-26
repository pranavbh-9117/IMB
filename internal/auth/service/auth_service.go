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
	"github.com/pranavbh-9117/IMB/pkg/retry"
	"github.com/pranavbh-9117/IMB/pkg/tokenutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

// authService implements AuthService
type authService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
	cfg       config.JWTConfig
	oauthCfg  config.OAuthConfig
	oauthConf *oauth2.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	cfg config.JWTConfig,
	oauthCfg config.OAuthConfig,
) AuthService {
	oauthConf := &oauth2.Config{
		ClientID:     oauthCfg.GoogleClientID,
		ClientSecret: oauthCfg.GoogleClientSecret,
		RedirectURL:  oauthCfg.GoogleCallbackURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		Endpoint: google.Endpoint,
	}

	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
		oauthCfg:  oauthCfg,
		oauthConf: oauthConf,
	}
}

// Login Service
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

// Refresh Token service
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

	if err := s.tokenRepo.RevokeByHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("auth service: refresh: revoke old token: %w", err)
	}

	return s.issueTokenPair(ctx, user)
}

// Logout Service
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

// ChangePassword Service
func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("auth service: change password: find user: %w", err)
	}

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

func (s *authService) GetGoogleLoginURL() (url string, state string) {
	state = uuid.New().String()
	url = s.oauthConf.AuthCodeURL(state)
	return url, state
}

// OAuthRetryConfig 
var OAuthRetryConfig = retry.Config{
	MaxAttempts:  3,
	InitialDelay: 200 * time.Millisecond,
	MaxDelay:     2 * time.Second,
	Multiplier:   2.0,
	Jitter:       0.3,
	ShouldRetry:  retry.IsOAuthTransientError,
}

func (s *authService) GoogleCallback(ctx context.Context, code string) (*LoginResult, error) {
	token, err := s.oauthConf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("auth service: google callback: exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("auth service: google callback: no id_token in response")
	}

	var payload *idtoken.Payload

	//retry mechanism
	err = retry.Do(ctx, func() error {
		var verifyErr error
		payload, verifyErr = idtoken.Validate(ctx, rawIDToken, s.oauthCfg.GoogleClientID)
		return verifyErr
	}, OAuthRetryConfig)
	if err != nil {
		return nil, fmt.Errorf("auth service: google callback: validate id_token: %w", err)
	}

	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		return nil, errors.New("auth service: google callback: invalid issuer")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, errors.New("auth service: google callback: email not found in id_token")
	}

	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return nil, ErrGoogleEmailUnverified
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAccountNotProvisioned
		}
		return nil, fmt.Errorf("auth service: google callback: find user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if user.GoogleID == "" {
		if err := s.userRepo.UpdateGoogleID(ctx, user.ID, payload.Subject); err != nil {
			return nil, fmt.Errorf("auth service: google callback: link google id: %w", err)
		}
	} else if user.GoogleID != payload.Subject {
		return nil, ErrGoogleProfileMismatch
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

// issueTokenPair helper function to generate JWT and RefreshToken
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
