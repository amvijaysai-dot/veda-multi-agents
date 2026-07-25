// Package runtime provides foundational runtime type definitions.
package runtime

// RuntimeStatus represents the current operational status of the runtime.
type RuntimeStatus int

const (
	StatusUninitialized RuntimeStatus = iota
	StatusInitializing
	StatusReady
	StatusBusy
	StatusSuspending
	StatusSuspended
	StatusResuming
	StatusShuttingDown
	StatusTerminated
)

// String returns the string representation of the runtime status.
func (s RuntimeStatus) String() string {
	switch s {
	case StatusUninitialized:
		return "uninitialized"
	case StatusInitializing:
		return "initializing"
	case StatusReady:
		return "ready"
	case StatusBusy:
		return "busy"
	case StatusSuspending:
		return "suspending"
	case StatusSuspended:
		return "suspended"
	case StatusResuming:
		return "resuming"
	case StatusShuttingDown:
		return "shutting_down"
	case StatusTerminated:
		return "terminated"
	default:
		return "unknown"
	}
}

// AgentState represents the current state of an agent instance.
type AgentState int

const (
	AgentNonExistent AgentState = iota
	AgentCreating
	AgentInitializing
	AgentReady
	AgentBusy
	AgentSuspending
	AgentSuspended
	AgentResuming
	AgentCheckpointing
	AgentRecovering
	AgentStopping
	AgentTerminated
)

// String returns the string representation of the agent state.
func (s AgentState) String() string {
	switch s {
	case AgentNonExistent:
		return "non_existent"
	case AgentCreating:
		return "creating"
	case AgentInitializing:
		return "initializing"
	case AgentReady:
		return "ready"
	case AgentBusy:
		return "busy"
	case AgentSuspending:
		return "suspending"
	case AgentSuspended:
		return "suspended"
	case AgentResuming:
		return "resuming"
	case AgentCheckpointing:
		return "checkpointing"
	case AgentRecovering:
		return "recovering"
	case AgentStopping:
		return "stopping"
	case AgentTerminated:
		return "terminated"
	default:
		return "unknown"
	}
}

// Version represents the version information for a component.
type Version struct {
	Major int
	Minor int
	Patch int
}
