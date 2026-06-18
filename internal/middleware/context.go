package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// Unexported context keys to prevent collisions and enforce the use of
// typed extraction functions.
const (
	userIDKey        = "auth.user_id"
	roleKey          = "auth.role"
	institutionIDKey = "auth.institution_id"
)

// ErrContextValueMissing is returned when a required authentication value
// is not present in the Gin context. This typically indicates that a protected
// handler was mounted without the RequireAuth middleware.
var ErrContextValueMissing = errors.New("authentication value missing from context")

// GetUserID retrieves the authenticated user's UUID from the Gin context.
// It returns an error if the value is missing or is not a valid UUID.
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(userIDKey)
	if !exists {
		return uuid.Nil, ErrContextValueMissing
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a valid UUID")
	}

	return id, nil
}

// GetRole retrieves the authenticated user's Role from the Gin context.
// It returns an error if the value is missing.
func GetRole(c *gin.Context) (domain.Role, error) {
	val, exists := c.Get(roleKey)
	if !exists {
		return "", ErrContextValueMissing
	}

	role, ok := val.(domain.Role)
	if !ok {
		return "", errors.New("role in context is not a valid domain.Role")
	}

	return role, nil
}

// GetInstitutionID retrieves the authenticated user's Institution UUID from
// the Gin context. It returns a nil pointer if the user has no institution
// (e.g., Super Admin), or an error if the value is missing or invalid.
func GetInstitutionID(c *gin.Context) (*uuid.UUID, error) {
	val, exists := c.Get(institutionIDKey)
	if !exists {
		return nil, ErrContextValueMissing
	}

	// Super admins will explicitly have a nil pointer in the context.
	if val == nil {
		return nil, nil
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return nil, errors.New("institution ID in context is not a valid UUID")
	}

	return &id, nil
}
