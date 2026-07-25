// Package privacy provides PII scrubbing and data masking capabilities for the
// VEDA Agent Runtime memory subsystem.
package privacy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/veda/agent-runtime/internal/memory/interfaces"
)

// RegexScrubber implements interfaces.PrivacyManager using regular expressions.
// It detects common PII patterns and replaces them with a redaction marker.
type RegexScrubber struct {
	patterns []*regexp.Regexp
	marker   string
}

// DefaultPatterns returns a basic set of regular expressions for common PII.
func DefaultPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		// Basic email pattern
		regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),

		// Basic phone pattern (e.g. 555-123-4567, (555) 123-4567, 1-555-123-4567)
		regexp.MustCompile(`(?:\+?1[-. ]?)?\(?([0-9]{3})\)?[-. ]?([0-9]{3})[-. ]?([0-9]{4})`),

		// Basic SSN pattern (e.g. 123-45-6789)
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
	}
}

// NewRegexScrubber creates a new RegexScrubber with the specified patterns and marker.
// If patterns is nil or empty, DefaultPatterns is used.
// If marker is empty, "[REDACTED]" is used.
func NewRegexScrubber(patterns []*regexp.Regexp, marker string) *RegexScrubber {
	if len(patterns) == 0 {
		patterns = DefaultPatterns()
	}
	if marker == "" {
		marker = "[REDACTED]"
	}
	return &RegexScrubber{
		patterns: patterns,
		marker:   marker,
	}
}

// Scrub inspects the input text and replaces matched PII patterns with the marker.
func (s *RegexScrubber) Scrub(ctx context.Context, text string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("privacy: context cancelled during scrub: %w", err)
	}

	scrubbed := text
	for _, pattern := range s.patterns {
		// Context check inside the loop to fail fast if cancelled during many patterns
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("privacy: context cancelled during scrub: %w", err)
		}
		scrubbed = pattern.ReplaceAllString(scrubbed, s.marker)
	}
	return scrubbed, nil
}

var _ interfaces.PrivacyManager = (*RegexScrubber)(nil)
