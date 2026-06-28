// Package scheduler orchestrates background cron jobs.
package scheduler

import (
	"context"

	"github.com/google/uuid"

	reminderSvc "github.com/pranavbh-9117/IMB/internal/reminder/service"
	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// LeaveReminderJob orchestrates sending pending leave reminders.
type LeaveReminderJob struct {
	reminderSvc reminderSvc.ReminderService
}

// NewLeaveReminderJob creates a new LeaveReminderJob instance.
func NewLeaveReminderJob(reminderSvc reminderSvc.ReminderService) *LeaveReminderJob {
	return &LeaveReminderJob{reminderSvc: reminderSvc}
}

// Run executes the leave reminder notification dispatch job.
func (j *LeaveReminderJob) Run() {
	traceID := "cron-reminder-" + uuid.New().String()[:8]
	ctx := logger.WithTraceID(context.Background(), traceID)
	logger.Info(ctx, "scheduler: starting leave reminder job")

	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "scheduler: job panic recovered", "job", "leave_reminder", "panic", r)
		}
	}()

	if err := j.reminderSvc.DispatchLeaveReminders(ctx); err != nil {
		logger.Error(ctx, "scheduler: job failed", "job", "leave_reminder", "error", err)
		return
	}

	logger.Info(ctx, "scheduler: completed leave reminder job")
}
