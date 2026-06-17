package domain

import (
	"time"

	"github.com/google/uuid"
)

// LeaveStatus represents the current approval state of a leave request.
type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "pending"
	LeaveStatusApproved LeaveStatus = "approved"
	LeaveStatusRejected LeaveStatus = "rejected"
)

// LeaveBalance tracks the total allocated leave days and how many have been
// consumed for a given user within their institution. Leave is not
// categorised by type.
type LeaveBalance struct {
	Base

	UserID        uuid.UUID   `gorm:"type:uuid;not null"`
	User          User        `gorm:"foreignKey:UserID"`
	InstitutionID uuid.UUID   `gorm:"type:uuid;not null"`
	Institution   Institution `gorm:"foreignKey:InstitutionID"`
	TotalDays     int         `gorm:"not null"`
	UsedDays      int         `gorm:"default:0"`
}

// LeaveRequest is a formal application for leave submitted by a user.
// ReviewedBy and ReviewedAt remain nil until an authorised reviewer acts on
// the request.
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
