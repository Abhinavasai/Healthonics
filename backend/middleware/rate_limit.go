package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// slidingWindowEntry tracks request timestamps for a single key.
type slidingWindowEntry struct {
	mu        sync.Mutex
	timestamps []time.Time
}

// RateLimiter holds per-key sliding window state.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*slidingWindowEntry
	limit    int
	window   time.Duration
}

func NewRateLimiter(ctx context.Context, limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*slidingWindowEntry),
		limit:   limit,
		window:  window,
	}
	// Periodically clean up stale buckets to prevent unbounded memory growth.
	go rl.cleanup(ctx)
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	e, ok := rl.buckets[key]
	if !ok {
		e = &slidingWindowEntry{}
		rl.buckets[key] = e
	}
	rl.mu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = valid

	if len(e.timestamps) >= rl.limit {
		return false
	}
	e.timestamps = append(e.timestamps, now)
	return true
}

func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-rl.window)
			rl.mu.Lock()
			for key, e := range rl.buckets {
				e.mu.Lock()
				alive := false
				for _, t := range e.timestamps {
					if t.After(cutoff) {
						alive = true
						break
					}
				}
				if !alive {
					delete(rl.buckets, key)
				}
				e.mu.Unlock()
			}
			rl.mu.Unlock()
		}
	}
}

// Middleware returns a Gin handler that rate-limits by client IP.
// limit: max requests, window: rolling time window.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before trying again.",
			})
			return
		}
		c.Next()
	}
}
