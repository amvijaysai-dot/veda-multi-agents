// Package exposition provides formatting for metrics data.
package exposition

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/veda/agent-runtime/internal/metrics"
)

// PrometheusExporter implements Exporter for the Prometheus text format.
type PrometheusExporter struct{}

// NewPrometheusExporter creates a new PrometheusExporter.
func NewPrometheusExporter() *PrometheusExporter {
	return &PrometheusExporter{}
}

func (p *PrometheusExporter) Format(snapshot []*metrics.MetricData) ([]byte, error) {
	var buf bytes.Buffer

	// Sort metrics by name for consistent output
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].Name < snapshot[j].Name
	})

	for _, md := range snapshot {
		if md.Help != "" {
			fmt.Fprintf(&buf, "# HELP %s %s\n", md.Name, md.Help)
		}

		typeStr := "unknown"
		switch md.Type {
		case metrics.TypeCounter:
			typeStr = "counter"
		case metrics.TypeGauge:
			typeStr = "gauge"
		case metrics.TypeHistogram:
			typeStr = "histogram"
		case metrics.TypeSummary:
			typeStr = "summary"
		}

		fmt.Fprintf(&buf, "# TYPE %s %s\n", md.Name, typeStr)

		// We need to iterate over the sync.Map safely
		md.Data.Range(func(key, value any) bool {
			sv, ok := value.(*metrics.SeriesValue)
			if !ok {
				return true
			}

			// Format labels
			var labelsStr string
			if len(sv.Labels) > 0 {
				var lbls []string
				for k, v := range sv.Labels {
					lbls = append(lbls, fmt.Sprintf(`%s="%s"`, k, v))
				}
				sort.Strings(lbls)
				labelsStr = "{" + strings.Join(lbls, ",") + "}"
			}

			// Print the series
			if md.Type == metrics.TypeHistogram || md.Type == metrics.TypeSummary {
				fmt.Fprintf(&buf, "%s_sum%s %d\n", md.Name, labelsStr, sv.Value)
				fmt.Fprintf(&buf, "%s_count%s %d\n", md.Name, labelsStr, sv.Count)
			} else {
				fmt.Fprintf(&buf, "%s%s %d\n", md.Name, labelsStr, sv.Value)
			}
			return true
		})
	}

	return buf.Bytes(), nil
}

func (p *PrometheusExporter) ContentType() string {
	return "text/plain; version=0.0.4"
}
