# ADR-0001: Event Subscriber Interface Design

## Status
Accepted

## Context
The Milestone v0.2 implementation backlog (V0.2.01) specified that `EventSubscriber` should define a `HandleEvent` method. This implied a push-style callback interface where subscribers would implement the interface, and the event bus would call `HandleEvent` on the subscriber object for all events. 
However, the actual implementation provided `SubscribeToEvent(eventType, handlerFunc)` and `UnsubscribeFromEvent(subscriptionID)`. This deviated from the frozen backlog specification without a formal Architecture Decision Record (ADR).

## Decision
We formally accept the registration-style interface (`SubscribeToEvent` / `UnsubscribeFromEvent`) as the standard for the `EventSubscriber` contract, rather than the handler-style interface (`HandleEvent`).

The `EventSubscriber` interface is defined as:
```go
type EventSubscriber interface {
	SubscribeToEvent(eventType event.Type, handler func(event.Event)) (SubscriptionID, error)
	UnsubscribeFromEvent(id SubscriptionID) error
}
```

## Consequences
### Positive consequences (+)
- Enables type-filtered subscriptions directly at the event bus level, avoiding the need for every subscriber to manually filter out irrelevant events.
- Allows a single component (e.g., a subsystem) to register multiple different, focused handler functions for different event types, rather than implementing a monolithic `HandleEvent` switch-statement.
- Avoids forcing subscribers to implement a specific interface just to receive events, providing more flexibility in how they are implemented.
- Follows the Interface Segregation Principle by decoupling the subscription mechanism from the event processing logic.
- Using an opaque `SubscriptionID` for unsubscription avoids the unreliability of Go function pointer comparisons.

### Negative consequences (-)
- Deviation from the initial architecture backlog required a retroactive ADR and update.
- Subscribers must maintain state (the `SubscriptionID`) to unsubscribe later, slightly increasing the complexity on the subscriber side compared to simply unregistering an object.

## Related Decisions
- None

## Notes
The `SubscriptionID` token pattern was implemented concurrently to resolve unreliability issues with `fmt.Sprintf("%p")` function pointer comparisons in Go.
