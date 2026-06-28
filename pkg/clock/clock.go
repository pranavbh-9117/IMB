// Package clock provides a time abstraction for deterministic testing across date boundaries.
package clock

import (
	"sync"
	"time"
)

// Clock abstracts time measurement.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using actual system time.
type RealClock struct{}

// NewRealClock creates a new RealClock instance.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current UTC time.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// FakeClock implements Clock for deterministic unit and integration testing.
type FakeClock struct {
	mu  sync.RWMutex
	now time.Time
}

// NewFakeClock creates a new FakeClock initialized to the given time.
func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{now: t}
}

// Now returns the simulated time.
func (c *FakeClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// Set updates the simulated time to t.
func (c *FakeClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

// Advance moves the simulated time forward by d.
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
