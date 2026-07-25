package error

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CircuitBreaker provides a mechanism to prevent cascading failures.
type CircuitBreaker struct {
	mu           sync.Mutex
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailure  time.Time
	state        State
}

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Fast failure
	StateHalfOpen              // Testing recovery
)

// NewCircuitBreaker creates a circuit breaker.
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute runs the function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker open")
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Success
	cb.failures = 0
	cb.state = StateClosed
	return nil
}

// RetryWithBackoff executes a function with exponential backoff and jitter.
func RetryWithBackoff(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !IsTransient(err) {
			return err
		}

		delay := time.Duration(1<<i) * baseDelay
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))

		select {
		case <-time.After(delay + jitter):
			// continue loop
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("max retries exceeded: %w", err)
}
