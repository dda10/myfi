// Package infra provides infrastructure adapters.
//
// logging.go implements centralized JSON structured logging for all services.
// All log output goes to stdout/stderr in JSON format for easy ingestion
// by log aggregation systems (CloudWatch, ELK, Loki, etc.).
// Requirements: 50.9
package infra

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// InitLogger configures the global slog logger with JSON output to stdout.
// Call this once at startup before any logging occurs.
// The level is configurable via the LOG_LEVEL env var (debug, info, warn, error).
func InitLogger() {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	})
	slog.SetDefault(slog.New(handler))
}

// JSONLoggerMiddleware replaces Gin's default text logger with structured JSON logging.
// Each request produces a single JSON log line with method, path, status, latency,
// client IP, and user agent.
func JSONLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		}

		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}

		// Include trace ID if present (from OpenTelemetry)
		if traceID := c.GetString("trace_id"); traceID != "" {
			attrs = append(attrs, slog.String("trace_id", traceID))
		}

		// Include user ID if authenticated
		if userID, exists := c.Get("user_id"); exists {
			attrs = append(attrs, slog.Any("user_id", userID))
		}

		// Include error if any
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		args := make([]any, len(attrs))
		for i, a := range attrs {
			args[i] = a
		}

		switch {
		case status >= 500:
			slog.Error("request", args...)
		case status >= 400:
			slog.Warn("request", args...)
		default:
			slog.Info("request", args...)
		}
	}
}
