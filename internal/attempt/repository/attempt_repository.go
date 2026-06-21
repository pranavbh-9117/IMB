// Package repository implements data access patterns for quiz attempts.
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/domain"
)

type attemptRepository struct {
	db *gorm.DB
}

// NewAttemptRepository creates a new instance of AttemptRepository.
func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

// CreateAttempt inserts a new attempt and its answers transactionally.
func (r *attemptRepository) CreateAttempt(ctx context.Context, attempt *domain.QuizAttempt, answers []domain.QuizAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(attempt).Error; err != nil {
			return fmt.Errorf("attempt repository: create attempt: %w", err)
		}

		for i := range answers {
			answers[i].AttemptID = attempt.ID
		}

		if len(answers) > 0 {
			if err := tx.Create(&answers).Error; err != nil {
				return fmt.Errorf("attempt repository: create answers: %w", err)
			}
		}

		return nil
	})
}

// HasAttempted checks if a student has already started or submitted an attempt for this quiz.
func (r *attemptRepository) HasAttempted(ctx context.Context, studentID uuid.UUID, quizID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("student_id = ? AND quiz_id = ?", studentID, quizID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("attempt repository: check has attempted: %w", err)
	}
	return count > 0, nil
}

// GetStudentResults fetches all attempts by a student, joined with the quiz title.
func (r *attemptRepository) GetStudentResults(ctx context.Context, studentID uuid.UUID) ([]dto.StudentResultResponse, error) {
	var results []dto.StudentResultResponse

	query := `
		SELECT 
			a.id as attempt_id, 
			q.id as quiz_id, 
			q.title as quiz_title, 
			a.score, 
			a.total_marks, 
			a.started_at, 
			a.submitted_at
		FROM quiz_attempts a
		JOIN quizzes q ON a.quiz_id = q.id
		WHERE a.student_id = ?
		ORDER BY a.started_at DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, studentID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("attempt repository: get student results: %w", err)
	}

	return results, nil
}

// GetQuizResults fetches all attempts for a quiz, joined with student details.
func (r *attemptRepository) GetQuizResults(ctx context.Context, quizID uuid.UUID) ([]dto.FacultyResultResponse, error) {
	var results []dto.FacultyResultResponse

	query := `
		SELECT 
			a.id as attempt_id, 
			u.id as student_id, 
			u.name as student_name, 
			u.email as student_email, 
			a.score, 
			a.total_marks, 
			a.started_at, 
			a.submitted_at
		FROM quiz_attempts a
		JOIN users u ON a.student_id = u.id
		WHERE a.quiz_id = ?
		ORDER BY a.score DESC, a.started_at ASC
	`

	if err := r.db.WithContext(ctx).Raw(query, quizID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("attempt repository: get quiz results: %w", err)
	}

	return results, nil
}
