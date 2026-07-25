// Package interfaces defines the contracts for the memory subsystem.
package interfaces

import "context"

// PrivacyManager provides PII scrubbing and data masking capabilities.
type PrivacyManager interface {
	// Scrub inspects the input text for sensitive patterns (PII) and replaces
	// them with redaction markers (e.g. [REDACTED]).
	Scrub(ctx context.Context, text string) (string, error)
}
