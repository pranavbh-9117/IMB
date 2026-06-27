package cache

import (
	"context"
	"time"
)

type noopClient struct{}

// NewNoopClient initializes a no-op CacheClient implementation.
func NewNoopClient() CacheClient {
	return &noopClient{}
}

func (c *noopClient) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (c *noopClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *noopClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *noopClient) Ping(ctx context.Context) error {
	return nil
}
