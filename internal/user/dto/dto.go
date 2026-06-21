// Package dto provides dto functionality for the IMB platform.
package dto

import "time"


type CreateRequest struct {
	Name          string  `json:"name" binding:"required,max=255"`
	Email         string  `json:"email" binding:"required,email,max=255"`
	Role          string  `json:"role" binding:"required,oneof=institute_admin faculty student"`
	InstitutionID *string `json:"institution_id" binding:"omitempty,uuid"`
}


type UpdateRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=255"`
	Email *string `json:"email" binding:"omitempty,email,max=255"`
	Role  *string `json:"role" binding:"omitempty"` // Present to detect and reject modifications
}


type UserResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	InstitutionID *string   `json:"institution_id,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}


type CreateResponse struct {
	User              UserResponse `json:"user"`
	TemporaryPassword string       `json:"temporary_password"`
}
