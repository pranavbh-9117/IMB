// Package cache provides caching functionality and client abstractions for IMB.
package cache

import (
	"context"
	"errors"
	"time"
)


var ErrCacheMiss = errors.New("cache: key not found")

// CacheClient defines the operations needed by the application.
type CacheClient interface {
	// Get retrieves the value for key. Returns ErrCacheMiss if not found.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores the value with the given TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, key string) error

	// Ping checks connectivity to the cache backend.
	Ping(ctx context.Context) error
}
