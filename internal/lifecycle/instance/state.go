// Package instance defines AgentInstance — the runtime representation of a live agent.
package instance

import "github.com/veda/agent-runtime/internal/types/runtime"

// allowedTransitions defines the canonical set of valid state transitions for
// an AgentInstance. The outer key is the current state; the inner value is the
// set of target states that are permitted from that state.
//
// This table is the single source of truth for state machine logic. All callers
// must go through TransitionTo, which consults this table.
var allowedTransitions = map[runtime.AgentState]map[runtime.AgentState]bool{
	// From Creating: the creator moves the agent to Initializing.
	runtime.AgentCreating: {
		runtime.AgentInitializing: true,
		runtime.AgentTerminated:   true, // creation failure path
	},

	// From Initializing: the initializer moves the agent to Ready on success
	// or records an error (SetError → Terminated) on failure.
	runtime.AgentInitializing: {
		runtime.AgentReady:      true,
		runtime.AgentTerminated: true, // initialization failure
	},

	// From Ready: agent can be dispatched work (Busy), suspended,
	// checkpointed, or terminated.
	runtime.AgentReady: {
		runtime.AgentBusy:          true,
		runtime.AgentSuspending:    true,
		runtime.AgentCheckpointing: true,
		runtime.AgentStopping:      true,
	},

	// From Busy: execution is ongoing; can complete (→ Ready), suspend, or stop.
	runtime.AgentBusy: {
		runtime.AgentReady:      true,
		runtime.AgentSuspending: true,
		runtime.AgentStopping:   true,
		runtime.AgentTerminated: true, // fatal execution error
	},

	// From Suspending: quiesce completes → Suspended.
	runtime.AgentSuspending: {
		runtime.AgentSuspended:  true,
		runtime.AgentTerminated: true, // suspend failure
	},

	// From Suspended: can be resumed or terminated.
	runtime.AgentSuspended: {
		runtime.AgentResuming: true,
		runtime.AgentStopping: true,
	},

	// From Resuming: resumes back to Ready.
	runtime.AgentResuming: {
		runtime.AgentReady:      true,
		runtime.AgentTerminated: true, // resume failure
	},

	// From Checkpointing: returns to Ready or terminates on failure.
	runtime.AgentCheckpointing: {
		runtime.AgentReady:      true,
		runtime.AgentTerminated: true,
	},

	// From Recovering: returns to Ready or terminates if recovery fails.
	runtime.AgentRecovering: {
		runtime.AgentReady:      true,
		runtime.AgentTerminated: true,
	},

	// From Stopping: cleanup completes → Terminated.
	runtime.AgentStopping: {
		runtime.AgentTerminated: true,
	},

	// Terminated and NonExistent are final; no further transitions allowed.
	runtime.AgentTerminated:  {},
	runtime.AgentNonExistent: {},
}

// isAllowedTransition returns true if transitioning from `from` to `to` is
// permitted by the canonical state machine table.
func isAllowedTransition(from, to runtime.AgentState) bool {
	targets, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}
