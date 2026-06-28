// Package repository implements data access patterns for quiz attempts.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/domain"
)


var ErrNotFound = errors.New("record not found")

// AttemptRepository defines the interface for attempt data access.
type AttemptRepository interface {
	DoInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	CreateAttempt(ctx context.Context, attempt *domain.QuizAttempt) error
	BulkCreateAnswers(ctx context.Context, attemptID uuid.UUID, answers []domain.QuizAnswer) error
	UpdateAttemptResult(ctx context.Context, attemptID uuid.UUID, score int, percentage float64) error
	UpsertLeaderboard(ctx context.Context, entry *domain.QuizLeaderboardEntry) error
	GetStudentRank(ctx context.Context, quizID uuid.UUID, studentID uuid.UUID) (int, error)
	GetStudentEmail(ctx context.Context, studentID uuid.UUID) (string, error)
	GetLeaderboard(ctx context.Context, quizID uuid.UUID) ([]domain.QuizLeaderboardRankedEntry, error)
	HasAttempted(ctx context.Context, studentID uuid.UUID, quizID uuid.UUID) (bool, error)
	GetStudentResults(ctx context.Context, studentID uuid.UUID) ([]dto.StudentResultResponse, error)
	GetQuizResults(ctx context.Context, quizID uuid.UUID) ([]dto.FacultyResultResponse, error)
	GetInstitutionQuizStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) (int, int, error)
	GetTopStudentsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time, limit int) ([]domain.TopStudentEntry, error)
}
