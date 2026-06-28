// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuizAttempt Model
type QuizAttempt struct {
	Base
	InstitutionID   uuid.UUID   `gorm:"type:uuid;not null;index"`
	Institution     Institution `gorm:"foreignKey:InstitutionID"`
	QuizID          uuid.UUID   `gorm:"type:uuid;not null;index"`
	Quiz            Quiz        `gorm:"foreignKey:QuizID"`
	StudentID       uuid.UUID   `gorm:"type:uuid;not null;index"`
	Student         User        `gorm:"foreignKey:StudentID"`
	StartedAt       time.Time   `gorm:"not null"`
	SubmittedAt     *time.Time
	Score           int         `gorm:"not null;default:0"`
	TotalMarks      int         `gorm:"not null;default:0"`
	Percentage      float64     `gorm:"not null;default:0.0"`
}

// QuizAnswer Model
type QuizAnswer struct {
	Base
	AttemptID        uuid.UUID   `gorm:"type:uuid;not null;index"`
	Attempt          QuizAttempt `gorm:"foreignKey:AttemptID"`
	QuestionID       uuid.UUID   `gorm:"type:uuid;not null;index"`
	Question         Question    `gorm:"foreignKey:QuestionID"`
	SelectedOptionID *uuid.UUID  `gorm:"type:uuid"` 
	SelectedOption   *Option     `gorm:"foreignKey:SelectedOptionID"`
}
