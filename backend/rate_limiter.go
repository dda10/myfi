package main

import (
	"fmt"
	"sync"
	"time"
)

// RateLimit tracks rate limiting state for a single source
type RateLimit struct {
	MaxRequests int
	Window      time.Duration
	Queue       chan struct{}
	Counter     int
	WindowStart time.Time
	mu          sync.Mutex
}

// RateLimiter manages per-source rate limits
type RateLimiter struct {
	limits map[string]*RateLimit
	mu     sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with default limits
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits: make(map[string]*RateLimit),
	}

	// Configure per-source limits based on API documentation
	// VCI: Conservative limit to avoid blocking
	rl.SetLimit("VCI", 100, time.Minute)

	// KBS: Will be configured when available
	rl.SetLimit("KBS", 100, time.Minute)

	// CoinGecko: Free tier limit
	rl.SetLimit("CoinGecko", 50, time.Minute)

	// Doji Gold API: Conservative limit
	rl.SetLimit("Doji", 60, time.Minute)

	return rl
}

// SetLimit configures rate limit for a source
func (rl *RateLimiter) SetLimit(source string, maxRequests int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limits[source] = &RateLimit{
		MaxRequests: maxRequests,
		Window:      window,
		Queue:       make(chan struct{}, 100), // Max queue depth of 100
		Counter:     0,
		WindowStart: time.Now(),
	}
}

// Allow checks if a request is allowed and blocks if rate limit is reached
func (rl *RateLimiter) Allow(source string) error {
	rl.mu.RLock()
	limit, exists := rl.limits[source]
	rl.mu.RUnlock()

	if !exists {
		// No limit configured for this source, allow by default
		return nil
	}

	limit.mu.Lock()
	defer limit.mu.Unlock()

	now := time.Now()

	// Reset counter if window has passed
	if now.Sub(limit.WindowStart) >= limit.Window {
		limit.Counter = 0
		limit.WindowStart = now
	}

	// Check if under limit
	if limit.Counter < limit.MaxRequests {
		limit.Counter++
		return nil
	}

	// Rate limit exceeded, try to queue
	select {
	case limit.Queue <- struct{}{}:
		// Successfully queued, wait for window to reset
		limit.mu.Unlock()

		// Wait for next window
		waitTime := limit.Window - now.Sub(limit.WindowStart)
		time.Sleep(waitTime)

		limit.mu.Lock()
		<-limit.Queue // Remove from queue

		// Reset if needed and increment
		now = time.Now()
		if now.Sub(limit.WindowStart) >= limit.Window {
			limit.Counter = 0
			limit.WindowStart = now
		}
		limit.Counter++
		return nil

	default:
		// Queue is full
		return fmt.Errorf("rate limit exceeded for %s: queue full (max 100)", source)
	}
}

// RateLimitMetrics contains metrics for a source
type RateLimitMetrics struct {
	Source         string `json:"source"`
	CurrentCount   int    `json:"currentCount"`
	MaxRequests    int    `json:"maxRequests"`
	QueueDepth     int    `json:"queueDepth"`
	WindowDuration string `json:"windowDuration"`
}

// GetMetrics returns current metrics for a source
func (rl *RateLimiter) GetMetrics(source string) RateLimitMetrics {
	rl.mu.RLock()
	limit, exists := rl.limits[source]
	rl.mu.RUnlock()

	if !exists {
		return RateLimitMetrics{
			Source: source,
		}
	}

	limit.mu.Lock()
	defer limit.mu.Unlock()

	return RateLimitMetrics{
		Source:         source,
		CurrentCount:   limit.Counter,
		MaxRequests:    limit.MaxRequests,
		QueueDepth:     len(limit.Queue),
		WindowDuration: limit.Window.String(),
	}
}

// GetAllMetrics returns metrics for all sources
func (rl *RateLimiter) GetAllMetrics() []RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	metrics := make([]RateLimitMetrics, 0, len(rl.limits))
	for source := range rl.limits {
		metrics = append(metrics, rl.GetMetrics(source))
	}
	return metrics
}
