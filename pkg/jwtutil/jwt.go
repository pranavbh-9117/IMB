// Package jwtutil provides jwtutil functionality for the IMB platform.
package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the custom payload embedded in every access token.
// UserID, Role, and InstitutionID are application-specific fields.
// InstitutionID is an empty string for super_admin accounts that have no
// institution affiliation.
// RegisteredClaims provides the standard JWT fields: ExpiresAt, IssuedAt.
type Claims struct {
	UserID        string `json:"user_id"`
	Role          string `json:"role"`
	InstitutionID string `json:"institution_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed HS256 JWT access token for the given
// user. expiry controls the token lifetime relative to the current time.
// secret is the HMAC signing key and must match the value used in
// ValidateAccessToken.
func GenerateAccessToken(userID, role, institutionID string, expiry time.Duration, secret string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:        userID,
		Role:          role,
		InstitutionID: institutionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("jwtutil: failed to sign access token: %w", err)
	}

	return signed, nil
}

// ValidateAccessToken parses and validates the provided JWT string.
// It enforces that the token was signed with HS256 and that the signature
// matches secret. Returns the embedded Claims on success, or a wrapped
// error on any failure (expired, malformed, wrong algorithm, bad signature).
func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwtutil: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwtutil: failed to parse access token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("jwtutil: access token is invalid")
	}

	return claims, nil
}
