package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New(ErrCodeNotFound, "resource not found")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if err.Code() != ErrCodeNotFound {
		t.Errorf("expected code %q, got %q", ErrCodeNotFound, err.Code())
	}
	if err.Message != "resource not found" {
		t.Errorf("expected message 'resource not found', got %q", err.Message)
	}
	if err.Cause != nil {
		t.Error("expected nil cause")
	}
}

func TestErrorString(t *testing.T) {
	err := New(ErrCodeValidation, "invalid input")
	expected := "[VALIDATION] invalid input"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("underlying error")
	err := Wrap(ErrCodeInternal, "operation failed", cause)

	if err.Cause != cause {
		t.Error("expected cause to match")
	}

	expected := "[INTERNAL] operation failed: underlying error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := Wrap(ErrCodeInternal, "wrapped", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("expected unwrapped error to match cause")
	}
}

func TestCode(t *testing.T) {
	err := New(ErrCodeTimeout, "timed out")
	if err.Code() != ErrCodeTimeout {
		t.Errorf("expected code %q, got %q", ErrCodeTimeout, err.Code())
	}
}

func TestIsNotFound(t *testing.T) {
	err := New(ErrCodeNotFound, "not found")
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}

	otherErr := New(ErrCodeInternal, "internal error")
	if IsNotFound(otherErr) {
		t.Error("expected IsNotFound to return false for internal error")
	}

	if IsNotFound(nil) {
		t.Error("expected IsNotFound to return false for nil")
	}
}

func TestIsValidation(t *testing.T) {
	err := New(ErrCodeValidation, "invalid")
	if !IsValidation(err) {
		t.Error("expected IsValidation to return true")
	}

	otherErr := New(ErrCodeNotFound, "not found")
	if IsValidation(otherErr) {
		t.Error("expected IsValidation to return false for not found")
	}
}

func TestIsConflict(t *testing.T) {
	err := New(ErrCodeConflict, "conflict")
	if !IsConflict(err) {
		t.Error("expected IsConflict to return true")
	}
}

func TestIsTimeout(t *testing.T) {
	err := New(ErrCodeTimeout, "timeout")
	if !IsTimeout(err) {
		t.Error("expected IsTimeout to return true")
	}
}

func TestIsUnavailable(t *testing.T) {
	err := New(ErrCodeUnavailable, "unavailable")
	if !IsUnavailable(err) {
		t.Error("expected IsUnavailable to return true")
	}
}

func TestIsInternal(t *testing.T) {
	err := New(ErrCodeInternal, "internal")
	if !IsInternal(err) {
		t.Error("expected IsInternal to return true")
	}
}

func TestWrappedErrorDetection(t *testing.T) {
	cause := New(ErrCodeNotFound, "original not found")
	wrapped := Wrap(ErrCodeInternal, "wrapped", cause)

	if !IsNotFound(wrapped) {
		t.Error("expected IsNotFound to detect wrapped error")
	}
}

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		code ErrorCode
		name string
	}{
		{ErrCodeInternal, "INTERNAL"},
		{ErrCodeValidation, "VALIDATION"},
		{ErrCodeNotFound, "NOT_FOUND"},
		{ErrCodeConflict, "CONFLICT"},
		{ErrCodeTimeout, "TIMEOUT"},
		{ErrCodeUnavailable, "UNAVAILABLE"},
		{ErrCodeInitFailed, "INIT_FAILED"},
		{ErrCodeShutdownFailed, "SHUTDOWN_FAILED"},
		{ErrCodeStateError, "STATE_ERROR"},
		{ErrCodeResourceExhausted, "RESOURCE_EXHAUSTED"},
		{ErrCodeDependencyUnavailable, "DEPENDENCY_UNAVAILABLE"},
		{ErrCodeDependencyTimeout, "DEPENDENCY_TIMEOUT"},
		{ErrCodeAuthentication, "AUTHENTICATION"},
		{ErrCodeAuthorization, "AUTHORIZATION"},
		{ErrCodePermissionDenied, "PERMISSION_DENIED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code) != tt.name {
				t.Errorf("expected %q, got %q", tt.name, string(tt.code))
			}
		})
	}
}
