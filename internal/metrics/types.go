// Package metrics provides metrics collection and exposition.
package metrics

// Labels represents a set of key-value dimensions for a metric.
type Labels map[string]string

// Counter represents a metric that only ever goes up.
type Counter interface {
	// Inc increments the counter by 1.
	Inc(labels Labels)

	// Add increments the counter by the given value.
	Add(val float64, labels Labels)
}

// Gauge represents a metric that can go up and down.
type Gauge interface {
	// Set sets the gauge to the given value.
	Set(val float64, labels Labels)

	// Inc increments the gauge by 1.
	Inc(labels Labels)

	// Dec decrements the gauge by 1.
	Dec(labels Labels)

	// Add increments the gauge by the given value.
	Add(val float64, labels Labels)

	// Sub decrements the gauge by the given value.
	Sub(val float64, labels Labels)
}

// Histogram tracks the distribution of a stream of values.
type Histogram interface {
	// Observe adds a single observation to the histogram.
	Observe(val float64, labels Labels)
}

// Summary tracks the distribution of a stream of values (similar to Histogram,
// but calculates quantiles over a sliding time window).
type Summary interface {
	// Observe adds a single observation to the summary.
	Observe(val float64, labels Labels)
}
