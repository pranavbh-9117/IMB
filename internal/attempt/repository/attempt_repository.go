// Package repository implements data access patterns for quiz attempts.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

type txKey struct{}

type attemptRepository struct {
	db *gorm.DB
}


func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return database.GetSession(ctx, r.db)
}

func (r *attemptRepository) DoInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

func (r *attemptRepository) CreateAttempt(ctx context.Context, attempt *domain.QuizAttempt) error {
	if err := r.getDB(ctx).Create(attempt).Error; err != nil {
		return fmt.Errorf("attempt repository: create attempt: %w", err)
	}
	return nil
}

func (r *attemptRepository) BulkCreateAnswers(ctx context.Context, attemptID uuid.UUID, answers []domain.QuizAnswer) error {
	for i := range answers {
		answers[i].AttemptID = attemptID
	}
	if len(answers) == 0 {
		return nil
	}
	if err := r.getDB(ctx).CreateInBatches(answers, 100).Error; err != nil {
		return fmt.Errorf("attempt repository: bulk create answers: %w", err)
	}
	return nil
}

func (r *attemptRepository) UpdateAttemptResult(ctx context.Context, attemptID uuid.UUID, score int, percentage float64) error {
	res := r.getDB(ctx).Model(&domain.QuizAttempt{}).Where("id = ?", attemptID).Updates(map[string]interface{}{
		"score":      score,
		"percentage": percentage,
	})
	if res.Error != nil {
		return fmt.Errorf("attempt repository: update attempt result: %w", res.Error)
	}
	return nil
}

func (r *attemptRepository) UpsertLeaderboard(ctx context.Context, entry *domain.QuizLeaderboardEntry) error {
	err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "quiz_id"}, {Name: "student_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"attempt_id", "score", "total_marks", "percentage", "submitted_at", "updated_at"}),
	}).Create(entry).Error
	if err != nil {
		return fmt.Errorf("attempt repository: upsert leaderboard: %w", err)
	}
	return nil
}

func (r *attemptRepository) GetStudentRank(ctx context.Context, quizID uuid.UUID, studentID uuid.UUID) (int, error) {
	var rank int
	query := `
		SELECT rank FROM (
			SELECT student_id, RANK() OVER (ORDER BY score DESC, submitted_at ASC, student_id ASC) AS rank
			FROM quiz_leaderboard WHERE quiz_id = ?
		) ranked WHERE student_id = ?
	`
	if err := r.getDB(ctx).Raw(query, quizID, studentID).Scan(&rank).Error; err != nil {
		return 0, fmt.Errorf("attempt repository: get student rank: %w", err)
	}
	return rank, nil
}

func (r *attemptRepository) GetStudentEmail(ctx context.Context, studentID uuid.UUID) (string, error) {
	var email string
	if err := r.getDB(ctx).Table("users").Select("email").Where("id = ?", studentID).Scan(&email).Error; err != nil {
		return "", fmt.Errorf("attempt repository: get student email: %w", err)
	}
	return email, nil
}

func (r *attemptRepository) GetLeaderboard(ctx context.Context, quizID uuid.UUID) ([]domain.QuizLeaderboardRankedEntry, error) {
	var entries []domain.QuizLeaderboardRankedEntry
	query := `
		SELECT 
			l.student_id,
			u.name AS student_name,
			l.score,
			l.total_marks,
			l.percentage,
			l.submitted_at,
			RANK() OVER (ORDER BY l.score DESC, l.submitted_at ASC, l.student_id ASC) AS rank
		FROM quiz_leaderboard l
		JOIN users u ON u.id = l.student_id
		WHERE l.quiz_id = ?
		ORDER BY rank ASC, l.student_id ASC
	`
	if err := r.getDB(ctx).Raw(query, quizID).Scan(&entries).Error; err != nil {
		return nil, fmt.Errorf("attempt repository: get leaderboard: %w", err)
	}
	return entries, nil
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

func (r *attemptRepository) GetInstitutionQuizStatsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time) (int, int, error) {
	var totalAttempts int64
	var uniqueStudents int64

	db := r.getDB(ctx).Model(&domain.QuizAttempt{}).Where("institution_id = ? AND submitted_at >= ? AND submitted_at < ?", institutionID, startTime, endTime)

	if err := db.Count(&totalAttempts).Error; err != nil {
		return 0, 0, fmt.Errorf("attempt repository: count total attempts: %w", err)
	}

	if err := db.Distinct("student_id").Count(&uniqueStudents).Error; err != nil {
		return 0, 0, fmt.Errorf("attempt repository: count unique students: %w", err)
	}

	return int(totalAttempts), int(uniqueStudents), nil
}

func (r *attemptRepository) GetTopStudentsByWindow(ctx context.Context, institutionID uuid.UUID, startTime, endTime time.Time, limit int) ([]domain.TopStudentEntry, error) {
	var entries []domain.TopStudentEntry

	query := `
		SELECT 
			CAST(a.student_id AS VARCHAR) AS student_id,
			u.name AS name,
			AVG(a.percentage) AS avg_score
		FROM quiz_attempts a
		JOIN users u ON u.id = a.student_id
		WHERE a.institution_id = ? AND a.submitted_at >= ? AND a.submitted_at < ?
		GROUP BY a.student_id, u.name
		ORDER BY avg_score DESC
		LIMIT ?
	`

	if err := r.getDB(ctx).Raw(query, institutionID, startTime, endTime, limit).Scan(&entries).Error; err != nil {
		return nil, fmt.Errorf("attempt repository: get top students: %w", err)
	}

	return entries, nil
}

