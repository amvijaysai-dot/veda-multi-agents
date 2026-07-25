package validation

import (
	"testing"
)

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		value     string
		field     string
		expectErr bool
	}{
		{"hello", "name", false},
		{"", "name", true},
		{"   ", "name", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := NotEmpty(tt.value, tt.field)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		value     string
		min       int
		field     string
		expectErr bool
	}{
		{"hello", 3, "name", false},
		{"hi", 3, "name", true},
		{"", 1, "name", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := MinLength(tt.value, tt.min, tt.field)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		value     string
		max       int
		field     string
		expectErr bool
	}{
		{"hello", 10, "name", false},
		{"hello world", 5, "name", true},
		{"", 10, "name", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := MaxLength(tt.value, tt.max, tt.field)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestInRange(t *testing.T) {
	tests := []struct {
		value     int
		min       int
		max       int
		field     string
		expectErr bool
	}{
		{5, 1, 10, "count", false},
		{0, 1, 10, "count", true},
		{11, 1, 10, "count", true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := InRange(tt.value, tt.min, tt.max, tt.field)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		value     string
		expectErr bool
	}{
		{"user@example.com", false},
		{"user.name+tag@example.co.uk", false},
		{"not-an-email", true},
		{"@example.com", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := IsEmail(tt.value, "email")
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		value     string
		expectErr bool
	}{
		{"https://example.com", false},
		{"http://example.com/path", false},
		{"example.com", false},
		{"not a url", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := IsURL(tt.value, "url")
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		value     string
		expectErr bool
	}{
		{"valid-id_123", false},
		{"agent-001", false},
		{"", true},
		{"spaces not allowed", true},
		{"special@chars", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := IsValidID(tt.value, "id")
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestNotNil(t *testing.T) {
	// Note: In Go, a typed nil pointer (e.g., *string) is NOT nil when cast to interface{}
	// So we test with an explicit interface{} nil value
	var nilVar interface{} = nil
	notNil := "hello"

	if err := NotNil(nilVar, "var"); err == nil {
		t.Error("expected error for nil value")
	}

	if err := NotNil(&notNil, "ptr"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "must not be empty"}
	expected := `validation failed for field "name": must not be empty`
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationErrors(t *testing.T) {
	var errs ValidationErrors
	if errs.HasErrors() {
		t.Error("expected HasErrors to return false for empty errors")
	}

	errs.Add("name", "must not be empty")
	errs.Add("age", "must be positive")

	if !errs.HasErrors() {
		t.Error("expected HasErrors to return true")
	}

	msg := errs.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}
