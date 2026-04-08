// Package platform provides cross-cutting glue that wires
// domain packages together into a running HTTP server.
package platform

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	// Server
	Port            int
	Env             string // "development" or "production"
	FrontendOrigin  string
	ShutdownTimeout time.Duration

	// Rate limiting
	IPRateLimit   int // requests per minute per IP
	UserRateLimit int // requests per minute per authenticated user
}

// LoadConfig reads configuration from environment variables with sensible defaults.
func LoadConfig() Config {
	cfg := Config{
		Port:            envInt("PORT", 8080),
		Env:             envStr("ENV", "development"),
		FrontendOrigin:  envStr("FRONTEND_ORIGIN", "http://localhost:3000"),
		ShutdownTimeout: time.Duration(envInt("SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
		IPRateLimit:     envInt("IP_RATE_LIMIT", 200),
		UserRateLimit:   envInt("USER_RATE_LIMIT", 100),
	}
	return cfg
}

// IsProduction returns true when running in production mode.
func (c Config) IsProduction() bool {
	return c.Env == "production"
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
