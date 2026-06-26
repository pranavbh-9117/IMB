package workerpool

import (
	"context"
	"runtime"
)

// Logger defines the structured logging interface required by the worker pool.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// Options configures the worker pool parameters.
type Options struct {
	WorkersCount int    // Number of concurrent worker goroutines
	QueueSize    int    // Buffer capacity of the job dispatch queue
	Logger       Logger // Injected structured logger
	ErrorHandler func(ctx context.Context, job Job, err error) // Optional error hook
}

// validate sanitizes configuration parameters and applies safe defaults.
func (o *Options) validate() {
	if o.WorkersCount <= 0 {
		o.WorkersCount = runtime.NumCPU()
	}
	if o.QueueSize <= 0 {
		o.QueueSize = 100
	}
	if o.Logger == nil {
		o.Logger = noopLogger{}
	}
}

type noopLogger struct{}

func (noopLogger) Info(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, string, ...any) {}
