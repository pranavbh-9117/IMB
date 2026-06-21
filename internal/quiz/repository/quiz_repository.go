// Package repository implements data access patterns for quizzes.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

type quizRepository struct {
	db *gorm.DB
}

// NewQuizRepository creates a new instance of QuizRepository.
func NewQuizRepository(db *gorm.DB) QuizRepository {
	return &quizRepository{db: db}
}

// CreateQuiz inserts a new quiz into the database.
func (r *quizRepository) CreateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	if err := r.db.WithContext(ctx).Create(quiz).Error; err != nil {
		return fmt.Errorf("quiz repository: create: %w", err)
	}
	return nil
}

// GetQuizByID retrieves a single quiz by its ID.
func (r *quizRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	err := r.db.WithContext(ctx).First(&quiz, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("quiz repository: get by id: %w", err)
	}
	return &quiz, nil
}

// GetQuizWithQuestions retrieves a quiz along with its questions and options.
func (r *quizRepository) GetQuizWithQuestions(ctx context.Context, id uuid.UUID) (*domain.Quiz, []domain.Question, []domain.Option, error) {
	quiz, err := r.GetQuizByID(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}

	var questions []domain.Question
	if err := r.db.WithContext(ctx).Where("quiz_id = ?", id).Order("order_index ASC").Find(&questions).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("quiz repository: get questions: %w", err)
	}

	var options []domain.Option
	if len(questions) > 0 {
		var questionIDs []uuid.UUID
		for _, q := range questions {
			questionIDs = append(questionIDs, q.ID)
		}
		if err := r.db.WithContext(ctx).Where("question_id IN ?", questionIDs).Order("order_index ASC").Find(&options).Error; err != nil {
			return nil, nil, nil, fmt.Errorf("quiz repository: get options: %w", err)
		}
	}

	return quiz, questions, options, nil
}

// UpdateQuiz modifies an existing quiz metadata.
func (r *quizRepository) UpdateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	if err := r.db.WithContext(ctx).Save(quiz).Error; err != nil {
		return fmt.Errorf("quiz repository: update: %w", err)
	}
	return nil
}

// DeleteQuiz deletes a quiz from the database.
func (r *quizRepository) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Quiz{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("quiz repository: delete: %w", err)
	}
	return nil
}

// ListQuizzes retrieves quizzes based on institution and optional faculty/published filters.
func (r *quizRepository) ListQuizzes(ctx context.Context, institutionID uuid.UUID, facultyID *uuid.UUID, publishedOnly bool) ([]domain.Quiz, error) {
	var quizzes []domain.Quiz
	query := r.db.WithContext(ctx).Where("institution_id = ?", institutionID)

	if facultyID != nil {
		query = query.Where("created_by = ?", *facultyID)
	}

	if publishedOnly {
		query = query.Where("is_published = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&quizzes).Error; err != nil {
		return nil, fmt.Errorf("quiz repository: list: %w", err)
	}
	return quizzes, nil
}

// CreateQuestion inserts a new question and its options atomically.
func (r *quizRepository) CreateQuestion(ctx context.Context, question *domain.Question, options []domain.Option) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(question).Error; err != nil {
			return fmt.Errorf("quiz repository: create question: %w", err)
		}

		for i := range options {
			options[i].QuestionID = question.ID
		}

		if err := tx.Create(&options).Error; err != nil {
			return fmt.Errorf("quiz repository: create options: %w", err)
		}

		return nil
	})
}

// UpdateQuizTotalMarks safely updates the total marks of a quiz.
func (r *quizRepository) UpdateQuizTotalMarks(ctx context.Context, quizID uuid.UUID, newTotal int) error {
	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("id = ?", quizID).Update("total_marks", newTotal).Error; err != nil {
		return fmt.Errorf("quiz repository: update total marks: %w", err)
	}
	return nil
}

// HasAttempts checks if a quiz has any registered attempts.
func (r *quizRepository) HasAttempts(ctx context.Context, quizID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).Where("quiz_id = ?", quizID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("quiz repository: check attempts: %w", err)
	}
	return count > 0, nil
}
