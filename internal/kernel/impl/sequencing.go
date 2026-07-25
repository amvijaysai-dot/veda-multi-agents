// Package impl provides the concrete implementation of the VEDA Agent Runtime kernel.
package impl

import (
	"context"
	"fmt"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
)

// Sequencer manages ordered initialization and shutdown of subsystems.
// It enforces the dependency invariant that subsystems are started in
// registration order and stopped in reverse order, preventing partial
// initialization states.
//
// Sequencer is not safe for concurrent use; the caller (Kernel) must
// hold appropriate locks when invoking its methods.
type Sequencer struct {
	registry *Registry
}

// newSequencer creates a Sequencer bound to the given registry.
func newSequencer(registry *Registry) *Sequencer {
	return &Sequencer{registry: registry}
}

// InitAll initializes all subsystems in the registry in registration order.
//
// If any subsystem fails to initialize, all previously initialized subsystems
// are stopped in reverse order (rollback) before the error is returned.
// This ensures the system is never left in a partially initialized state.
//
// Returns an error wrapping the first subsystem failure encountered.
func (s *Sequencer) InitAll(ctx context.Context) error {
	names := s.registry.Names()
	initialized := make([]string, 0, len(names))

	for _, name := range names {
		sub, err := s.registry.Get(name)
		if err != nil {
			s.stopRange(ctx, initialized)
			return fmt.Errorf("sequencer.InitAll: could not retrieve subsystem %q: %w", name, err)
		}

		if err := sub.Init(ctx); err != nil {
			s.stopRange(ctx, initialized)
			return fmt.Errorf("sequencer.InitAll: subsystem %q failed to initialize: %w", name, err)
		}
		initialized = append(initialized, name)
	}
	return nil
}

// StartAll starts all subsystems in the registry in registration order.
//
// Assumes InitAll has been called successfully on all subsystems.
// If any subsystem fails to start, already-started subsystems are stopped
// in reverse order before the error is returned.
//
// Returns an error wrapping the first subsystem failure encountered.
func (s *Sequencer) StartAll(ctx context.Context) error {
	names := s.registry.Names()
	started := make([]string, 0, len(names))

	for _, name := range names {
		sub, err := s.registry.Get(name)
		if err != nil {
			s.stopRange(ctx, started)
			return fmt.Errorf("sequencer.StartAll: could not retrieve subsystem %q: %w", name, err)
		}

		if err := sub.Start(ctx); err != nil {
			s.stopRange(ctx, started)
			return fmt.Errorf("sequencer.StartAll: subsystem %q failed to start: %w", name, err)
		}
		started = append(started, name)
	}
	return nil
}

// StopAll stops all subsystems in the registry in reverse registration order.
//
// All subsystems are stopped regardless of individual failures. All stop errors
// are collected and returned as a combined error. Returns nil if all subsystems
// stop cleanly.
func (s *Sequencer) StopAll(ctx context.Context) error {
	names := s.registry.Names()
	var errs []string

	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		sub, err := s.registry.Get(name)
		if err != nil {
			// Subsystem may have been unregistered during shutdown; skip it.
			continue
		}

		if err := sub.Stop(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("subsystem %q: %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("sequencer.StopAll encountered %d stop error(s): %s",
			len(errs), joinErrors(errs))
	}
	return nil
}

// DoubleInitGuard prevents double-initialization of the same subsystem.
// It tracks which subsystem names have been initialized and returns an error
// if the same name is initialized a second time.
type DoubleInitGuard struct {
	initialized map[string]bool
}

// newDoubleInitGuard creates a new DoubleInitGuard.
func newDoubleInitGuard() *DoubleInitGuard {
	return &DoubleInitGuard{initialized: make(map[string]bool)}
}

// Check returns an error if name has already been initialized; otherwise it
// records it as initialized and returns nil.
func (g *DoubleInitGuard) Check(name string) error {
	if g.initialized[name] {
		return fmt.Errorf("subsystem %q has already been initialized", name)
	}
	g.initialized[name] = true
	return nil
}

// Reset clears the guard, allowing the named subsystem to be re-initialized.
// This is used during recovery scenarios.
func (g *DoubleInitGuard) Reset(name string) {
	delete(g.initialized, name)
}

// IsInitialized reports whether name has been recorded as initialized.
func (g *DoubleInitGuard) IsInitialized(name string) bool {
	return g.initialized[name]
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// stopRange stops a list of subsystems in reverse order, ignoring errors.
// Used during rollback when a later subsystem fails to initialize or start.
func (s *Sequencer) stopRange(ctx context.Context, names []string) {
	for i := len(names) - 1; i >= 0; i-- {
		sub, err := s.registry.Get(names[i])
		if err != nil {
			continue
		}
		_ = sub.Stop(ctx)
	}
}

// OrderedSubsystems returns the subsystems in registration order as a slice of
// (name, Subsystem) pairs. Useful for inspection during testing.
func (s *Sequencer) OrderedSubsystems() []namedSubsystem {
	names := s.registry.Names()
	result := make([]namedSubsystem, 0, len(names))
	for _, name := range names {
		sub, err := s.registry.Get(name)
		if err != nil {
			continue
		}
		result = append(result, namedSubsystem{Name: name, Subsystem: sub})
	}
	return result
}

// namedSubsystem pairs a subsystem with its registry name.
type namedSubsystem struct {
	Name      string
	Subsystem interfaces.Subsystem
}

// joinErrors joins error strings with "; " for human-readable combined error messages.
func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "; "
		}
		result += e
	}
	return result
}
