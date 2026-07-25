// Package spec defines the AgentSpec structure and related validation logic.
package spec

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// validSpec returns a minimal, fully-valid AgentSpec for use in tests.
func validSpec() *AgentSpec {
	return &AgentSpec{
		ID:      "agent-001",
		Name:    "Test Agent",
		ModelID: "gpt-4",
	}
}

// ---------------------------------------------------------------------------
// Validate — happy paths
// ---------------------------------------------------------------------------

func TestValidate_MinimalValidSpec(t *testing.T) {
	if err := Validate(validSpec()); err != nil {
		t.Fatalf("expected no error for minimal valid spec, got: %v", err)
	}
}

func TestValidate_FullValidSpec(t *testing.T) {
	s := &AgentSpec{
		ID:            "agent-full-001",
		Name:          "Full Agent",
		Description:   "A complete spec for testing.",
		ModelID:       "claude-3",
		Capabilities:  []string{"search", "calculator"},
		MaxIterations: 20,
		Config:        map[string]string{"key": "value"},
		Tags:          map[string]string{"env": "test"},
		Limits: ResourceLimits{
			MaxMemoryBytes:   512 * 1024 * 1024,
			MaxCPUMillicores: 1000,
			MaxTurnDuration:  30 * time.Second,
		},
	}
	if err := Validate(s); err != nil {
		t.Fatalf("expected no error for full valid spec, got: %v", err)
	}
}

func TestValidate_ZeroMaxIterationsIsValid(t *testing.T) {
	s := validSpec()
	s.MaxIterations = 0
	if err := Validate(s); err != nil {
		t.Fatalf("expected MaxIterations=0 to pass validation (normalized later), got: %v", err)
	}
}

func TestValidate_EmptyCapabilitiesIsValid(t *testing.T) {
	s := validSpec()
	s.Capabilities = nil
	if err := Validate(s); err != nil {
		t.Fatalf("expected nil Capabilities to pass validation, got: %v", err)
	}
}

func TestValidate_EmptyConfigIsValid(t *testing.T) {
	s := validSpec()
	s.Config = nil
	if err := Validate(s); err != nil {
		t.Fatalf("expected nil Config to pass validation, got: %v", err)
	}
}

func TestValidate_EmptyDescriptionIsValid(t *testing.T) {
	s := validSpec()
	s.Description = ""
	if err := Validate(s); err != nil {
		t.Fatalf("expected empty Description to pass validation, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Validate — nil input
// ---------------------------------------------------------------------------

func TestValidate_NilSpecReturnsError(t *testing.T) {
	if err := Validate(nil); err == nil {
		t.Fatal("expected error for nil spec, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validate — ID field
// ---------------------------------------------------------------------------

func TestValidate_EmptyIDReturnsError(t *testing.T) {
	s := validSpec()
	s.ID = ""
	if err := Validate(s); err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}
}

func TestValidate_InvalidIDCharactersReturnsError(t *testing.T) {
	s := validSpec()
	s.ID = "invalid id!" // spaces and ! not allowed
	if err := Validate(s); err == nil {
		t.Fatal("expected error for ID with invalid characters, got nil")
	}
}

func TestValidate_ValidIDFormats(t *testing.T) {
	cases := []string{"a", "agent-001", "agent_001", "AGENT001", "a1-b2_C3"}
	for _, id := range cases {
		s := validSpec()
		s.ID = id
		if err := Validate(s); err != nil {
			t.Errorf("expected valid ID %q to pass, got: %v", id, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Validate — Name field
// ---------------------------------------------------------------------------

func TestValidate_EmptyNameReturnsError(t *testing.T) {
	s := validSpec()
	s.Name = ""
	if err := Validate(s); err == nil {
		t.Fatal("expected error for empty Name, got nil")
	}
}

func TestValidate_WhitespaceOnlyNameReturnsError(t *testing.T) {
	s := validSpec()
	s.Name = "   "
	if err := Validate(s); err == nil {
		t.Fatal("expected error for whitespace-only Name, got nil")
	}
}

func TestValidate_NameTooLongReturnsError(t *testing.T) {
	s := validSpec()
	s.Name = strings.Repeat("a", maxNameLength+1)
	if err := Validate(s); err == nil {
		t.Fatalf("expected error for Name exceeding %d chars, got nil", maxNameLength)
	}
}

func TestValidate_NameAtMaxLengthIsValid(t *testing.T) {
	s := validSpec()
	s.Name = strings.Repeat("a", maxNameLength)
	if err := Validate(s); err != nil {
		t.Fatalf("expected Name of exactly %d chars to pass, got: %v", maxNameLength, err)
	}
}

// ---------------------------------------------------------------------------
// Validate — ModelID field
// ---------------------------------------------------------------------------

func TestValidate_EmptyModelIDReturnsError(t *testing.T) {
	s := validSpec()
	s.ModelID = ""
	if err := Validate(s); err == nil {
		t.Fatal("expected error for empty ModelID, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validate — MaxIterations field
// ---------------------------------------------------------------------------

func TestValidate_NegativeMaxIterationsReturnsError(t *testing.T) {
	s := validSpec()
	s.MaxIterations = -1
	if err := Validate(s); err == nil {
		t.Fatal("expected error for negative MaxIterations, got nil")
	}
}

func TestValidate_PositiveMaxIterationsIsValid(t *testing.T) {
	s := validSpec()
	s.MaxIterations = 50
	if err := Validate(s); err != nil {
		t.Fatalf("expected positive MaxIterations to pass, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Validate — Capabilities field
// ---------------------------------------------------------------------------

func TestValidate_EmptyCapabilityEntryReturnsError(t *testing.T) {
	s := validSpec()
	s.Capabilities = []string{"search", "", "calculator"}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for empty capability entry, got nil")
	}
}

func TestValidate_TooManyCapabilitiesReturnsError(t *testing.T) {
	s := validSpec()
	caps := make([]string, maxCapabilities+1)
	for i := range caps {
		caps[i] = "cap"
	}
	s.Capabilities = caps
	if err := Validate(s); err == nil {
		t.Fatalf("expected error for >%d capabilities, got nil", maxCapabilities)
	}
}

func TestValidate_CapabilitiesAtMaxLimitIsValid(t *testing.T) {
	s := validSpec()
	caps := make([]string, maxCapabilities)
	for i := range caps {
		caps[i] = "cap"
	}
	s.Capabilities = caps
	if err := Validate(s); err != nil {
		t.Fatalf("expected exactly %d capabilities to pass, got: %v", maxCapabilities, err)
	}
}

// ---------------------------------------------------------------------------
// Validate — Config field
// ---------------------------------------------------------------------------

func TestValidate_TooManyConfigEntriesReturnsError(t *testing.T) {
	s := validSpec()
	s.Config = make(map[string]string, maxConfigEntries+1)
	for i := 0; i <= maxConfigEntries; i++ {
		s.Config[strings.Repeat("k", i+1)] = "v"
	}
	if err := Validate(s); err == nil {
		t.Fatalf("expected error for >%d config entries, got nil", maxConfigEntries)
	}
}

// ---------------------------------------------------------------------------
// Validate — Description field
// ---------------------------------------------------------------------------

func TestValidate_DescriptionTooLongReturnsError(t *testing.T) {
	s := validSpec()
	s.Description = strings.Repeat("d", maxDescriptionLength+1)
	if err := Validate(s); err == nil {
		t.Fatalf("expected error for Description exceeding %d chars, got nil", maxDescriptionLength)
	}
}

// ---------------------------------------------------------------------------
// Normalize
// ---------------------------------------------------------------------------

func TestNormalize_ZeroMaxIterationsDefaulted(t *testing.T) {
	s := validSpec()
	s.MaxIterations = 0
	Normalize(s)
	if s.MaxIterations != defaultMaxIterations {
		t.Errorf("expected MaxIterations=%d after Normalize, got %d", defaultMaxIterations, s.MaxIterations)
	}
}

func TestNormalize_NonZeroMaxIterationsPreserved(t *testing.T) {
	s := validSpec()
	s.MaxIterations = 25
	Normalize(s)
	if s.MaxIterations != 25 {
		t.Errorf("expected MaxIterations=25 preserved after Normalize, got %d", s.MaxIterations)
	}
}
