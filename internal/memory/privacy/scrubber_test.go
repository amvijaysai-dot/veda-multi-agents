// Package privacy provides PII scrubbing and data masking capabilities.
package privacy

import (
	"context"
	"regexp"
	"testing"
)

func TestRegexScrubber_DefaultPatterns(t *testing.T) {
	scrubber := NewRegexScrubber(nil, "")
	ctx := context.Background()

	cases := []struct {
		input    string
		expected string
	}{
		{"Contact me at john@example.com for info.", "Contact me at [REDACTED] for info."},
		{"My email is test.user123@domain.co.uk.", "My email is [REDACTED]."},
		{"Call me: 555-123-4567.", "Call me: [REDACTED]."},
		{"Or (555) 123-4567.", "Or [REDACTED]."},
		{"SSN: 123-45-6789 is private.", "SSN: [REDACTED] is private."},
		{"No PII here, just normal text.", "No PII here, just normal text."},
	}

	for _, tc := range cases {
		result, err := scrubber.Scrub(ctx, tc.input)
		if err != nil {
			t.Errorf("unexpected error for %q: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, result)
		}
	}
}

func TestRegexScrubber_CustomPatternsAndMarker(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`secret_code_\d+`),
	}
	scrubber := NewRegexScrubber(patterns, "***")
	ctx := context.Background()

	input := "My code is secret_code_12345."
	expected := "My code is ***."

	result, err := scrubber.Scrub(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRegexScrubber_ContextCancellation(t *testing.T) {
	scrubber := NewRegexScrubber(nil, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := scrubber.Scrub(ctx, "text")
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}
