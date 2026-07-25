// Package integration provides the top-level orchestration for all VEDA runtime subsystems.
package integration

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/config"
	"github.com/veda/agent-runtime/internal/kernel/impl"
	kernelinterfaces "github.com/veda/agent-runtime/internal/kernel/interfaces"
	"github.com/veda/agent-runtime/internal/lifecycle"
	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	recoveryinterfaces "github.com/veda/agent-runtime/internal/recovery/interfaces"
	registryinterfaces "github.com/veda/agent-runtime/internal/registry/interfaces"
)

// Subsystems holds references to all initialized top-level components.
type Subsystems struct {
	Kernel       kernelinterfaces.Kernel
	Config       *config.Config
	Execution    any
	Memory       any
	Planner      any
	Registry     registryinterfaces.CapabilityRegistry
	Recovery     recoveryinterfaces.RecoveryCoordinator
}

// RuntimeIntegrator orchestrates the initialization and lifecycle of the full system.
type RuntimeIntegrator struct {
	subs *Subsystems
}

// NewRuntimeIntegrator creates a new integrator instance.
func NewRuntimeIntegrator() *RuntimeIntegrator {
	return &RuntimeIntegrator{
		subs: &Subsystems{},
	}
}

// InitSubsystems configures and wires up all subsystems in the correct order.
// Since the backlog dictates this handles dependency injection, we accept the concrete implementations
// or factories here to avoid circular dependencies and tight coupling with concrete packages.
func (r *RuntimeIntegrator) InitSubsystems(
	ctx context.Context,
	cfg *config.Config,
	exec any,
	mem any,
	plan any,
	reg registryinterfaces.CapabilityRegistry,
	rec recoveryinterfaces.RecoveryCoordinator,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 1. Config
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	r.subs.Config = cfg

	// 2. Kernel
	k := impl.NewKernel()
	r.subs.Kernel = k

	// 3. Observability
	// We'll set up standard loggers via hooks if needed, though they usually self-register

	// 4. Register Subsystems
	r.subs.Execution = exec
	r.subs.Memory = mem
	r.subs.Planner = plan
	r.subs.Registry = reg
	r.subs.Recovery = rec

	// The Kernel manages the agent lifecycle. If it requires explicit wiring, we do it here.
	if r.subs.Registry != nil {
		// e.g. initialize the registry via some config
	}

	// 5. Initialize the Kernel
	if err := r.subs.Kernel.Init(ctx); err != nil {
		return fmt.Errorf("failed to init kernel: %w", err)
	}

	return nil
}

// Start brings up the runtime and starts accepting agents.
func (r *RuntimeIntegrator) Start(ctx context.Context) error {
	if r.subs.Kernel == nil {
		return fmt.Errorf("kernel not initialized")
	}
	return r.subs.Kernel.Start(ctx)
}

// Stop gracefully shuts down the runtime and all subsystems.
func (r *RuntimeIntegrator) Stop(ctx context.Context) error {
	if r.subs.Kernel == nil {
		return nil
	}
	return r.subs.Kernel.Stop(ctx)
}

// CreateAgent leverages the lifecycle package to spin up a new agent.
func (r *RuntimeIntegrator) CreateAgent(ctx context.Context, specification *spec.AgentSpec) (string, error) {
	creator := lifecycle.NewCreator()
	agent, err := creator.Create(ctx, specification)
	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}
	return agent.ID(), nil
}
