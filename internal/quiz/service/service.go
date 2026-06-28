// Package service provides service functionality for the quiz module.
package service

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/quiz/dto"
	"github.com/pranavbh-9117/IMB/pkg/apperror"
)

var (
	ErrQuizNotFound         = apperror.NotFound("quiz not found")
	ErrUnauthorizedQuiz     = apperror.Forbidden("you do not have permission to manage this quiz")
	ErrQuizAlreadyPublished = apperror.Conflict("published quizzes cannot be modified")
	ErrQuizHasAttempts      = apperror.Conflict("quiz cannot be deleted because it has registered attempts")
	ErrInvalidOptions       = apperror.BadRequest("a question must have exactly one correct option")
)

// QuizService defines the interface for the quiz business logic.
type QuizService interface {
	CreateQuiz(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, req *dto.CreateQuizRequest) (*dto.QuizResponse, error)
	GetQuiz(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID) (*dto.QuizResponse, error)
	UpdateQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID, req *dto.UpdateQuizRequest) error
	DeleteQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID) error
	PublishQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID) error
	ListQuizzes(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role) ([]dto.QuizResponse, error)
	CreateQuestion(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID, req *dto.CreateQuestionRequest) (*dto.QuestionResponse, error)
	GetQuizForEvaluation(ctx context.Context, quizID uuid.UUID) (*dto.QuizResponse, error)
}

// MaterialService defines the interface for quiz material uploads and downloads.
type MaterialService interface {
	UploadMaterials(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, quizID uuid.UUID, files []*multipart.FileHeader) (*dto.UploadMaterialsResponse, error)
	ListMaterials(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID) ([]dto.MaterialResponse, error)
	DownloadMaterial(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID, materialID uuid.UUID) (*dto.DownloadResult, error)
}

