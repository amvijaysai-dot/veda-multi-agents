package error

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestResilientError(t *testing.T) {
	baseErr := errors.New("base error")
	err := Wrap(baseErr, CategoryTransient, "transient issue")

	if !IsTransient(err) {
		t.Fatal("expected error to be transient")
	}

	if !errors.Is(err, baseErr) {
		t.Fatal("expected unwrapped error to be base error")
	}

	terminalErr := Wrap(baseErr, CategoryTerminal, "terminal issue")
	if IsTransient(terminalErr) {
		t.Fatal("expected error to not be transient")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	errFunc := func() error {
		return errors.New("fail")
	}
	successFunc := func() error {
		return nil
	}

	// 1st failure
	err := cb.Execute(errFunc)
	if err == nil {
		t.Fatal("expected error")
	}

	// 2nd failure - circuit breaks
	err = cb.Execute(errFunc)
	if err == nil {
		t.Fatal("expected error")
	}

	// 3rd attempt - circuit is open
	err = cb.Execute(successFunc)
	if err == nil || err.Error() != "circuit breaker open" {
		t.Fatalf("expected circuit breaker open, got %v", err)
	}

	// wait for reset
	time.Sleep(60 * time.Millisecond)

	// half-open success
	err = cb.Execute(successFunc)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// circuit closed
	err = cb.Execute(successFunc)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return Wrap(errors.New("fail"), CategoryTransient, "retry")
		}
		return nil
	}

	err := RetryWithBackoff(context.Background(), 5, 10*time.Millisecond, fn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	// Terminal error should not retry
	attempts = 0
	fnTerminal := func() error {
		attempts++
		return Wrap(errors.New("fail"), CategoryTerminal, "no retry")
	}

	err = RetryWithBackoff(context.Background(), 5, 10*time.Millisecond, fnTerminal)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
