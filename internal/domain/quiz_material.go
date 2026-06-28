package domain

import (
	"github.com/google/uuid"
)

// QuizMaterial represents supplementary files attached to a quiz.
type QuizMaterial struct {
	Base
	QuizID           uuid.UUID `gorm:"type:uuid;not null;index"`
	Quiz             Quiz      `gorm:"foreignKey:QuizID"`
	UploadedBy       uuid.UUID `gorm:"type:uuid;not null;index"`
	Uploader         User      `gorm:"foreignKey:UploadedBy"`
	OriginalFilename string    `gorm:"type:varchar(255);not null"`
	StoredFilename   string    `gorm:"type:varchar(255);not null"`
	StoragePath      string    `gorm:"type:text;not null"`
	ContentType      string    `gorm:"type:varchar(100);not null"`
	FileSize         int64     `gorm:"not null"`
}
