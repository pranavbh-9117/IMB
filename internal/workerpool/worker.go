package workerpool

import "context"

// workerLoop runs the continuous dispatch loop for a worker goroutine.
func (p *workerPool) workerLoop(workerID int) {
	defer p.wg.Done()

	for item := range p.jobQueue {
		// Queued-but-not-started cancellation check
		if err := item.ctx.Err(); err != nil {
			p.failedJobs.Add(1)
			p.opts.Logger.Error(item.ctx, "skipping queued job due to canceled context", "worker_id", workerID, "error", err)
			if cj, ok := item.job.(interface{ OnCancel() }); ok {
				cj.OnCancel()
			}
			continue
		}

		p.activeWorkers.Add(1)
		p.opts.Logger.Info(item.ctx, "worker started job execution", "worker_id", workerID)

		p.executeJob(workerID, item)

		p.activeWorkers.Add(-1)
	}

	p.opts.Logger.Info(context.Background(), "worker stopped", "worker_id", workerID)
}

// executeJob runs the submitted job within a panic-recovery block and records completion metrics.
func (p *workerPool) executeJob(workerID int, item jobItem) {
	defer func() {
		if r := recover(); r != nil {
			p.failedJobs.Add(1)
			p.opts.Logger.Error(item.ctx, "worker panicked during job execution", "worker_id", workerID, "panic", r)
		}
	}()

	err := item.job.Execute(item.ctx)
	if err != nil {
		p.failedJobs.Add(1)
		p.opts.Logger.Error(item.ctx, "job execution failed", "worker_id", workerID, "error", err)
		if p.opts.ErrorHandler != nil {
			p.opts.ErrorHandler(item.ctx, item.job, err)
		}
	} else {
		p.completedJobs.Add(1)
		p.opts.Logger.Info(item.ctx, "job completed successfully", "worker_id", workerID)
	}
}
