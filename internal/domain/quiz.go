// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Quiz Model
type Quiz struct {
	Base
	InstitutionID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Institution     Institution    `gorm:"foreignKey:InstitutionID"`
	CreatedBy       uuid.UUID      `gorm:"type:uuid;not null;index"`
	Creator         User           `gorm:"foreignKey:CreatedBy"`
	Title           string         `gorm:"type:varchar(255);not null"`
	Description     string         `gorm:"type:text"`
	DurationMinutes int            `gorm:"not null"`
	TotalMarks      int            `gorm:"not null;default:0"`
	IsPublished     bool           `gorm:"not null;default:false"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// Quiz Question Model
type Question struct {
	Base
	QuizID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Quiz       Quiz      `gorm:"foreignKey:QuizID"`
	Text       string    `gorm:"type:text;not null"`
	Marks      int       `gorm:"not null"`
	OrderIndex int       `gorm:"not null"`
}

// Quiz Questions Option Model
type Option struct {
	Base
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index"`
	Question   Question  `gorm:"foreignKey:QuestionID"`
	Text       string    `gorm:"type:varchar(255);not null"`
	IsCorrect  bool      `gorm:"not null;default:false"`
	OrderIndex int       `gorm:"not null"`
}
