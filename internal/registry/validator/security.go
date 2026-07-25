// Package validator provides capability validation mechanisms.
package validator

import (
	"context"
	"strings"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

// SecurityValidator implements interfaces.CapabilityValidator.
// It checks capabilities for known dangerous permission combinations
// or patterns before they are allowed into the registry.
type SecurityValidator struct {
	// In v0.8+ this would be loaded from a dynamic policy configuration.
	deniedPermissions map[string]bool
}

// NewSecurityValidator creates a new SecurityValidator.
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		deniedPermissions: map[string]bool{
			"system:root":   true,
			"system:reboot": true,
			"fs:write_all":  true,
		},
	}
}

// Validate checks the capability metadata for security risks.
func (v *SecurityValidator) Validate(ctx context.Context, cap interfaces.LoadedCapability) (interfaces.ValidationResult, error) {
	if err := ctx.Err(); err != nil {
		return interfaces.ValidationResult{}, err
	}

	var errors []string
	riskScore := 0

	// 1. Basic Metadata Validation
	if cap.Metadata.ID == "" {
		errors = append(errors, "missing capability ID")
	}
	if cap.Metadata.Version == "" {
		errors = append(errors, "missing capability version")
	}

	// 2. Permission Validation
	networkPerms := 0
	fsWritePerms := 0
	for _, perm := range cap.Metadata.RequiredPermissions {
		// Check for outright denied permissions
		if v.deniedPermissions[perm] {
			errors = append(errors, "requests denied permission: "+perm)
			riskScore = 100
		}

		if strings.HasPrefix(perm, "network:") {
			networkPerms++
		}
		if strings.HasPrefix(perm, "fs:write") {
			fsWritePerms++
		}
	}

	// 3. Heuristic Risk Scoring
	// E.g., combining network access with filesystem write increases risk profile.
	if networkPerms > 0 && fsWritePerms > 0 {
		riskScore += 50
	} else if networkPerms > 0 || fsWritePerms > 0 {
		riskScore += 20
	}

	// Any errors mean it's invalid
	isValid := len(errors) == 0

	return interfaces.ValidationResult{
		IsValid:   isValid,
		Errors:    errors,
		RiskScore: riskScore,
	}, nil
}

var _ interfaces.CapabilityValidator = (*SecurityValidator)(nil)
