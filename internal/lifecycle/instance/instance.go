// Package instance defines AgentInstance — the runtime representation of a live agent.
// An AgentInstance is created from an AgentSpec by the lifecycle creator and tracks
// the agent through its complete lifecycle: creating → initializing → ready → ...
// → stopping → terminated.
//
// Package instance is internal to the lifecycle subsystem. External packages must
// not import this package directly; they should interact via lifecycle interfaces.
package instance

import (
	"fmt"
	"sync"
	"time"

	"github.com/veda/agent-runtime/internal/lifecycle/spec"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// AgentInstance is the runtime representation of a live agent.
// It holds a reference to the AgentSpec from which it was created and tracks
// all mutable lifecycle state.
//
// AgentInstance is safe for concurrent use.
type AgentInstance struct {
	mu        sync.RWMutex
	id        string
	spec      *spec.AgentSpec
	state     runtime.AgentState
	createdAt time.Time
	updatedAt time.Time
	// errorMsg records the last error message when state is AgentError-adjacent.
	errorMsg string
}

// New creates and returns a new AgentInstance in the Creating state.
// The provided spec must have already been validated and normalized.
func New(s *spec.AgentSpec) *AgentInstance {
	now := time.Now().UTC()
	return &AgentInstance{
		id:        s.ID,
		spec:      s,
		state:     runtime.AgentCreating,
		createdAt: now,
		updatedAt: now,
	}
}

// ID returns the agent's unique identifier.
func (a *AgentInstance) ID() string {
	return a.id
}

// Spec returns the AgentSpec from which this instance was created.
// The returned pointer must not be modified.
func (a *AgentInstance) Spec() *spec.AgentSpec {
	return a.spec
}

// State returns the current lifecycle state of the agent.
// Safe to call concurrently.
func (a *AgentInstance) State() runtime.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// CreatedAt returns the UTC timestamp when this instance was created.
func (a *AgentInstance) CreatedAt() time.Time {
	return a.createdAt
}

// UpdatedAt returns the UTC timestamp of the most recent state transition.
// Safe to call concurrently.
func (a *AgentInstance) UpdatedAt() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.updatedAt
}

// ErrorMessage returns the error message recorded during the last failed
// state transition, or an empty string if the agent has never entered an
// error-adjacent state.
func (a *AgentInstance) ErrorMessage() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.errorMsg
}

// TransitionTo attempts to move the agent to the target state.
// It validates that the transition is allowed from the current state using the
// canonical allowed-transitions table. Returns an error if the transition is
// not permitted.
//
// TransitionTo is safe for concurrent use; it holds the write lock for the
// duration of the state update.
func (a *AgentInstance) TransitionTo(target runtime.AgentState) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !isAllowedTransition(a.state, target) {
		return fmt.Errorf("agent %q: invalid state transition %s → %s",
			a.id, a.state, target)
	}

	a.state = target
	a.updatedAt = time.Now().UTC()
	a.errorMsg = ""
	return nil
}

// SetError records an error message and transitions the agent to AgentTerminated
// if it is in a recoverable in-progress state. The error message is preserved for
// diagnostic purposes.
//
// SetError is called by the initializer and terminator when an unrecoverable
// error occurs. It does not follow the normal transition table; instead, it forces
// the agent into a terminal error-acknowledged state.
//
// The error state is represented via AgentTerminated + a non-empty ErrorMessage.
// Future milestones may introduce a dedicated AgentError state.
func (a *AgentInstance) SetError(msg string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorMsg = msg
	a.state = runtime.AgentTerminated
	a.updatedAt = time.Now().UTC()
}
