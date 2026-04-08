package infra

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		if !rl.HTTPAllow(key) {
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
		if !rl.HTTPAllow(key) {
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
