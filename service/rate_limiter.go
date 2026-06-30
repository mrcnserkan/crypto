package service

import (
	"sync"
	"time"
)

const (
	defaultRateLimitPerMinute = 10
	proRateLimitPerMinute     = 30
)

// RateLimiter provides a simple token-bucket rate limiter for API calls.
type RateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = defaultRateLimitPerMinute
	}
	return &RateLimiter{
		interval: time.Minute / time.Duration(requestsPerMinute),
		last:     time.Time{},
	}
}

func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.last.IsZero() {
		rl.last = time.Now()
		return
	}
	elapsed := time.Since(rl.last)
	if elapsed < rl.interval {
		time.Sleep(rl.interval - elapsed)
	}
	rl.last = time.Now()
}

var globalRateLimiter = NewRateLimiter(defaultRateLimitPerMinute)

func setGlobalRateLimiter(requestsPerMinute int) {
	globalRateLimiter = NewRateLimiter(requestsPerMinute)
}

func waitForRateLimit() {
	globalRateLimiter.Wait()
}
