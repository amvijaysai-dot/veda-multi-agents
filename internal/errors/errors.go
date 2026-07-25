// Package errors provides custom error types and error handling utilities for the runtime.
package errors

import (
	"fmt"
)

// ErrorCode represents a machine-readable error identifier.
type ErrorCode string

const (
	// General errors
	ErrCodeInternal    ErrorCode = "INTERNAL"
	ErrCodeValidation  ErrorCode = "VALIDATION"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeConflict    ErrorCode = "CONFLICT"
	ErrCodeTimeout     ErrorCode = "TIMEOUT"
	ErrCodeUnavailable ErrorCode = "UNAVAILABLE"

	// System errors
	ErrCodeInitFailed        ErrorCode = "INIT_FAILED"
	ErrCodeShutdownFailed    ErrorCode = "SHUTDOWN_FAILED"
	ErrCodeStateError        ErrorCode = "STATE_ERROR"
	ErrCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"

	// Dependency errors
	ErrCodeDependencyUnavailable ErrorCode = "DEPENDENCY_UNAVAILABLE"
	ErrCodeDependencyTimeout     ErrorCode = "DEPENDENCY_TIMEOUT"

	// Security errors
	ErrCodeAuthentication   ErrorCode = "AUTHENTICATION"
	ErrCodeAuthorization    ErrorCode = "AUTHORIZATION"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
)

// RuntimeError is the base error type for the runtime.
type RuntimeError struct {
	code    ErrorCode
	Message string
	Cause   error
}

// New creates a new RuntimeError with the given code and message.
func New(code ErrorCode, message string) *RuntimeError {
	return &RuntimeError{
		code:    code,
		Message: message,
	}
}

// Wrap creates a new RuntimeError wrapping an existing error.
func Wrap(code ErrorCode, message string, cause error) *RuntimeError {
	return &RuntimeError{
		code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error returns the string representation of the error.
func (e *RuntimeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.Message)
}

// Unwrap returns the underlying cause of this error.
func (e *RuntimeError) Unwrap() error {
	return e.Cause
}

// Code returns the error code.
func (e *RuntimeError) Code() ErrorCode {
	return e.code
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	return hasCode(err, ErrCodeNotFound)
}

// IsValidation returns true if the error indicates invalid input.
func IsValidation(err error) bool {
	return hasCode(err, ErrCodeValidation)
}

// IsConflict returns true if the error indicates a conflict.
func IsConflict(err error) bool {
	return hasCode(err, ErrCodeConflict)
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	return hasCode(err, ErrCodeTimeout)
}

// IsUnavailable returns true if the error indicates a resource is unavailable.
func IsUnavailable(err error) bool {
	return hasCode(err, ErrCodeUnavailable)
}

// IsInternal returns true if the error indicates an internal system error.
func IsInternal(err error) bool {
	return hasCode(err, ErrCodeInternal)
}

// hasCode checks whether any error in the chain has the given code.
func hasCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if rerr, ok := err.(*RuntimeError); ok {
		if rerr.code == code {
			return true
		}
		if rerr.Cause != nil {
			return hasCode(rerr.Cause, code)
		}
	}
	return false
}
