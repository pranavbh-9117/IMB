package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// MaterialRepository defines data access operations for quiz materials.
type MaterialRepository interface {
	CreateBatch(ctx context.Context, materials []domain.QuizMaterial) error
	GetByQuizID(ctx context.Context, quizID uuid.UUID) ([]domain.QuizMaterial, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.QuizMaterial, error)
	CountByQuizID(ctx context.Context, quizID uuid.UUID) (int64, error)
}

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) CreateBatch(ctx context.Context, materials []domain.QuizMaterial) error {
	if len(materials) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&materials).Error; err != nil {
			return fmt.Errorf("material repository: create batch: %w", err)
		}
		return nil
	})
}

func (r *materialRepository) GetByQuizID(ctx context.Context, quizID uuid.UUID) ([]domain.QuizMaterial, error) {
	var materials []domain.QuizMaterial
	if err := r.db.WithContext(ctx).Where("quiz_id = ?", quizID).Order("created_at ASC").Find(&materials).Error; err != nil {
		return nil, fmt.Errorf("material repository: get by quiz id: %w", err)
	}
	return materials, nil
}

func (r *materialRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuizMaterial, error) {
	var material domain.QuizMaterial
	err := r.db.WithContext(ctx).First(&material, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("material repository: get by id: %w", err)
	}
	return &material, nil
}

func (r *materialRepository) CountByQuizID(ctx context.Context, quizID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.QuizMaterial{}).Where("quiz_id = ?", quizID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("material repository: count by quiz id: %w", err)
	}
	return count, nil
}
