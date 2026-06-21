// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrAccountInactive = errors.New("account is deactivated")

	ErrTokenInvalid = errors.New("refresh token is invalid or expired")

	ErrWrongPassword = errors.New("current password is incorrect")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserInfo struct {
	ID    string
	Name  string
	Email string
	Role  string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         UserInfo
}

// Authentication business logics
type AuthService interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)

	Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error)

	Logout(ctx context.Context, rawRefreshToken string) error

	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
