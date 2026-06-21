// Package dto provides data transfer objects for the attempt module.
package dto

import (
	"time"

	"github.com/google/uuid"
)


type AnswerSubmission struct {
	QuestionID       uuid.UUID  `json:"question_id" validate:"required"`
	SelectedOptionID *uuid.UUID `json:"selected_option_id"` 
}


type SubmitAttemptRequest struct {
	Answers []AnswerSubmission `json:"answers" validate:"required"`
}

type StudentResultResponse struct {
	AttemptID   uuid.UUID  `json:"attempt_id"`
	QuizID      uuid.UUID  `json:"quiz_id"`
	QuizTitle   string     `json:"quiz_title"`
	Score       int        `json:"score"`
	TotalMarks  int        `json:"total_marks"`
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

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
