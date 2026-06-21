// Package service provides service functionality for the attempt module.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/attempt/repository"
	"github.com/pranavbh-9117/IMB/internal/domain"
	quizservice "github.com/pranavbh-9117/IMB/internal/quiz/service"
)

type attemptService struct {
	repo    repository.AttemptRepository
	quizSvc quizservice.QuizService
}

// NewAttemptService creates a new AttemptService.
func NewAttemptService(repo repository.AttemptRepository, quizSvc quizservice.QuizService) AttemptService {
	return &attemptService{
		repo:    repo,
		quizSvc: quizSvc,
	}
}

// SubmitAttempt validates and grades a student's submission.
func (s *attemptService) SubmitAttempt(ctx context.Context, institutionID uuid.UUID, studentID uuid.UUID, quizID uuid.UUID, req *dto.SubmitAttemptRequest) error {
	// 1. Fetch Quiz for Evaluation (bypasses standard role masking, returns raw structure)
	quizRes, err := s.quizSvc.GetQuizForEvaluation(ctx, quizID)
	if err != nil {
		return ErrQuizNotAvailable
	}

	// 2. Validate multi-tenant boundary and publish status
	if quizRes.InstitutionID != institutionID || !quizRes.IsPublished {
		return ErrQuizNotAvailable
	}

	// 3. Check for existing attempts
	hasAttempted, err := s.repo.HasAttempted(ctx, studentID, quizID)
	if err != nil {
		return fmt.Errorf("attempt service: check attempts: %w", err)
	}
	if hasAttempted {
		return ErrAlreadyAttempted
	}

	// 4. Build correct answer mapping for O(1) lookups
	// correctAnswers maps QuestionID -> OptionID
	correctAnswers := make(map[uuid.UUID]uuid.UUID)
	questionMarks := make(map[uuid.UUID]int)

	for _, q := range quizRes.Questions {
		questionMarks[q.ID] = q.Marks
		for _, o := range q.Options {
			if o.IsCorrect {
				correctAnswers[q.ID] = o.ID
				break
			}
		}
	}

	// 5. Evaluate Submission
	totalScore := 0
	var domainAnswers []domain.QuizAnswer

	for _, sub := range req.Answers {
		// Verify this question actually belongs to the quiz
		marks, exists := questionMarks[sub.QuestionID]
		if !exists {
			return ErrInvalidSubmission
		}

		// Check if correct
		if sub.SelectedOptionID != nil {
			if correctAnswers[sub.QuestionID] == *sub.SelectedOptionID {
				totalScore += marks
			}
		}

		domainAnswers = append(domainAnswers, domain.QuizAnswer{
			QuestionID:       sub.QuestionID,
			SelectedOptionID: sub.SelectedOptionID,
		})
	}

	// 6. Build the Domain Attempt
	now := time.Now()
	attempt := &domain.QuizAttempt{
		InstitutionID: institutionID,
		QuizID:        quizID,
		StudentID:     studentID,
		StartedAt:     now, // For bulk submission, started and submitted are effectively the same in Phase 2
		SubmittedAt:   &now,
		Score:         totalScore,
		TotalMarks:    quizRes.TotalMarks,
	}

	// 7. Save Transactionally
	if err := s.repo.CreateAttempt(ctx, attempt, domainAnswers); err != nil {
		return fmt.Errorf("attempt service: save attempt: %w", err)
	}

	return nil
}

// GetStudentResults returns the student's own attempts.
func (s *attemptService) GetStudentResults(ctx context.Context, studentID uuid.UUID) ([]dto.StudentResultResponse, error) {
	results, err := s.repo.GetStudentResults(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("attempt service: get student results: %w", err)
	}
	if results == nil {
		return []dto.StudentResultResponse{}, nil
	}
	return results, nil
}

// GetQuizResults returns all attempts for a quiz, verifying the caller is the quiz creator.
func (s *attemptService) GetQuizResults(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, quizID uuid.UUID) ([]dto.FacultyResultResponse, error) {
	// 1. Verify Ownership (GetQuiz handles ownership verification if callerRole = FACULTY)
	// We can use the standard GetQuiz method to enforce this.
	_, err := s.quizSvc.GetQuiz(ctx, institutionID, facultyID, domain.RoleFaculty, quizID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	results, err := s.repo.GetQuizResults(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("attempt service: get quiz results: %w", err)
	}
	if results == nil {
		return []dto.FacultyResultResponse{}, nil
	}
	return results, nil
}
