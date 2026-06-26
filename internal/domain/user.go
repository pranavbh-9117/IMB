// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Types of Roles a user can exist with
type Role string

const (
	RoleSuperAdmin     Role = "super_admin"
	RoleInstituteAdmin Role = "institute_admin"
	RoleFaculty        Role = "faculty"
	RoleStudent        Role = "student"
)

// User Model
type User struct {
	Base

	Name               string       `gorm:"type:varchar(255);not null"`
	Email              string       `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash       string       `gorm:"type:text"`
	Role               Role         `gorm:"type:varchar(50);not null"`
	InstitutionID      *uuid.UUID   `gorm:"type:uuid"`
	Institution        *Institution `gorm:"foreignKey:InstitutionID"`
	GoogleID           string       `gorm:"type:varchar(255)"`
	IsActive           bool         `gorm:"default:true"`
	MustChangePassword bool         `gorm:"default:false"`
}

// RefreshToken Model
type RefreshToken struct {
	Base

	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	User      User      `gorm:"foreignKey:UserID"`
	TokenHash string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
}

// PasswordResetToken Model
type PasswordResetToken struct {
	Base

	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	User      User      `gorm:"foreignKey:UserID"`
	TokenHash string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
}

