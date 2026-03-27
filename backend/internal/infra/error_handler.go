package infra

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the structured JSON error returned to clients.
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// GlobalErrorHandler returns Gin middleware that:
//   - Recovers from panics (complementing RecoveryMiddleware)
//   - Catches unhandled errors left on the Gin context
//   - Logs errors with structured slog fields
//   - Returns a consistent JSON error envelope
func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in handler",
					"error", r,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Error:     "internal server error",
					Code:      http.StatusInternalServerError,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			// Log any errors that handlers attached to the context
			if len(c.Errors) > 0 {
				for _, e := range c.Errors {
					slog.Error("request error",
						"error", e.Error(),
						"method", c.Request.Method,
						"path", c.Request.URL.Path,
						"status", c.Writer.Status(),
						"latency_ms", time.Since(start).Milliseconds(),
					)
				}
			}

			// Log slow requests (>2s) as warnings
			latency := time.Since(start)
			if latency > 2*time.Second {
				slog.Warn("slow request",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"latency_ms", latency.Milliseconds(),
					"status", c.Writer.Status(),
				)
			}
		}()

		c.Next()
	}
}
