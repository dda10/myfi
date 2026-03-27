package infra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"myfi-backend/internal/model"
)

// RateLimiter manages per-source token-bucket rate limits using golang.org/x/time/rate.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	configs  map[string]model.RateLimitMetrics
	mu       sync.RWMutex
}

// NewRateLimiter creates a rate limiter with default per-source limits.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		configs:  make(map[string]model.RateLimitMetrics),
	}
	rl.SetLimit("VCI", 100, time.Minute)
	rl.SetLimit("KBS", 100, time.Minute)
	rl.SetLimit("CoinGecko", 50, time.Minute)
	rl.SetLimit("Doji", 60, time.Minute)
	return rl
}

// SetLimit configures a token-bucket limiter for source: maxRequests per window.
func (rl *RateLimiter) SetLimit(source string, maxRequests int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	r := rate.Every(window / time.Duration(maxRequests))
	rl.limiters[source] = rate.NewLimiter(r, maxRequests)
	rl.configs[source] = model.RateLimitMetrics{
		Source:         source,
		MaxRequests:    maxRequests,
		WindowDuration: window.String(),
	}
}

// Allow checks if a request is allowed. Blocks until a token is available or
// returns an error if the source has no limiter configured.
func (rl *RateLimiter) Allow(source string) error {
	rl.mu.RLock()
	limiter, exists := rl.limiters[source]
	rl.mu.RUnlock()

	if !exists {
		return nil // no limit configured — allow
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded for %s: %w", source, err)
	}
	return nil
}

// GetMetrics returns current metrics for a source.
func (rl *RateLimiter) GetMetrics(source string) model.RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	cfg, exists := rl.configs[source]
	if !exists {
		return model.RateLimitMetrics{Source: source}
	}
	return cfg
}

// GetAllMetrics returns metrics for all configured sources.
func (rl *RateLimiter) GetAllMetrics() []model.RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	metrics := make([]model.RateLimitMetrics, 0, len(rl.configs))
	for _, m := range rl.configs {
		metrics = append(metrics, m)
	}
	return metrics
}
