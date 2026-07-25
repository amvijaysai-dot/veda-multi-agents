// Package exposition provides formatting for metrics data.
package exposition

import (
	"github.com/veda/agent-runtime/internal/metrics"
)

// Exporter defines the interface for exporting a metrics snapshot.
type Exporter interface {
	// Format takes a snapshot of metrics and returns the formatted representation.
	Format(snapshot []*metrics.MetricData) ([]byte, error)

	// ContentType returns the MIME type of the formatted data.
	ContentType() string
}
