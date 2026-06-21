// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Base holds the attributes which are shared by all domains
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
