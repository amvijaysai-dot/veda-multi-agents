// Package validator provides capability validation mechanisms.
package validator

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

func TestSecurityValidator_ValidateSuccess(t *testing.T) {
	val := NewSecurityValidator()
	ctx := context.Background()

	cap := interfaces.LoadedCapability{
		Metadata: interfaces.CapabilityMetadata{
			ID:                  "cap.safe",
			Version:             "1.0.0",
			RequiredPermissions: []string{"fs:read"},
		},
	}

	res, err := val.Validate(ctx, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsValid {
		t.Errorf("expected valid, got errors: %v", res.Errors)
	}
	if res.RiskScore != 0 {
		t.Errorf("expected risk score 0, got %d", res.RiskScore)
	}
}

func TestSecurityValidator_ValidateMissingMetadata(t *testing.T) {
	val := NewSecurityValidator()
	ctx := context.Background()

	cap := interfaces.LoadedCapability{
		Metadata: interfaces.CapabilityMetadata{}, // missing ID and Version
	}

	res, err := val.Validate(ctx, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsValid {
		t.Error("expected invalid due to missing metadata")
	}
	if len(res.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(res.Errors))
	}
}

func TestSecurityValidator_ValidateDeniedPermission(t *testing.T) {
	val := NewSecurityValidator()
	ctx := context.Background()

	cap := interfaces.LoadedCapability{
		Metadata: interfaces.CapabilityMetadata{
			ID:                  "cap.danger",
			Version:             "1.0.0",
			RequiredPermissions: []string{"system:root"},
		},
	}

	res, err := val.Validate(ctx, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsValid {
		t.Error("expected invalid due to denied permission")
	}
	if res.RiskScore != 100 {
		t.Errorf("expected risk score 100, got %d", res.RiskScore)
	}
}

func TestSecurityValidator_ValidateHeuristicRisk(t *testing.T) {
	val := NewSecurityValidator()
	ctx := context.Background()

	cap := interfaces.LoadedCapability{
		Metadata: interfaces.CapabilityMetadata{
			ID:      "cap.suspicious",
			Version: "1.0.0",
			// Combining network out with fs write triggers heuristic
			RequiredPermissions: []string{"network:outbound", "fs:write"},
		},
	}

	res, err := val.Validate(ctx, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// It's still "valid" by default unless we specifically deny it,
	// but it gets a high risk score.
	if !res.IsValid {
		t.Errorf("expected valid but risky, got errors: %v", res.Errors)
	}

	if res.RiskScore != 50 {
		t.Errorf("expected risk score 50, got %d", res.RiskScore)
	}
}
