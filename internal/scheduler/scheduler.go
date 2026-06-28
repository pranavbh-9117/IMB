// Package scheduler orchestrates background cron jobs.
package scheduler

import (
	"context"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// Runner manages the registration and execution lifecycle of scheduled background jobs.
type Runner struct {
	cron           *cron.Cron
	statsJob       *DailyStatsJob
	reminderJob    *LeaveReminderJob
	dailyStatsCron string
	reminderCron   string
}

// New creates a new scheduler Runner.
func New(statsJob *DailyStatsJob, reminderJob *LeaveReminderJob, dailyStatsCron, reminderCron string) *Runner {
	return &Runner{
		cron:           cron.New(),
		statsJob:       statsJob,
		reminderJob:    reminderJob,
		dailyStatsCron: dailyStatsCron,
		reminderCron:   reminderCron,
	}
}

// Start registers configured jobs and begins the background scheduler in a non-blocking goroutine.
func (r *Runner) Start() {
	traceID := "cron-runner-" + uuid.New().String()[:8]
	ctx := logger.WithTraceID(context.Background(), traceID)
	logger.Info(ctx, "scheduler: registering cron jobs")

	_, err := r.cron.AddJob(r.dailyStatsCron, r.statsJob)
	if err != nil {
		logger.Error(ctx, "scheduler: failed to register daily stats job", "error", err, "cron", r.dailyStatsCron)
	}

	_, err = r.cron.AddJob(r.reminderCron, r.reminderJob)
	if err != nil {
		logger.Error(ctx, "scheduler: failed to register reminder job", "error", err, "cron", r.reminderCron)
	}

	r.cron.Start()
	logger.Info(ctx, "scheduler: runner started successfully")
}

// Stop initiates a graceful shutdown of the background scheduler, waiting for active jobs to complete.
func (r *Runner) Stop() {
	traceID := "cron-runner-" + uuid.New().String()[:8]
	ctx := logger.WithTraceID(context.Background(), traceID)
	logger.Info(ctx, "scheduler: stopping runner")
	ctxStop := r.cron.Stop()
	<-ctxStop.Done()
	logger.Info(ctx, "scheduler: runner stopped successfully")
}
