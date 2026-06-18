package dto

import "time"

// CreateRequest defines the payload for creating a user.
// InstitutionID is required for Super Admins but ignored for Institute Admins.
type CreateRequest struct {
	Name          string  `json:"name" binding:"required,max=255"`
	Email         string  `json:"email" binding:"required,email,max=255"`
	Role          string  `json:"role" binding:"required,oneof=institute_admin faculty student"`
	InstitutionID *string `json:"institution_id" binding:"omitempty,uuid"`
}

// UpdateRequest defines the payload for updating a user.
// Role is immutable (ADR-006). If included, the handler will reject the request.
type UpdateRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=255"`
	Email *string `json:"email" binding:"omitempty,email,max=255"`
	Role  *string `json:"role" binding:"omitempty"` // Present to detect and reject modifications
}

// UserResponse defines the standard payload for returning user data.
type UserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	InstitutionID *string   `json:"institution_id,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateResponse defines the payload returned upon successful creation.
// It includes the generated temporary password (ADR-004).
type CreateResponse struct {
	User              UserResponse `json:"user"`
	TemporaryPassword string       `json:"temporary_password"`
}
