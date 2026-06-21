// Package service provides service functionality for the quiz module.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/quiz/dto"
	"github.com/pranavbh-9117/IMB/internal/quiz/repository"
)

type quizService struct {
	repo repository.QuizRepository
}

// NewQuizService creates a new QuizService.
func NewQuizService(repo repository.QuizRepository) QuizService {
	return &quizService{repo: repo}
}

// CreateQuiz initializes a new draft quiz.
func (s *quizService) CreateQuiz(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, req *dto.CreateQuizRequest) (*dto.QuizResponse, error) {
	quiz := &domain.Quiz{
		InstitutionID:   institutionID,
		CreatedBy:       facultyID,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		IsPublished:     false,
		TotalMarks:      0,
	}

	if err := s.repo.CreateQuiz(ctx, quiz); err != nil {
		return nil, fmt.Errorf("quiz service: create: %w", err)
	}

	return s.mapQuizToResponse(quiz, nil, nil), nil
}

// GetQuiz retrieves a quiz. Enforces tenant boundaries.
func (s *quizService) GetQuiz(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID) (*dto.QuizResponse, error) {
	quiz, questions, options, err := s.repo.GetQuizWithQuestions(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQuizNotFound
		}
		return nil, fmt.Errorf("quiz service: get: %w", err)
	}

	// Enforce Multi-Tenant rule
	if quiz.InstitutionID != institutionID {
		return nil, ErrQuizNotFound // Lie to prevent leakage
	}

	// Enforce Ownership & Publishing Rules
	if callerRole == domain.RoleFaculty {
		if quiz.CreatedBy != callerID {
			return nil, ErrQuizNotFound
		}
	} else if callerRole == domain.RoleStudent {
		if !quiz.IsPublished {
			return nil, ErrQuizNotFound
		}
	} else {
		return nil, ErrQuizNotFound // Institute/Super Admins cannot view quizzes
	}

	return s.mapQuizToResponse(quiz, questions, options), nil
}

// UpdateQuiz modifies a draft quiz.
func (s *quizService) UpdateQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID, req *dto.UpdateQuizRequest) error {
	quiz, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrQuizNotFound
		}
		return fmt.Errorf("quiz service: update fetch: %w", err)
	}

	if quiz.CreatedBy != facultyID {
		return ErrUnauthorizedQuiz
	}

	if quiz.IsPublished {
		return ErrQuizAlreadyPublished
	}

	quiz.Title = req.Title
	quiz.Description = req.Description
	quiz.DurationMinutes = req.DurationMinutes

	if err := s.repo.UpdateQuiz(ctx, quiz); err != nil {
		return fmt.Errorf("quiz service: update save: %w", err)
	}

	return nil
}

// DeleteQuiz deletes a quiz if it has no attempts.
func (s *quizService) DeleteQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID) error {
	quiz, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrQuizNotFound
		}
		return fmt.Errorf("quiz service: delete fetch: %w", err)
	}

	if quiz.CreatedBy != facultyID {
		return ErrUnauthorizedQuiz
	}

	hasAttempts, err := s.repo.HasAttempts(ctx, quizID)
	if err != nil {
		return fmt.Errorf("quiz service: delete check attempts: %w", err)
	}
	if hasAttempts {
		return ErrQuizHasAttempts
	}

	if err := s.repo.DeleteQuiz(ctx, quizID); err != nil {
		return fmt.Errorf("quiz service: delete execute: %w", err)
	}

	return nil
}

// PublishQuiz flips the IsPublished flag, making it visible to students and locking edits.
func (s *quizService) PublishQuiz(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID) error {
	quiz, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrQuizNotFound
		}
		return fmt.Errorf("quiz service: publish fetch: %w", err)
	}

	if quiz.CreatedBy != facultyID {
		return ErrUnauthorizedQuiz
	}

	quiz.IsPublished = true

	if err := s.repo.UpdateQuiz(ctx, quiz); err != nil {
		return fmt.Errorf("quiz service: publish save: %w", err)
	}

	return nil
}

// ListQuizzes returns quizzes based on the caller's role.
func (s *quizService) ListQuizzes(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role) ([]dto.QuizResponse, error) {
	var quizzes []domain.Quiz
	var err error

	if callerRole == domain.RoleFaculty {
		quizzes, err = s.repo.ListQuizzes(ctx, institutionID, &callerID, false)
	} else if callerRole == domain.RoleStudent {
		quizzes, err = s.repo.ListQuizzes(ctx, institutionID, nil, true)
	} else {
		return []dto.QuizResponse{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("quiz service: list: %w", err)
	}

	var responses []dto.QuizResponse
	for _, q := range quizzes {
		responses = append(responses, *s.mapQuizToResponse(&q, nil, nil))
	}

	return responses, nil
}

// CreateQuestion adds a question to a draft quiz.
func (s *quizService) CreateQuestion(ctx context.Context, facultyID uuid.UUID, quizID uuid.UUID, req *dto.CreateQuestionRequest) (*dto.QuestionResponse, error) {
	quiz, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQuizNotFound
		}
		return nil, fmt.Errorf("quiz service: create question fetch quiz: %w", err)
	}

	if quiz.CreatedBy != facultyID {
		return nil, ErrUnauthorizedQuiz
	}

	if quiz.IsPublished {
		return nil, ErrQuizAlreadyPublished
	}

	// Validate Options
	correctCount := 0
	for _, opt := range req.Options {
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return nil, ErrInvalidOptions
	}

	question := &domain.Question{
		QuizID:     quizID,
		Text:       req.Text,
		Marks:      req.Marks,
		OrderIndex: 0, // Simplified for now
	}

	var options []domain.Option
	for i, optReq := range req.Options {
		options = append(options, domain.Option{
			Text:       optReq.Text,
			IsCorrect:  optReq.IsCorrect,
			OrderIndex: i,
		})
	}

	if err := s.repo.CreateQuestion(ctx, question, options); err != nil {
		return nil, fmt.Errorf("quiz service: create question save: %w", err)
	}

	// Update TotalMarks
	newTotal := quiz.TotalMarks + question.Marks
	if err := s.repo.UpdateQuizTotalMarks(ctx, quizID, newTotal); err != nil {
		return nil, fmt.Errorf("quiz service: update total marks: %w", err)
	}

	// Construct Response
	qRes := &dto.QuestionResponse{
		ID:         question.ID,
		QuizID:     question.QuizID,
		Text:       question.Text,
		Marks:      question.Marks,
		OrderIndex: question.OrderIndex,
	}

	for _, o := range options {
		qRes.Options = append(qRes.Options, dto.OptionResponse{
			ID:         o.ID,
			QuestionID: o.QuestionID,
			Text:       o.Text,
			IsCorrect:  o.IsCorrect,
			OrderIndex: o.OrderIndex,
		})
	}

	return qRes, nil
}

// GetQuizForEvaluation retrieves a quiz with all correct answers explicitly visible.
// This is strictly for internal service-to-service communication (e.g., AttemptService).
func (s *quizService) GetQuizForEvaluation(ctx context.Context, quizID uuid.UUID) (*dto.QuizResponse, error) {
	quiz, questions, options, err := s.repo.GetQuizWithQuestions(ctx, quizID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQuizNotFound
		}
		return nil, fmt.Errorf("quiz service: get for evaluation: %w", err)
	}

	return s.mapQuizToResponse(quiz, questions, options), nil
}

// mapQuizToResponse is a helper to assemble the DTO.
func (s *quizService) mapQuizToResponse(quiz *domain.Quiz, questions []domain.Question, options []domain.Option) *dto.QuizResponse {
	res := &dto.QuizResponse{
		ID:              quiz.ID,
		InstitutionID:   quiz.InstitutionID,
		CreatedBy:       quiz.CreatedBy,
		Title:           quiz.Title,
		Description:     quiz.Description,
		DurationMinutes: quiz.DurationMinutes,
		TotalMarks:      quiz.TotalMarks,
		IsPublished:     quiz.IsPublished,
		CreatedAt:       quiz.CreatedAt,
		UpdatedAt:       quiz.UpdatedAt,
	}

	if len(questions) > 0 {
		qMap := make(map[uuid.UUID]*dto.QuestionResponse)
		for _, q := range questions {
			qRes := &dto.QuestionResponse{
				ID:         q.ID,
				QuizID:     q.QuizID,
				Text:       q.Text,
				Marks:      q.Marks,
				OrderIndex: q.OrderIndex,
				Options:    []dto.OptionResponse{},
			}
			qMap[q.ID] = qRes
			res.Questions = append(res.Questions, *qRes)
		}

		for _, o := range options {
			if qRes, exists := qMap[o.QuestionID]; exists {
				qRes.Options = append(qRes.Options, dto.OptionResponse{
					ID:         o.ID,
					QuestionID: o.QuestionID,
					Text:       o.Text,
					IsCorrect:  o.IsCorrect,
					OrderIndex: o.OrderIndex,
				})
			}
		}

		// Overwrite res.Questions with the updated pointers
		res.Questions = nil
		for _, q := range questions {
			res.Questions = append(res.Questions, *qMap[q.ID])
		}
	}

	return res
}
