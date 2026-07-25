// Package error provides resilient error handling mechanisms and recovery patterns.
package error

import (
	"errors"
	"fmt"
)

// ErrorCategory categorizes errors to determine recovery strategies.
type ErrorCategory int

const (
	// CategoryTransient implies the error might succeed if retried.
	CategoryTransient ErrorCategory = iota
	// CategoryTerminal implies the error is unrecoverable.
	CategoryTerminal
	// CategoryResource limits (e.g., rate limits, out of memory).
	CategoryResource
)

// ResilientError wraps an underlying error with a category.
type ResilientError struct {
	Category ErrorCategory
	Err      error
	Message  string
}

// Error implements the error interface.
func (e *ResilientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Category, e.Message)
}

// Unwrap returns the underlying error.
func (e *ResilientError) Unwrap() error {
	return e.Err
}

// Wrap creates a ResilientError.
func Wrap(err error, category ErrorCategory, message string) *ResilientError {
	return &ResilientError{
		Category: category,
		Err:      err,
		Message:  message,
	}
}

// IsTransient returns true if the error or its wrapped errors are transient.
func IsTransient(err error) bool {
	var re *ResilientError
	if errors.As(err, &re) {
		return re.Category == CategoryTransient
	}
	return false
}
