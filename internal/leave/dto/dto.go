package dto

import "time"

// ApplyLeaveRequest represents the payload for creating a new leave request.
type ApplyLeaveRequest struct {
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	Reason    string    `json:"reason" binding:"required,max=500"`
}

// ProcessLeaveRequest represents the payload for approving or rejecting a leave.
type ProcessLeaveRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Note   string `json:"note" binding:"max=500"` // Optional comment
}

// LeaveResponse represents the standard response format for a leave request.
type LeaveResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	InstitutionID string     `json:"institution_id"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ReviewedBy    *string    `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote    string     `json:"review_note,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
