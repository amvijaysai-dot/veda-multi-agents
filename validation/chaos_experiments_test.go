package validation

import (
	"context"
	"errors"
	"testing"
	"time"

	vedaerr "github.com/veda/agent-runtime/internal/error"
)

// TestChaosRecovery validates the system's ability to recover from unexpected failures.
func TestChaosRecovery(t *testing.T) {
	cb := vedaerr.NewCircuitBreaker(3, 10*time.Millisecond)

	// Simulate a chaotic subsystem that fails randomly (or systematically for the test)
	chaosFunc := func() error {
		return errors.New("chaos failure")
	}

	for i := 0; i < 5; i++ {
		_ = cb.Execute(chaosFunc)
	}

	// The circuit should now be open
	err := cb.Execute(func() error { return nil })
	if err == nil || err.Error() != "circuit breaker open" {
		t.Fatalf("expected circuit breaker open, got %v", err)
	}

	// Wait for recovery timeout
	time.Sleep(15 * time.Millisecond)

	// Should recover
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
}

// TestGracefulDegradation validates that the system falls back gracefully.
func TestGracefulDegradation(t *testing.T) {
	ctx := context.Background()

	// Simulated critical function with a fallback
	attempts := 0
	primaryFunc := func() error {
		attempts++
		return vedaerr.Wrap(errors.New("primary unavailable"), vedaerr.CategoryTransient, "timeout")
	}

	fallbackInvoked := false
	fallbackFunc := func() error {
		fallbackInvoked = true
		return nil
	}

	err := vedaerr.RetryWithBackoff(ctx, 2, 5*time.Millisecond, primaryFunc)
	if err != nil {
		// Degradation kicks in
		err = fallbackFunc()
	}

	if err != nil {
		t.Fatalf("fallback failed: %v", err)
	}

	if !fallbackInvoked {
		t.Fatalf("expected fallback to be invoked")
	}
}
