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
	"github.com/pranavbh-9117/IMB/pkg/email"
)

type attemptService struct {
	repo    repository.AttemptRepository
	quizSvc quizservice.QuizService
	emailSvc email.EmailService
}

func NewAttemptService(repo repository.AttemptRepository, quizSvc quizservice.QuizService, emailSvc email.EmailService) AttemptService {
	return &attemptService{
		repo:     repo,
		quizSvc:  quizSvc,
		emailSvc: emailSvc,
	}
}

// SubmitAttempt validates and grades a student's submission atomically.
func (s *attemptService) SubmitAttempt(ctx context.Context, institutionID uuid.UUID, studentID uuid.UUID, quizID uuid.UUID, req *dto.SubmitAttemptRequest) (*dto.SubmitResultResponse, error) {
	quizRes, err := s.quizSvc.GetQuizForEvaluation(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotAvailable
	}

	if quizRes.InstitutionID != institutionID || !quizRes.IsPublished {
		return nil, ErrQuizNotAvailable
	}

	hasAttempted, err := s.repo.HasAttempted(ctx, studentID, quizID)
	if err != nil {
		return nil, fmt.Errorf("attempt service: check attempts: %w", err)
	}
	if hasAttempted {
		return nil, ErrAlreadyAttempted
	}

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

	totalScore := 0
	var domainAnswers []domain.QuizAnswer

	for _, sub := range req.Answers {
		marks, exists := questionMarks[sub.QuestionID]
		if !exists {
			return nil, ErrInvalidSubmission
		}

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

	percentage := 0.0
	if quizRes.TotalMarks > 0 {
		percentage = (float64(totalScore) / float64(quizRes.TotalMarks)) * 100.0
	}

	now := time.Now()
	attempt := &domain.QuizAttempt{
		InstitutionID: institutionID,
		QuizID:        quizID,
		StudentID:     studentID,
		StartedAt:     now,
		SubmittedAt:   &now,
		Score:         0,
		TotalMarks:    quizRes.TotalMarks,
		Percentage:    0.0,
	}

	err = s.repo.DoInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateAttempt(txCtx, attempt); err != nil {
			return fmt.Errorf("create attempt: %w", err)
		}

		if err := s.repo.BulkCreateAnswers(txCtx, attempt.ID, domainAnswers); err != nil {
			return fmt.Errorf("bulk create answers: %w", err)
		}

		if err := s.repo.UpdateAttemptResult(txCtx, attempt.ID, totalScore, percentage); err != nil {
			return fmt.Errorf("update attempt result: %w", err)
		}

		leaderboardEntry := &domain.QuizLeaderboardEntry{
			QuizID:      quizID,
			StudentID:   studentID,
			AttemptID:   attempt.ID,
			Score:       totalScore,
			TotalMarks:  quizRes.TotalMarks,
			Percentage:  percentage,
			SubmittedAt: now,
		}
		if err := s.repo.UpsertLeaderboard(txCtx, leaderboardEntry); err != nil {
			return fmt.Errorf("upsert leaderboard: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("attempt service: submit transaction failed: %w", err)
	}

	rank, err := s.repo.GetStudentRank(ctx, quizID, studentID)
	if err != nil {
		return nil, fmt.Errorf("attempt service: post-commit get rank: %w", err)
	}

	res := &dto.SubmitResultResponse{
		AttemptID:       attempt.ID,
		QuizID:          quizID,
		Score:           totalScore,
		TotalMarks:      quizRes.TotalMarks,
		Percentage:      percentage,
		LeaderboardRank: rank,
		SubmittedAt:     now,
	}

	if s.emailSvc != nil {
		studentEmail, _ := s.repo.GetStudentEmail(ctx, studentID)
		if studentEmail != "" {
			subject := fmt.Sprintf("Quiz Submission Confirmation: %s", quizRes.Title)
			body := fmt.Sprintf("Hello,\n\nYou have successfully submitted the quiz '%s'.\n\nScore: %d/%d (%.2f%%)\nLeaderboard Rank: %d\nSubmitted At: %s\n\nBest regards,\nIMB Platform",
				quizRes.Title, totalScore, quizRes.TotalMarks, percentage, rank, now.Format(time.RFC1123))
			s.emailSvc.SendAsync(ctx, email.Message{
				To:      studentEmail,
				Subject: subject,
				Body:    body,
			})
		}
	}

	return res, nil
}

// GetLeaderboard retrieves the ranked leaderboard projections for a quiz.
func (s *attemptService) GetLeaderboard(ctx context.Context, institutionID uuid.UUID, quizID uuid.UUID) (*dto.LeaderboardResponse, error) {
	quizRes, err := s.quizSvc.GetQuizForEvaluation(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotAvailable
	}
	if quizRes.InstitutionID != institutionID {
		return nil, ErrUnauthorized
	}

	projections, err := s.repo.GetLeaderboard(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("attempt service: get leaderboard: %w", err)
	}

	entries := make([]dto.LeaderboardEntryResponse, len(projections))
	for i, p := range projections {
		entries[i] = dto.LeaderboardEntryResponse{
			Rank:        p.Rank,
			StudentName: p.StudentName,
			Score:       p.Score,
			Percentage:  p.Percentage,
			SubmittedAt: p.SubmittedAt,
		}
	}

	return &dto.LeaderboardResponse{
		QuizID:  quizID,
		Entries: entries,
	}, nil
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
