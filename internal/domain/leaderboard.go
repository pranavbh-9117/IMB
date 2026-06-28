// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuizLeaderboardEntry represents the persistent domain entity mapped to the quiz_leaderboard table.
type QuizLeaderboardEntry struct {
	Base
	QuizID      uuid.UUID   `gorm:"type:uuid;not null;index"`
	Quiz        Quiz        `gorm:"foreignKey:QuizID"`
	StudentID   uuid.UUID   `gorm:"type:uuid;not null"`
	Student     User        `gorm:"foreignKey:StudentID"`
	AttemptID   uuid.UUID   `gorm:"type:uuid;not null"`
	Score       int         `gorm:"not null;default:0"`
	TotalMarks  int         `gorm:"not null;default:0"`
	Percentage  float64     `gorm:"not null;default:0.0"`
	SubmittedAt time.Time   `gorm:"not null"`
}

// QuizLeaderboardRankedEntry represents a read projection returned exclusively by leaderboard SQL queries.
// It is not a persisted entity.
type QuizLeaderboardRankedEntry struct {
	Rank        int
	StudentID   uuid.UUID
	StudentName string
	Score       int
	TotalMarks  int
	Percentage  float64
	SubmittedAt time.Time
}
