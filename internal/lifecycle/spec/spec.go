// Package spec defines the AgentSpec structure and related validation logic.
// An AgentSpec is the complete description of an agent that the runtime uses to
// create an AgentInstance. It is analogous to a container spec in container runtimes.
//
// Callers must validate an AgentSpec with Validate before passing it to the lifecycle
// creator. Once validated, a spec is treated as immutable.
package spec

import "time"

// ResourceLimits defines the compute resource constraints for an agent.
// All fields are optional; zero values mean "no limit".
type ResourceLimits struct {
	// MaxMemoryBytes is the maximum resident memory (bytes) the agent may use.
	MaxMemoryBytes int64
	// MaxCPUMillicores is the maximum CPU share in millicores (1000 = 1 core).
	MaxCPUMillicores int64
	// MaxTurnDuration is the maximum wall-clock time allowed per reasoning turn.
	MaxTurnDuration time.Duration
}

// AgentSpec is the complete declarative description of an agent.
// It describes everything the lifecycle manager needs to create and initialize
// the corresponding AgentInstance.
//
// All required fields must be non-empty; optional fields may be zero-valued.
type AgentSpec struct {
	// ID is the globally unique identifier for this agent.
	// It must be a non-empty string containing only alphanumeric characters,
	// underscores, and hyphens (validated by IsValidID).
	ID string

	// Name is a human-readable display name for the agent.
	// It must be non-empty and may contain arbitrary printable characters.
	Name string

	// Description is an optional human-readable description of the agent's purpose.
	Description string

	// ModelID is the identifier of the model to use for reasoning steps.
	// It must be non-empty; the model must be available via the Model Interface.
	ModelID string

	// Capabilities is the list of capability identifiers that should be bound
	// to this agent on initialization. May be empty if the agent needs no tools.
	Capabilities []string

	// Config is a map of arbitrary string key-value pairs used to configure
	// the agent and its bound capabilities. May be nil.
	Config map[string]string

	// Limits defines the resource constraints for this agent.
	// Zero values mean "no limit" for the corresponding resource.
	Limits ResourceLimits

	// MaxIterations is the maximum number of reasoning/acting cycles the agent
	// is allowed to perform per turn. Must be > 0; defaults to 10 if not set.
	MaxIterations int

	// Tags are arbitrary key-value pairs for discovery and routing.
	// They are not interpreted by the runtime.
	Tags map[string]string
}
