package platform

import (
	"net/http"
	"strings"

	"myfi-backend/internal/infra"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware re-exports the infra panic recovery middleware.
func RecoveryMiddleware() gin.HandlerFunc {
	return infra.RecoveryMiddleware()
}

// GlobalErrorHandler re-exports the infra global error handler middleware.
func GlobalErrorHandler() gin.HandlerFunc {
	return infra.GlobalErrorHandler()
}

// HTTPSRedirectMiddleware re-exports the infra HTTPS redirect middleware.
func HTTPSRedirectMiddleware() gin.HandlerFunc {
	return infra.HTTPSRedirectMiddleware()
}

// PerIPRateLimitMiddleware re-exports the infra per-IP rate limiter.
func PerIPRateLimitMiddleware(maxReqPerMin int) gin.HandlerFunc {
	return infra.PerIPRateLimitMiddleware(maxReqPerMin)
}

// PerUserRateLimitMiddleware re-exports the infra per-user rate limiter.
func PerUserRateLimitMiddleware(maxReqPerMin int) gin.HandlerFunc {
	return infra.PerUserRateLimitMiddleware(maxReqPerMin)
}

// securityHeadersMiddleware adds security headers to every response:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - Strict-Transport-Security (HSTS) in production
func securityHeadersMiddleware(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		if cfg.IsProduction() {
			// HSTS: max-age 1 year, include subdomains
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// maxRequestBodySize is the maximum allowed request body size (1 MB).
const maxRequestBodySize = 1 << 20

// inputValidationMiddleware rejects oversized request bodies and performs
// basic input sanitization on query parameters.
func inputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Reject oversized bodies for non-GET requests
		if c.Request.Method != http.MethodGet && c.Request.ContentLength > maxRequestBodySize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body too large",
			})
			return
		}

		// Basic query parameter sanitization: reject null bytes
		for key, values := range c.Request.URL.Query() {
			for _, v := range values {
				if strings.ContainsRune(v, 0) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "invalid characters in query parameter: " + key,
					})
					return
				}
			}
		}

		c.Next()
	}
}
