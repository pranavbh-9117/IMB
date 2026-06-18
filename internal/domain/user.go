// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role is the access-control designation for a platform user. It is stored
// as a VARCHAR column and enforced at the application layer.
type Role string

const (
	RoleSuperAdmin     Role = "super_admin"
	RoleInstituteAdmin Role = "institute_admin"
	RoleFaculty        Role = "faculty"
	RoleStudent        Role = "student"
)

// User represents a human actor on the platform. A super_admin has no
// institution affiliation (InstitutionID is nil). All other roles belong to
// exactly one institution. An account may be created via email+password or
// Google OAuth; both paths share the same entity.
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

// RefreshToken stores a hashed refresh token issued to a user. The raw token
// is never persisted; only its SHA-256 hash is stored. Tokens can be
// explicitly revoked or expire naturally.
type RefreshToken struct {
	Base

	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	User      User      `gorm:"foreignKey:UserID"`
	TokenHash string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
}
