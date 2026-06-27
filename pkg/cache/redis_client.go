package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/pranavbh-9117/IMB/pkg/config"
)

type redisClient struct {
	rdb *redis.Client
}

// NewRedisClient initializes a Redis cache client and validates connectivity via Ping.
func NewRedisClient(cfg config.RedisConfig) (CacheClient, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.TLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis client: connectivity validation failed: %w", err)
	}

	return &redisClient{rdb: rdb}, nil
}

func (c *redisClient) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis client: get: %w", err)
	}
	return val, nil
}

func (c *redisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis client: set: %w", err)
	}
	return nil
}

func (c *redisClient) Delete(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis client: delete: %w", err)
	}
	return nil
}

func (c *redisClient) Ping(ctx context.Context) error {
	if err := c.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis client: ping: %w", err)
	}
	return nil
}
