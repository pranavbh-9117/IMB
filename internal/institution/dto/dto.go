// Package dto provides dto functionality for the IMB platform.
package dto

import "time"


type CreateRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	Code    string `json:"code" binding:"required,max=50"`
	Address string `json:"address"`
	Phone   string `json:"phone" binding:"omitempty,max=20"`
	Email   string `json:"email" binding:"omitempty,email,max=255"`
}


type UpdateInstitutionInput struct {
	Name    *string `json:"name" binding:"omitempty,max=255"`
	Address *string `json:"address"`
	Phone   *string `json:"phone" binding:"omitempty,max=20"`
	Email   *string `json:"email" binding:"omitempty,email,max=255"`
}

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
