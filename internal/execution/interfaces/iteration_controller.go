// Package interfaces defines the contracts for all execution engine components.
package interfaces

// IterationDecision signals whether the ReAct loop should continue or terminate.
type IterationDecision int

const (
	// Continue means another ReAct cycle should be executed.
	Continue IterationDecision = iota

	// Terminate means the loop should stop and return the current result.
	Terminate
)

// String returns the human-readable label for the decision.
func (d IterationDecision) String() string {
	switch d {
	case Continue:
		return "continue"
	case Terminate:
		return "terminate"
	default:
		return "unknown"
	}
}

// IterationController tracks ReAct loop state and decides whether another cycle
// should run. The controller is stateless across turns; callers must call Reset()
// between turns when reusing the same instance.
type IterationController interface {
	// Decide evaluates the current reasoning output and iteration count and
	// returns the appropriate IterationDecision.
	//
	// Parameters:
	//   - output: the latest ReasoningOutput from the reasoning engine.
	//   - iteration: 1-based index of the cycle just completed.
	//   - maxIterations: the per-turn cap configured on the agent.
	Decide(output ReasoningOutput, iteration, maxIterations int) IterationDecision

	// Reset clears accumulated state so the controller can be reused across turns.
	Reset()
}
