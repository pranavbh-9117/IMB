// Package repository implements data access patterns for quiz attempts.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/domain"
)

// ErrNotFound is returned when an attempt is not found.
var ErrNotFound = errors.New("record not found")

// AttemptRepository defines the interface for attempt data access.
type AttemptRepository interface {
	CreateAttempt(ctx context.Context, attempt *domain.QuizAttempt, answers []domain.QuizAnswer) error
	HasAttempted(ctx context.Context, studentID uuid.UUID, quizID uuid.UUID) (bool, error)
	GetStudentResults(ctx context.Context, studentID uuid.UUID) ([]dto.StudentResultResponse, error)
	GetQuizResults(ctx context.Context, quizID uuid.UUID) ([]dto.FacultyResultResponse, error)
}
