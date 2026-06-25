package retry

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/pranavbh-9117/IMB/pkg/logger"
)

// Retry behavior configuration.
type Config struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
	ShouldRetry  func(err error) bool
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.5,
		ShouldRetry:  IsTransientError,
	}
}

// Do executes fn, retrying on transient errors per cfg.
// Returns the last error if all attempts are exhausted.
func Do(ctx context.Context, fn func() error, cfg Config) error {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if cfg.ShouldRetry != nil && !cfg.ShouldRetry(lastErr) {
			return lastErr
		}

		if attempt == cfg.MaxAttempts {
			logger.Error(ctx, "retry: all attempts exhausted", "attempts", cfg.MaxAttempts, "error", lastErr.Error())
			break
		}

		delay := computeDelay(cfg, attempt)

		logger.Warn(ctx, "retry: attempt failed", "attempt", attempt, "max_attempts", cfg.MaxAttempts, "error", lastErr.Error(), "next_delay_ms", delay.Milliseconds())

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():

			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()

		case <-timer.C:
		}
	}

	return lastErr
}

func computeDelay(cfg Config, attempt int) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt-1))
	if cfg.MaxDelay > 0 && delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}
	if cfg.Jitter > 0 {
		delay += rand.Float64() * cfg.Jitter * float64(cfg.InitialDelay)
	}
	return time.Duration(delay)
}
