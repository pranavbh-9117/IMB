package workerpool

import "context"

// Job represents an executable unit of work that can be submitted to the pool.
type Job interface {
	Execute(ctx context.Context) error
}

// JobFunc is an adapter to allow the use of ordinary functions as Jobs.
type JobFunc func(ctx context.Context) error

// Execute calls f(ctx).
func (f JobFunc) Execute(ctx context.Context) error {
	return f(ctx)
}
