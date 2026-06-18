// Package service provides service functionality for the IMB platform.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Sentinel errors returned by AuthService methods. Handlers use errors.Is to
// map these to HTTP status codes without importing any HTTP package here.
var (
	// ErrInvalidCredentials is returned when email/password do not match or
	// when a Google-only account attempts email+password login.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrAccountInactive is returned when the matched user has IsActive=false.
	ErrAccountInactive = errors.New("account is deactivated")

	// ErrTokenInvalid is returned when the provided refresh token is not
	// found, is revoked, or has expired.
	ErrTokenInvalid = errors.New("refresh token is invalid or expired")

	// ErrWrongPassword is returned when the supplied current password does
	// not match the stored hash during a password change request.
	ErrWrongPassword = errors.New("current password is incorrect")
)

// TokenPair holds the access token and raw refresh token issued after a
// successful login or token refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// UserInfo carries the user profile fields needed by the handler layer.
// It avoids exposing the domain.User struct outside the service package.
type UserInfo struct {
	ID    string
	Name  string
	Email string
	Role  string
}

// LoginResult is returned by Login and bundles the issued token pair with
// the authenticated user's profile.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         UserInfo
}

// AuthService defines the authentication business logic contract consumed by
// the handler layer. Implementations must not contain HTTP or database logic.
type AuthService interface {
	// Login verifies email+password credentials, issues a token pair, and
	// returns the authenticated user's profile.
	Login(ctx context.Context, email, password string) (*LoginResult, error)

	// Refresh validates the incoming raw refresh token, applies rotation,
	// and issues a new TokenPair.
	Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error)

	// Logout revokes the refresh token associated with the provided raw
	// token string, terminating the current session.
	Logout(ctx context.Context, rawRefreshToken string) error

	// ChangePassword verifies oldPassword against the stored hash, replaces
	// it with a hash of newPassword, and revokes all active sessions for
	// the user.
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
