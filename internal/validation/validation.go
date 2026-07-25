// Package validation provides input validation and sanitization utilities for the runtime.
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Common regular expressions for validation.
var (
	reEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	reURL   = regexp.MustCompile(`^(https?://)?([\w-]+\.)+[\w-]+(/[\w-./?%&=]*)?$`)
	reID    = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the string representation of the validation error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %q: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error returns the string representation of all validation errors.
func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add adds a validation error to the collection.
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// NotEmpty validates that a string is not empty.
func NotEmpty(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "must not be empty"}
	}
	return nil
}

// MinLength validates that a string meets a minimum length.
func MinLength(value string, min int, field string) error {
	if len(value) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters long", min),
		}
	}
	return nil
}

// MaxLength validates that a string does not exceed a maximum length.
func MaxLength(value string, max int, field string) error {
	if len(value) > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d characters", max),
		}
	}
	return nil
}

// InRange validates that an integer is within a range.
func InRange(value, min, max int, field string) error {
	if value < min || value > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d", min, max),
		}
	}
	return nil
}

// IsEmail validates that a string is a valid email address.
func IsEmail(value, field string) error {
	if !reEmail.MatchString(value) {
		return &ValidationError{Field: field, Message: "must be a valid email address"}
	}
	return nil
}

// IsURL validates that a string is a valid URL.
func IsURL(value, field string) error {
	if !reURL.MatchString(value) {
		return &ValidationError{Field: field, Message: "must be a valid URL"}
	}
	return nil
}

// IsValidID validates that a string is a valid identifier.
func IsValidID(value, field string) error {
	if err := NotEmpty(value, field); err != nil {
		return err
	}
	if !reID.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must contain only alphanumeric characters, underscores, and hyphens",
		}
	}
	return nil
}

// NotNil validates that a value is not nil.
func NotNil(value interface{}, field string) error {
	if value == nil {
		return &ValidationError{Field: field, Message: "must not be nil"}
	}
	return nil
}
