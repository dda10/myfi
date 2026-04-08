package infra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitMetrics contains metrics for a source.
// (Moved from model/rate_limit_types.go — rate limiting is an infrastructure concern.)
type RateLimitMetrics struct {
	Source         string `json:"source"`
	CurrentCount   int    `json:"currentCount"`
	MaxRequests    int    `json:"maxRequests"`
	QueueDepth     int    `json:"queueDepth"`
	WindowDuration string `json:"windowDuration"`
}

// RateLimiter manages per-source token-bucket rate limits using golang.org/x/time/rate.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	configs  map[string]RateLimitMetrics
	mu       sync.RWMutex
}

// NewRateLimiter creates a rate limiter with default per-source limits.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		configs:  make(map[string]RateLimitMetrics),
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
	rl.configs[source] = RateLimitMetrics{
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
func (rl *RateLimiter) GetMetrics(source string) RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	cfg, exists := rl.configs[source]
	if !exists {
		return RateLimitMetrics{Source: source}
	}
	return cfg
}

// GetAllMetrics returns metrics for all configured sources.
func (rl *RateLimiter) GetAllMetrics() []RateLimitMetrics {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	metrics := make([]RateLimitMetrics, 0, len(rl.configs))
	for _, m := range rl.configs {
		metrics = append(metrics, m)
	}
	return metrics
}

// ---------------------------------------------------------------------------
// HTTP Rate Limiting Middleware (consolidated from http_rate_limiter.go)
// ---------------------------------------------------------------------------

// httpLimiterEntry holds a token-bucket limiter and last-seen time for cleanup.
type httpLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// HTTPRateLimiter provides per-key (user or IP) rate limiting for HTTP requests.
type HTTPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*httpLimiterEntry
	rps     rate.Limit
	burst   int
	ttl     time.Duration
}

// NewHTTPRateLimiter creates a limiter with maxRequests per minute and a cleanup TTL.
func NewHTTPRateLimiter(maxRequestsPerMin int) *HTTPRateLimiter {
	rl := &HTTPRateLimiter{
		entries: make(map[string]*httpLimiterEntry),
		rps:     rate.Every(time.Minute / time.Duration(maxRequestsPerMin)),
		burst:   maxRequestsPerMin,
		ttl:     5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *HTTPRateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.entries[key]
	if !ok {
		entry = &httpLimiterEntry{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.entries[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// HTTPAllow returns true if the key is within rate limits.
func (rl *HTTPRateLimiter) HTTPAllow(key string) bool {
	return rl.getLimiter(key).Allow()
}

// cleanupLoop removes stale entries every minute to prevent unbounded memory growth.
func (rl *HTTPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.ttl)
		for k, e := range rl.entries {
			if e.lastSeen.Before(cutoff) {
				delete(rl.entries, k)
			}
		}
		rl.mu.Unlock()
	}
}
