// Package lifecycle implements the agent lifecycle subsystem for the VEDA Agent Runtime.
// It provides the Creator, Initializer, and Terminator that move agents through the
// stages defined in Section 4 of the architecture specification.
//
// Package-level dependency rules:
//   - lifecycle depends on lifecycle/spec, lifecycle/instance, and types/runtime.
//   - lifecycle must not import kernel/impl; it may import kernel/interfaces.
//   - All kernel interaction is performed through the interfaces.Kernel interface.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/lifecycle/instance"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// Creator is responsible for creating AgentInstances from AgentSpecs.
// It validates the spec, normalizes defaults, and instantiates the agent
// in the Creating state ready for subsequent initialization.
//
// Creator does not perform initialization; that is the responsibility of
// Initializer (V0.3.04). This separation follows the architecture's distinction
// between the Creation and Initialization phases of the agent lifecycle.
type Creator struct{}

// NewCreator creates and returns a new Creator.
func NewCreator() *Creator {
	return &Creator{}
}

// Create validates and normalizes the provided AgentSpec, then creates and
// returns a new AgentInstance in the Creating state.
//
// Returns an error if:
//   - spec is nil
//   - spec fails validation (see spec.Validate)
//
// On success, the returned AgentInstance is in the AgentCreating state and
// is ready to be passed to Initializer.Initialize.
func (c *Creator) Create(ctx context.Context, s *spec.AgentSpec) (*instance.AgentInstance, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("creator.Create: context cancelled: %w", err)
	}

	if err := spec.Validate(s); err != nil {
		return nil, fmt.Errorf("creator.Create: %w", err)
	}

	spec.Normalize(s)

	inst := instance.New(s)

	// Post-condition: the new instance must be in the Creating state.
	// This invariant is guaranteed by instance.New, but we verify it explicitly
	// to catch any future regression.
	if inst.State() != runtime.AgentCreating {
		return nil, fmt.Errorf("creator.Create: expected new instance to be in Creating state, got %s", inst.State())
	}

	return inst, nil
}
