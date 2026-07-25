// Package lifecycle provides capability lifecycle coordination.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/registry/interfaces"
)

// LifecycleManager coordinates discovering, loading, validating, and
// registering capabilities.
type LifecycleManager struct {
	loader    interfaces.CapabilityLoader
	validator interfaces.CapabilityValidator
	registry  interfaces.CapabilityRegistry
}

// NewLifecycleManager creates a new LifecycleManager.
func NewLifecycleManager(
	loader interfaces.CapabilityLoader,
	validator interfaces.CapabilityValidator,
	registry interfaces.CapabilityRegistry,
) *LifecycleManager {
	if loader == nil || validator == nil || registry == nil {
		panic("lifecycle manager components cannot be nil")
	}
	return &LifecycleManager{
		loader:    loader,
		validator: validator,
		registry:  registry,
	}
}

// Boot scans a root source, loads all discovered capabilities, validates them,
// and registers the valid ones. It returns the number of successfully registered
// capabilities and a list of errors for those that failed.
func (m *LifecycleManager) Boot(ctx context.Context, rootSource interfaces.CapabilitySource) (int, []error) {
	sources, err := m.loader.Discover(ctx, rootSource)
	if err != nil {
		return 0, []error{fmt.Errorf("discovery failed: %w", err)}
	}

	var registered int
	var errs []error

	for _, src := range sources {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}

		// 1. Load
		loadedCap, err := m.loader.Load(ctx, src)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to load %s: %w", src.URI, err))
			continue
		}

		// 2. Validate
		valRes, err := m.validator.Validate(ctx, loadedCap)
		if err != nil {
			errs = append(errs, fmt.Errorf("validation error for %s: %w", loadedCap.Metadata.ID, err))
			continue
		}
		if !valRes.IsValid {
			errs = append(errs, fmt.Errorf("capability %s is invalid: %v", loadedCap.Metadata.ID, valRes.Errors))
			continue
		}

		// 3. Register
		err = m.registry.Register(ctx, loadedCap.Metadata)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to register %s: %w", loadedCap.Metadata.ID, err))
			continue
		}

		registered++
	}

	return registered, errs
}
