package hotreload

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
	registryinterfaces "github.com/veda/agent-runtime/internal/registry/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

// CapabilityReloader manages runtime capability updates and hot-swapping.
type CapabilityReloader struct {
	mu       sync.RWMutex
	registry registryinterfaces.CapabilityRegistry
	kernel   interfaces.Kernel
}

// NewCapabilityReloader creates a new CapabilityReloader.
func NewCapabilityReloader(reg registryinterfaces.CapabilityRegistry, kernel interfaces.Kernel) *CapabilityReloader {
	return &CapabilityReloader{
		registry: reg,
		kernel:   kernel,
	}
}

// ReloadCapability attempts to register or update a capability in the registry.
// If registration succeeds, an event is published for hot-reloading components.
func (cr *CapabilityReloader) ReloadCapability(ctx context.Context, metadata registryinterfaces.CapabilityMetadata) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 1. Deregister existing capability if it exists (for a true hot-swap)
	_ = cr.registry.Deregister(ctx, metadata.ID, metadata.Version)

	// 2. Validate and Register the new capability
	if err := cr.registry.Register(ctx, metadata); err != nil {
		// Rollback or retain failure state is handled naturally if Register fails,
		// but since we deregistered, the rollback strategy is complex.
		// For v1.0, we just return the error. A true rollback would require caching the old capability.
		return fmt.Errorf("failed to reload capability %q: %w", metadata.ID, err)
	}

	// 3. Notify components of the update
	if cr.kernel != nil {
		evt := event.NewBaseEvent(
			fmt.Sprintf("cap-reload-%d", time.Now().UnixNano()),
			event.TypeCapabilityLoaded,
			"CapabilityReloader",
			event.WithPayload(metadata),
		)
		_ = cr.kernel.PublishEvent(evt)
	}

	return nil
}
