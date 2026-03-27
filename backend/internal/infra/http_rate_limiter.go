package infra

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// httpLimiterEntry holds a token-bucket limiter and last-seen time for cleanup.
type httpLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// HTTPRateLimiter provides per-key (user or IP) rate limiting for HTTP requests.
// Requirement 36.9: per-user 100 req/min, per-IP 200 req/min for unauth endpoints.
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

// Allow returns true if the key is within rate limits.
func (rl *HTTPRateLimiter) Allow(key string) bool {
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

// PerUserRateLimitMiddleware limits authenticated users to maxRequestsPerMin req/min.
// Reads user_id from the Gin context (set by JWTMiddleware).
func PerUserRateLimitMiddleware(maxRequestsPerMin int) gin.HandlerFunc {
	rl := NewHTTPRateLimiter(maxRequestsPerMin)
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}
		key := "user:" + formatID(userID)
		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded — max 100 requests per minute",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			return
		}
		c.Next()
	}
}

// PerIPRateLimitMiddleware limits unauthenticated requests by client IP.
func PerIPRateLimitMiddleware(maxRequestsPerMin int) gin.HandlerFunc {
	rl := NewHTTPRateLimiter(maxRequestsPerMin)
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded — max 200 requests per minute",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			return
		}
		c.Next()
	}
}

// HTTPSRedirectMiddleware redirects HTTP to HTTPS in production.
// Checks the X-Forwarded-Proto header set by load balancers/proxies.
func HTTPSRedirectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Forwarded-Proto") == "http" {
			target := "https://" + c.Request.Host + c.Request.RequestURI
			c.Redirect(http.StatusMovedPermanently, target)
			c.Abort()
			return
		}
		c.Next()
	}
}

func formatID(v any) string {
	switch id := v.(type) {
	case int:
		return itoa(id)
	case int64:
		return itoa(int(id))
	default:
		return "unknown"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
