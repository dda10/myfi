package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

// Server wraps an HTTP server with graceful shutdown support.
type Server struct {
	cfg    Config
	engine *gin.Engine
	srv    *http.Server
}

// NewServer creates a Server that will listen on the configured port.
func NewServer(cfg Config, engine *gin.Engine) *Server {
	return &Server{
		cfg:    cfg,
		engine: engine,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: engine,
		},
	}
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
// It performs graceful shutdown, allowing in-flight requests to complete
// within the configured timeout.
func (s *Server) Run() error {
	// Channel to capture startup errors
	errCh := make(chan error, 1)

	go func() {
		slog.Info("starting server", "port", s.cfg.Port, "env", s.cfg.Env)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Wait for interrupt signal or startup error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	slog.Info("shutting down server", "timeout", s.cfg.ShutdownTimeout)
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
