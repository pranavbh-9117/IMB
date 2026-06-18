package dto

import "time"

// CreateRequest defines the payload for POST /api/v1/institutions.
// Validation rules enforce required fields and constraints before the
// service layer is invoked.
type CreateRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	Code    string `json:"code" binding:"required,max=50"`
	Address string `json:"address"`
	Phone   string `json:"phone" binding:"omitempty,max=20"`
	Email   string `json:"email" binding:"omitempty,email,max=255"`
}

// UpdateRequest defines the payload for PATCH /api/v1/institutions/{id}.
// All fields are optional (pointers) so clients can send only what they
// want to change.
type UpdateRequest struct {
	Name    *string `json:"name" binding:"omitempty,max=255"`
	Address *string `json:"address"`
	Phone   *string `json:"phone" binding:"omitempty,max=20"`
	Email   *string `json:"email" binding:"omitempty,email,max=255"`
}

// InstitutionResponse defines the standard payload for returning institution
// data to the client, decoupling the database model from the API contract.
type InstitutionResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
