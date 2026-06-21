// Package repository implements data access patterns for quizzes.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

var ErrNotFound = errors.New("record not found")

// QuizRepository defines the interface for quiz data access.
type QuizRepository interface {
	CreateQuiz(ctx context.Context, quiz *domain.Quiz) error

	GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error)

	GetQuizWithQuestions(ctx context.Context, id uuid.UUID) (*domain.Quiz, []domain.Question, []domain.Option, error)

	UpdateQuiz(ctx context.Context, quiz *domain.Quiz) error

	DeleteQuiz(ctx context.Context, id uuid.UUID) error

	ListQuizzes(ctx context.Context, institutionID uuid.UUID, facultyID *uuid.UUID, publishedOnly bool) ([]domain.Quiz, error)

	CreateQuestion(ctx context.Context, question *domain.Question, options []domain.Option) error

	UpdateQuizTotalMarks(ctx context.Context, quizID uuid.UUID, newTotal int) error

	HasAttempts(ctx context.Context, quizID uuid.UUID) (bool, error)
}
