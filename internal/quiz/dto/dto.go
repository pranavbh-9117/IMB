// Package dto provides data transfer objects for the quiz module.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateQuizRequest represents the payload for creating a new quiz.
type CreateQuizRequest struct {
	Title           string `json:"title" validate:"required,max=255"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes" validate:"required,min=1"`
}

// UpdateQuizRequest represents the payload for updating an existing draft quiz.
type UpdateQuizRequest struct {
	Title           string `json:"title" validate:"required,max=255"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes" validate:"required,min=1"`
}

// OptionRequest represents the payload for a single question option.
type OptionRequest struct {
	Text      string `json:"text" validate:"required,max=255"`
	IsCorrect bool   `json:"is_correct"`
}

// CreateQuestionRequest represents the payload for creating a new question within a quiz.
type CreateQuestionRequest struct {
	Text    string          `json:"text" validate:"required"`
	Marks   int             `json:"marks" validate:"required,min=1"`
	Options []OptionRequest `json:"options" validate:"required,min=2"`
}

// QuizResponse represents the response payload for a quiz.
type QuizResponse struct {
	ID              uuid.UUID          `json:"id"`
	InstitutionID   uuid.UUID          `json:"institution_id"`
	CreatedBy       uuid.UUID          `json:"created_by"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	DurationMinutes int                `json:"duration_minutes"`
	TotalMarks      int                `json:"total_marks"`
	IsPublished     bool               `json:"is_published"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	Questions       []QuestionResponse `json:"questions,omitempty"`
}

// QuestionResponse represents the response payload for a question.
type QuestionResponse struct {
	ID         uuid.UUID        `json:"id"`
	QuizID     uuid.UUID        `json:"quiz_id"`
	Text       string           `json:"text"`
	Marks      int              `json:"marks"`
	OrderIndex int              `json:"order_index"`
	Options    []OptionResponse `json:"options"`
}

// OptionResponse represents the response payload for an option.
type OptionResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"is_correct,omitempty"` // May be hidden from students depending on endpoint
	OrderIndex int       `json:"order_index"`
}
