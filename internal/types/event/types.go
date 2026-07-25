package event

// Type represents the category of an event.
type Type string

// Event type constants organized by category.
const (
	// Lifecycle events
	TypeAgentCreated      Type = "agent.created"
	TypeAgentInitialized  Type = "agent.initialized"
	TypeAgentRegistered   Type = "agent.registered"
	TypeAgentActivated    Type = "agent.activated"
	TypeAgentSuspended    Type = "agent.suspended"
	TypeAgentResumed      Type = "agent.resumed"
	TypeAgentCheckpointed Type = "agent.checkpointed"
	TypeAgentRecovered    Type = "agent.recovered"
	TypeAgentStopping     Type = "agent.stopping"
	TypeAgentStopped      Type = "agent.stopped"

	// Execution events
	TypeTurnStarted   Type = "turn.started"
	TypeTurnCompleted Type = "turn.completed"
	TypeToolInvoked   Type = "tool.invoked"
	TypeToolCompleted Type = "tool.completed"
	TypeToolFailed    Type = "tool.failed"
	TypeLLMInvoked    Type = "llm.invoked"
	TypeLLMCompleted  Type = "llm.completed"
	TypeLLMFailed     Type = "llm.failed"
	TypeContextBuilt  Type = "context.built"

	// Capability events
	TypeCapabilityLoaded   Type = "capability.loaded"
	TypeCapabilityUnloaded Type = "capability.unloaded"
	TypeCapabilityBound    Type = "capability.bound"
	TypeCapabilityUnbound  Type = "capability.unbound"
	TypeCapabilityInvoked  Type = "capability.invoked"
	TypeCapabilityFailed   Type = "capability.failed"

	// Memory events
	TypeMemoryStored       Type = "memory.stored"
	TypeMemoryRetrieved    Type = "memory.retrieved"
	TypeMemoryConsolidated Type = "memory.consolidated"
	TypeMemoryForgotten    Type = "memory.forgotten"
	TypeMemoryScrubbed     Type = "memory.scrubbed"

	// Model events
	TypeModelLoaded    Type = "model.loaded"
	TypeModelUnloaded  Type = "model.unloaded"
	TypeModelInvoked   Type = "model.invoked"
	TypeModelCompleted Type = "model.completed"
	TypeModelFailed    Type = "model.failed"

	// Security events
	TypeAccessGranted Type = "security.access.granted"
	TypeAccessDenied  Type = "security.access.denied"
	TypeViolation     Type = "security.violation"
	TypeSandboxBreach Type = "security.sandbox.breach"

	// Resource events
	TypeCPUExceeded    Type = "resource.cpu.exceeded"
	TypeMemoryExceeded Type = "resource.memory.exceeded"
	TypeIOExceeded     Type = "resource.io.exceeded"

	// Health events
	TypeHeartbeat     Type = "health.heartbeat"
	TypeDegradation   Type = "health.degradation"
	TypeRecovery      Type = "health.recovery"
	TypeComponentFail Type = "health.component.fail"
)

// String returns the string representation of the event type.
func (t Type) String() string {
	return string(t)
}
