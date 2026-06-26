package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
)

// Pool defines the contract for submitting jobs and managing worker concurrency.
type Pool interface {
	Submit(ctx context.Context, job Job) error
	NewGroup(ctx context.Context) Group
	Metrics() Metrics
	Shutdown(ctx context.Context) error
}

// jobItem encapsulates a submitted job and its request-scoped context.
type jobItem struct {
	ctx context.Context
	job Job
}

type workerPool struct {
	opts     Options
	jobQueue chan jobItem
	wg       sync.WaitGroup

	closed        atomic.Bool
	activeWorkers atomic.Int64
	completedJobs atomic.Int64
	failedJobs    atomic.Int64
	shutdownOnce  sync.Once
}

// New initializes the worker pool and immediately launches worker goroutines.
func New(opts Options) (Pool, error) {
	opts.validate()
	p := &workerPool{
		opts:     opts,
		jobQueue: make(chan jobItem, opts.QueueSize),
	}

	p.opts.Logger.Info(context.Background(), "starting worker pool", "workers_count", opts.WorkersCount, "queue_size", opts.QueueSize)

	p.wg.Add(opts.WorkersCount)
	for i := 0; i < opts.WorkersCount; i++ {
		go p.workerLoop(i + 1)
	}

	return p, nil
}

// Submit dispatches a job to the pool. It blocks until the job is accepted into the queue buffer
// or the supplied context is canceled. Returns ErrPoolClosed if shutdown has been initiated.
func (p *workerPool) Submit(ctx context.Context, job Job) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	item := jobItem{ctx: ctx, job: job}

	select {
	case p.jobQueue <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Metrics returns a real-time snapshot of the worker pool's runtime statistics.
func (p *workerPool) Metrics() Metrics {
	return Metrics{
		WorkersCount:  p.opts.WorkersCount,
		ActiveWorkers: p.activeWorkers.Load(),
		QueuedJobs:    len(p.jobQueue),
		CompletedJobs: p.completedJobs.Load(),
		FailedJobs:    p.failedJobs.Load(),
	}
}

// Shutdown initiates a graceful stop of the pool. It rejects new submissions, closes the job queue,
// and blocks until all running and queued jobs finish execution or ctx expires.
func (p *workerPool) Shutdown(ctx context.Context) error {
	p.closed.Store(true)

	p.shutdownOnce.Do(func() {
		p.opts.Logger.Info(ctx, "initiating worker pool shutdown")
		close(p.jobQueue)
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.opts.Logger.Info(ctx, "worker pool shut down successfully")
		return nil
	case <-ctx.Done():
		p.opts.Logger.Error(ctx, "worker pool shutdown timed out or canceled", "error", ctx.Err())
		return ctx.Err()
	}
}
