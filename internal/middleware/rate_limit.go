// Package middleware provides Gin middlewares for security, context, and observability.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/pkg/logger"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// KeyExtractor defines a reusable extractor function abstraction for obtaining rate limit keys.
type KeyExtractor func(*gin.Context) (string, error)

// ByIP extracts the client IP address from the request.
func ByIP() KeyExtractor {
	return func(c *gin.Context) (string, error) {
		ip := c.ClientIP()
		if ip == "" {
			return "", errors.New("client IP is empty")
		}
		return ip, nil
	}
}

// ByBodyField extracts a specific JSON field value from the request body.
func ByBodyField(field string) KeyExtractor {
	return func(c *gin.Context) (string, error) {
		if c.Request == nil || c.Request.Body == nil {
			return "", errors.New("request body is nil")
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return "", err
		}

		// Restore request body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var data map[string]any
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			return "", err
		}

		val, exists := data[field]
		if !exists || val == nil {
			return "", errors.New("field not found in request body")
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" {
			return "", errors.New("field value is not a non-empty string")
		}

		return strVal, nil
	}
}

// entry holds rate limit timestamps for a specific key.
type entry struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// RateLimitStore manages sliding window rate limit state in memory.
type RateLimitStore interface {
	Allow(key string, limit int, now time.Time) (allowed bool, remaining int, reset time.Time)
}

type rateLimitStore struct {
	mu      sync.RWMutex
	entries map[string]*entry
	window  time.Duration
}

// NewRateLimitStore initializes an in-memory rate limit store with a context lifecycle and 5-minute cleanup interval.
func NewRateLimitStore(ctx context.Context, window time.Duration) RateLimitStore {
	store := &rateLimitStore{
		entries: make(map[string]*entry),
		window:  window,
	}

	go store.cleanup(ctx)
	return store
}

// cleanup runs periodically every 5 minutes to evict expired entries.
func (s *rateLimitStore) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			cutoff := now.Add(-s.window)
			s.mu.Lock()
			for k, e := range s.entries {
				e.mu.Lock()
				idx := 0
				for idx < len(e.timestamps) && e.timestamps[idx].Before(cutoff) {
					idx++
				}
				if idx > 0 {
					e.timestamps = append([]time.Time(nil), e.timestamps[idx:]...)
				}
				empty := len(e.timestamps) == 0
				e.mu.Unlock()

				if empty {
					delete(s.entries, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Allow evaluates quota against the sliding window.
func (s *rateLimitStore) Allow(key string, limit int, now time.Time) (bool, int, time.Time) {
	cutoff := now.Add(-s.window)

	s.mu.RLock()
	e, exists := s.entries[key]
	s.mu.RUnlock()

	if !exists {
		s.mu.Lock()
		e, exists = s.entries[key]
		if !exists {
			e = &entry{}
			s.entries[key] = e
		}
		s.mu.Unlock()
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	idx := 0
	for idx < len(e.timestamps) && !e.timestamps[idx].After(cutoff) {
		idx++
	}
	if idx > 0 {
		e.timestamps = append([]time.Time(nil), e.timestamps[idx:]...)
	}

	reset := now.Add(s.window)
	if len(e.timestamps) > 0 {
		reset = e.timestamps[0].Add(s.window)
	}

	if len(e.timestamps) < limit {
		e.timestamps = append(e.timestamps, now)
		remaining := limit - len(e.timestamps)
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, reset
	}

	return false, 0, reset
}

// RateLimitConfig configures the rate limiting middleware.
type RateLimitConfig struct {
	Store     RateLimitStore
	Limit     int
	Window    time.Duration
	Extractor KeyExtractor
}

// RateLimit creates a sliding window rate limiting Gin middleware.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := cfg.Extractor(c)
		if err != nil {
			logger.Warn(c.Request.Context(), "rate limit key extraction failed, failing open", "error", err.Error())
			c.Next()
			return
		}

		now := time.Now()
		allowed, remaining, reset := cfg.Store.Allow(key, cfg.Limit, now)

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		if !allowed {
			retryAfter := int(math.Ceil(reset.Sub(now).Seconds()))
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))

			endpoint := c.FullPath()
			if endpoint == "" {
				endpoint = c.Request.URL.Path
			}

			logger.Warn(c.Request.Context(), "Rate limit exceeded",
				"endpoint", endpoint,
				"extracted_key", key,
				"configured_limit", cfg.Limit,
				"window_duration", cfg.Window.String(),
			)

			response.TooManyRequests(c, "too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
