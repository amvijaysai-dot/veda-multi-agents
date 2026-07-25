// Package interfaces defines the capability registry contracts.
package interfaces

import "context"

// ValidationResult represents the outcome of a capability validation check.
type ValidationResult struct {
	IsValid bool
	Errors  []string
	// RiskScore is a heuristic (0-100) indicating potential security risk.
	RiskScore int
}

// CapabilityValidator screens loaded capabilities for security and compatibility.
type CapabilityValidator interface {
	// Validate checks a capability manifest for malicious patterns, unauthorized
	// permission requests, missing dependencies, and runtime compatibility.
	Validate(ctx context.Context, cap LoadedCapability) (ValidationResult, error)
}
