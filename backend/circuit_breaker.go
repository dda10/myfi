package main

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half_open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	state        CircuitState
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
		failures:    0,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if circuit should transition from Open to HalfOpen
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) >= cb.timeout {
			cb.state = StateHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}

	currentState := cb.state
	cb.mu.Unlock()

	// Execute the function
	err := fn()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	// Success
	if currentState == StateHalfOpen {
		cb.RecordSuccess()
	}

	return nil
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
	}
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	// Check if we should open the circuit
	// Threshold: 3 failures within 60 seconds
	if cb.failures >= cb.maxFailures {
		if cb.state == StateClosed || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
}
