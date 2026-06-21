// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Types of Leave Status
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

// LeaveBalance Model
type LeaveBalance struct {
	Base

	UserID        uuid.UUID   `gorm:"type:uuid;not null"`
	User          User        `gorm:"foreignKey:UserID"`
	InstitutionID uuid.UUID   `gorm:"type:uuid;not null"`
	Institution   Institution `gorm:"foreignKey:InstitutionID"`
	TotalDays     int         `gorm:"not null"`
	UsedDays      int         `gorm:"default:0"`
}

// LeaveRequest Model
type LeaveRequest struct {
	Base

	UserID        uuid.UUID   `gorm:"type:uuid;not null"`
	User          User        `gorm:"foreignKey:UserID"`
	InstitutionID uuid.UUID   `gorm:"type:uuid;not null"`
	Institution   Institution `gorm:"foreignKey:InstitutionID"`
	StartDate     time.Time   `gorm:"not null"`
	EndDate       time.Time   `gorm:"not null"`
	Reason        string      `gorm:"type:text"`
	Status        LeaveStatus `gorm:"type:varchar(50);not null;default:'pending'"`
	ReviewedBy    *uuid.UUID  `gorm:"type:uuid"`
	ReviewerUser  *User       `gorm:"foreignKey:ReviewedBy"`
	ReviewedAt    *time.Time
	ReviewNote    string `gorm:"type:text"`
}
