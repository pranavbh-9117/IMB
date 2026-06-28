// Package repository implements data access patterns for daily statistics.
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

// StatsRepository defines persistence operations for daily statistics.
type StatsRepository interface {
	UpsertStatistic(ctx context.Context, stat *domain.DailyInstituteStatistic) error
}

type statsRepository struct {
	db *gorm.DB
}

// NewStatsRepository creates a new StatsRepository instance.
func NewStatsRepository(db *gorm.DB) StatsRepository {
	return &statsRepository{db: db}
}

// UpsertStatistic saves or updates daily statistics idempotently based on (institution_id, report_date).
// Transaction participation is encapsulated within the infrastructure session coordinator.
func (r *statsRepository) UpsertStatistic(ctx context.Context, stat *domain.DailyInstituteStatistic) error {
	session := database.GetSession(ctx, r.db)

	stat.UpdatedAt = time.Now()
	if stat.CreatedAt.IsZero() {
		stat.CreatedAt = stat.UpdatedAt
	}

	err := session.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "institution_id"}, {Name: "report_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"total_quiz_attempts",
			"unique_students_tested",
			"leaves_approved",
			"leaves_rejected",
			"leaves_pending",
			"top_students",
			"faculty_leave_stats",
			"updated_at",
		}),
	}).Create(stat).Error

	if err != nil {
		return fmt.Errorf("stats repository: upsert: %w", err)
	}
	return nil
}
