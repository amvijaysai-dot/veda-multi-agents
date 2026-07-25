// Package spec defines the AgentSpec structure and related validation logic.
package spec

import (
	"fmt"

	"github.com/veda/agent-runtime/internal/validation"
)

const (
	// defaultMaxIterations is used when AgentSpec.MaxIterations is not set (zero).
	defaultMaxIterations = 10

	// maxNameLength is the maximum byte length of AgentSpec.Name.
	maxNameLength = 256

	// maxDescriptionLength is the maximum byte length of AgentSpec.Description.
	maxDescriptionLength = 2048

	// maxCapabilities is the maximum number of capabilities that can be declared.
	maxCapabilities = 100

	// maxConfigEntries is the maximum number of config key-value pairs.
	maxConfigEntries = 256
)

// Validate checks that the AgentSpec is well-formed and returns a descriptive
// error if any required field is missing or any field violates its constraint.
//
// Validate must be called before passing an AgentSpec to the lifecycle creator.
// A spec that passes Validate is considered safe for further processing.
//
// Rules:
//   - ID must be non-empty and match [a-zA-Z0-9_-]+
//   - Name must be non-empty and ≤ maxNameLength bytes
//   - Description may be empty but must be ≤ maxDescriptionLength bytes
//   - ModelID must be non-empty
//   - MaxIterations must be ≥ 0; if 0 it will be normalized to defaultMaxIterations
//   - Capabilities list may be empty but no individual entry may be empty
//   - Capabilities list must have ≤ maxCapabilities entries
//   - Config map may be nil but must have ≤ maxConfigEntries entries
//
// Returns the first validation error encountered, not a combined list, to keep
// error messages focused and actionable.
func Validate(s *AgentSpec) error {
	if s == nil {
		return fmt.Errorf("spec must not be nil")
	}

	if err := validation.IsValidID(s.ID, "ID"); err != nil {
		return fmt.Errorf("invalid AgentSpec: %w", err)
	}

	if err := validation.NotEmpty(s.Name, "Name"); err != nil {
		return fmt.Errorf("invalid AgentSpec: %w", err)
	}
	if err := validation.MaxLength(s.Name, maxNameLength, "Name"); err != nil {
		return fmt.Errorf("invalid AgentSpec: %w", err)
	}

	if err := validation.MaxLength(s.Description, maxDescriptionLength, "Description"); err != nil {
		return fmt.Errorf("invalid AgentSpec: %w", err)
	}

	if err := validation.NotEmpty(s.ModelID, "ModelID"); err != nil {
		return fmt.Errorf("invalid AgentSpec: %w", err)
	}

	if s.MaxIterations < 0 {
		return fmt.Errorf("invalid AgentSpec: MaxIterations must be ≥ 0, got %d", s.MaxIterations)
	}

	if len(s.Capabilities) > maxCapabilities {
		return fmt.Errorf("invalid AgentSpec: Capabilities has %d entries; maximum is %d",
			len(s.Capabilities), maxCapabilities)
	}
	for i, cap := range s.Capabilities {
		if cap == "" {
			return fmt.Errorf("invalid AgentSpec: Capabilities[%d] must not be empty", i)
		}
	}

	if len(s.Config) > maxConfigEntries {
		return fmt.Errorf("invalid AgentSpec: Config has %d entries; maximum is %d",
			len(s.Config), maxConfigEntries)
	}

	return nil
}

// Normalize applies defaults to an AgentSpec whose fields have been validated.
// It must be called after Validate to ensure the spec's optional fields have
// sensible values before the lifecycle creator uses them.
//
// Currently normalizes:
//   - MaxIterations: 0 → defaultMaxIterations (10)
func Normalize(s *AgentSpec) {
	if s.MaxIterations == 0 {
		s.MaxIterations = defaultMaxIterations
	}
}
