// Package event provides foundational event types for the runtime event system.
package event

import "time"

// Event is the core interface that all events in the runtime must implement.
type Event interface {
	// ID returns the unique identifier for this event.
	ID() string

	// Type returns the type of this event.
	Type() Type

	// Timestamp returns when this event was created.
	Timestamp() time.Time

	// Source returns the component that created this event.
	Source() string

	// AgentID returns the agent identifier associated with this event, if any.
	AgentID() string

	// SessionID returns the session identifier associated with this event, if any.
	SessionID() string

	// CorrelationID returns the correlation identifier for tracing related events.
	CorrelationID() string

	// CausationID returns the identifier of the event that caused this event.
	CausationID() string

	// Payload returns the event-specific data.
	Payload() interface{}
}

// BaseEvent provides a default implementation of the Event interface.
type BaseEvent struct {
	id            string
	eventType     Type
	timestamp     time.Time
	source        string
	agentID       string
	sessionID     string
	correlationID string
	causationID   string
	payload       interface{}
}

// NewBaseEvent creates a new BaseEvent with the given parameters.
func NewBaseEvent(id string, eventType Type, source string, opts ...EventOption) *BaseEvent {
	e := &BaseEvent{
		id:        id,
		eventType: eventType,
		timestamp: time.Now().UTC(),
		source:    source,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// EventOption is a function that configures a BaseEvent.
type EventOption func(*BaseEvent)

// WithAgentID sets the agent identifier for the event.
func WithAgentID(agentID string) EventOption {
	return func(e *BaseEvent) {
		e.agentID = agentID
	}
}

// WithSessionID sets the session identifier for the event.
func WithSessionID(sessionID string) EventOption {
	return func(e *BaseEvent) {
		e.sessionID = sessionID
	}
}

// WithCorrelationID sets the correlation identifier for the event.
func WithCorrelationID(correlationID string) EventOption {
	return func(e *BaseEvent) {
		e.correlationID = correlationID
	}
}

// WithCausationID sets the causation identifier for the event.
func WithCausationID(causationID string) EventOption {
	return func(e *BaseEvent) {
		e.causationID = causationID
	}
}

// WithPayload sets the payload for the event.
func WithPayload(payload interface{}) EventOption {
	return func(e *BaseEvent) {
		e.payload = payload
	}
}

// ID returns the unique identifier for this event.
func (e *BaseEvent) ID() string {
	return e.id
}

// Type returns the type of this event.
func (e *BaseEvent) Type() Type {
	return e.eventType
}

// Timestamp returns when this event was created.
func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// Source returns the component that created this event.
func (e *BaseEvent) Source() string {
	return e.source
}

// AgentID returns the agent identifier associated with this event, if any.
func (e *BaseEvent) AgentID() string {
	return e.agentID
}

// SessionID returns the session identifier associated with this event, if any.
func (e *BaseEvent) SessionID() string {
	return e.sessionID
}

// CorrelationID returns the correlation identifier for tracing related events.
func (e *BaseEvent) CorrelationID() string {
	return e.correlationID
}

// CausationID returns the identifier of the event that caused this event.
func (e *BaseEvent) CausationID() string {
	return e.causationID
}

// Payload returns the event-specific data.
func (e *BaseEvent) Payload() interface{} {
	return e.payload
}
