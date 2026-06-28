// Package scheduler orchestrates background cron jobs.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	instRepo "github.com/pranavbh-9117/IMB/internal/institution/repository"
	statsSvc "github.com/pranavbh-9117/IMB/internal/statistics/service"
	"github.com/pranavbh-9117/IMB/internal/workerpool"
	"github.com/pranavbh-9117/IMB/pkg/clock"
	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// DailyStatsJob orchestrates daily aggregation of statistics for all active institutions.
type DailyStatsJob struct {
	instRepo   instRepo.InstitutionRepository
	statsSvc   statsSvc.StatisticsService
	workerPool workerpool.Pool
	clk        clock.Clock
}

// NewDailyStatsJob creates a new DailyStatsJob instance.
func NewDailyStatsJob(
	instRepo instRepo.InstitutionRepository,
	statsSvc statsSvc.StatisticsService,
	workerPool workerpool.Pool,
	clk clock.Clock,
) *DailyStatsJob {
	return &DailyStatsJob{
		instRepo:   instRepo,
		statsSvc:   statsSvc,
		workerPool: workerPool,
		clk:        clk,
	}
}

// Run executes the statistics generation job across all active institutions using the worker pool.
func (j *DailyStatsJob) Run() {
	traceID := "cron-stats-" + uuid.New().String()[:8]
	ctx := logger.WithTraceID(context.Background(), traceID)
	logger.Info(ctx, "scheduler: starting daily stats job")

	// Report date covers the previous UTC day
	now := j.clk.Now().UTC()
	reportDate := now.Add(-24 * time.Hour)

	institutions, err := j.instRepo.List(ctx, 0, 0)
	if err != nil {
		logger.Error(ctx, "scheduler: job failed", "job", "daily_stats", "error", fmt.Errorf("list institutions: %w", err))
		return
	}

	var wg sync.WaitGroup
	for _, inst := range institutions {
		instID := inst.ID
		wg.Add(1)
		err := j.workerPool.Submit(ctx, workerpool.JobFunc(func(jobCtx context.Context) error {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error(jobCtx, "scheduler: job panic recovered", "job", "daily_stats", "institution_id", instID, "panic", r)
				}
			}()

			execCtx, cancel := context.WithTimeout(jobCtx, 10*time.Minute)
			defer cancel()

			if genErr := j.statsSvc.GenerateInstitutionStatistics(execCtx, instID, reportDate); genErr != nil {
				logger.Error(execCtx, "scheduler: job failed for institution", "job", "daily_stats", "institution_id", instID, "error", genErr)
				return genErr
			}
			return nil
		}))

		if err != nil {
			wg.Done()
			logger.Error(ctx, "scheduler: failed to submit job to worker pool", "job", "daily_stats", "institution_id", instID, "error", err)
		}
	}

	wg.Wait()
	logger.Info(ctx, "scheduler: completed daily stats job")
}
