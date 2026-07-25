// Package interfaces defines the contracts for all execution engine components.
package interfaces

// ErrorClassification indicates how the ErrorHandler has categorised an error.
type ErrorClassification int

const (
	// Recoverable errors can be retried after a backoff delay.
	Recoverable ErrorClassification = iota

	// Fatal errors cannot be retried and must abort the current turn.
	Fatal
)

// String returns the human-readable label for the classification.
func (e ErrorClassification) String() string {
	switch e {
	case Recoverable:
		return "recoverable"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// ErrorHandler classifies execution errors, drives retry logic with exponential
// backoff, and maintains a circuit-breaker to prevent cascading failures.
type ErrorHandler interface {
	// Classify determines whether err is Recoverable or Fatal.
	Classify(err error) ErrorClassification

	// ShouldRetry returns true if the operation that produced err should be
	// retried, taking the current attempt number and circuit-breaker state into
	// account.
	ShouldRetry(err error, attempt int) bool

	// RecordFailure records a failure event for circuit-breaker tracking.
	// Must be called every time an operation fails, regardless of retry decision.
	RecordFailure()

	// RecordSuccess records a success event, resetting transient failure counters.
	RecordSuccess()

	// IsCircuitOpen returns true when the circuit breaker has tripped and no
	// further retries should be attempted until the circuit resets.
	IsCircuitOpen() bool
}
