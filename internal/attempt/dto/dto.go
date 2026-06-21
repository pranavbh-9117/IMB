// Package dto provides data transfer objects for the attempt module.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// AnswerSubmission represents a single answer in a bulk submission.
type AnswerSubmission struct {
	QuestionID       uuid.UUID  `json:"question_id" validate:"required"`
	SelectedOptionID *uuid.UUID `json:"selected_option_id"` // Nullable if the student skips the question
}

// SubmitAttemptRequest represents the bulk payload for submitting a quiz attempt.
type SubmitAttemptRequest struct {
	Answers []AnswerSubmission `json:"answers" validate:"required"`
}

// StudentResultResponse represents a single result item in the student's history.
type StudentResultResponse struct {
	AttemptID   uuid.UUID  `json:"attempt_id"`
	QuizID      uuid.UUID  `json:"quiz_id"`
	QuizTitle   string     `json:"quiz_title"`
	Score       int        `json:"score"`
	TotalMarks  int        `json:"total_marks"`
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

// FacultyResultResponse represents a single result item for a specific quiz, viewed by faculty.
type FacultyResultResponse struct {
	AttemptID    uuid.UUID  `json:"attempt_id"`
	StudentID    uuid.UUID  `json:"student_id"`
	StudentName  string     `json:"student_name"`
	StudentEmail string     `json:"student_email"`
	Score        int        `json:"score"`
	TotalMarks   int        `json:"total_marks"`
	StartedAt    time.Time  `json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
}
