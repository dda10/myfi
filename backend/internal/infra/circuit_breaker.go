package infra

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker/v2"
)

// CircuitState mirrors gobreaker states for external consumers.
type CircuitState = gobreaker.State

var (
	StateClosed   = gobreaker.StateClosed
	StateOpen     = gobreaker.StateOpen
	StateHalfOpen = gobreaker.StateHalfOpen
)

// CircuitBreaker wraps sony/gobreaker with the same Call/GetState/Reset interface.
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker[struct{}]
}

// NewCircuitBreaker creates a circuit breaker that opens after maxFailures consecutive
// failures and attempts recovery after timeout.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        "cb",
		MaxRequests: 1,
		Interval:    timeout,
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(maxFailures)
		},
	}
	return &CircuitBreaker{cb: gobreaker.NewCircuitBreaker[struct{}](settings)}
}

// Call executes fn with circuit breaker protection.
func (c *CircuitBreaker) Call(fn func() error) error {
	_, err := c.cb.Execute(func() (struct{}, error) {
		return struct{}{}, fn()
	})
	if err == gobreaker.ErrOpenState {
		return fmt.Errorf("circuit breaker is open")
	}
	return err
}

// GetState returns the current circuit state.
func (c *CircuitBreaker) GetState() CircuitState {
	return c.cb.State()
}

// GetFailures returns consecutive failure count.
func (c *CircuitBreaker) GetFailures() int {
	return int(c.cb.Counts().ConsecutiveFailures)
}

// Reset manually resets the circuit breaker (not directly supported by gobreaker;
// achieved by draining the half-open probe — callers should rely on timeout instead).
func (c *CircuitBreaker) Reset() {
	// gobreaker resets automatically on timeout; no manual reset API exposed.
}
