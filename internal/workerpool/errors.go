// Package workerpool provides a reusable, concurrent worker pool infrastructure.
package workerpool

import "errors"

// ErrPoolClosed is returned when jobs are submitted to a pool that has initiated shutdown.
var ErrPoolClosed = errors.New("workerpool: pool closed")
