// Package service provides service functionality for the attempt module.
package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/pkg/apperror"
)

// Sentinel errors
var (
	ErrQuizNotAvailable  = apperror.NotFound("quiz not found or not published")
	ErrAlreadyAttempted  = apperror.Conflict("you have already attempted this quiz")
	ErrInvalidSubmission = apperror.BadRequest("invalid submission payload")
	ErrUnauthorized      = apperror.Forbidden("you do not have permission to view these results")
)

// AttemptService defines the interface for attempt business logic.
type AttemptService interface {
	SubmitAttempt(ctx context.Context, institutionID uuid.UUID, studentID uuid.UUID, quizID uuid.UUID, req *dto.SubmitAttemptRequest) error
	GetStudentResults(ctx context.Context, studentID uuid.UUID) ([]dto.StudentResultResponse, error)
	GetQuizResults(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, quizID uuid.UUID) ([]dto.FacultyResultResponse, error)
}
