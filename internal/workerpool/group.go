package workerpool

import (
	"context"
	"sync"
)

// Group coordinates a collection of related jobs submitted to the pool.
type Group interface {
	// Submit dispatches a job to the pool as part of the group. Returns ErrPoolClosed if pool
	// is shut down, or context error if group/submission context expires while enqueuing.
	Submit(job Job) error

	// Wait blocks until all jobs submitted to the group finish execution.
	// Error semantics: Wait for ALL submitted jobs to finish, and return the FIRST encountered error.
	Wait() error
}

type taskGroup struct {
	pool   *workerPool
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

// NewGroup creates a new Group abstraction tied to the worker pool.
func (p *workerPool) NewGroup(ctx context.Context) Group {
	groupCtx, cancel := context.WithCancel(ctx)
	return &taskGroup{
		pool:   p,
		ctx:    groupCtx,
		cancel: cancel,
	}
}

type groupJob struct {
	job Job
	g   *taskGroup
}

func (gj *groupJob) Execute(ctx context.Context) error {
	defer gj.g.wg.Done()
	if err := gj.g.ctx.Err(); err != nil {
		return err
	}
	err := gj.job.Execute(ctx)
	if err != nil {
		gj.g.errOnce.Do(func() {
			gj.g.err = err
			gj.g.cancel()
		})
	}
	return err
}

func (gj *groupJob) OnCancel() {
	gj.g.wg.Done()
}

// Submit enqueues the job into the pool. If any prior job in the group returned an error or the context
// expired, Submit aborts early.
func (g *taskGroup) Submit(job Job) error {
	if err := g.ctx.Err(); err != nil {
		return err
	}

	g.wg.Add(1)

	gj := &groupJob{job: job, g: g}
	err := g.pool.Submit(g.ctx, gj)
	if err != nil {
		g.wg.Done()
		return err
	}
	return nil
}

// Wait blocks until all jobs submitted to the group complete execution, returning the first error.
func (g *taskGroup) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}
