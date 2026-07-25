// Package exposition provides formatting for metrics data.
package exposition

import (
	"encoding/json"

	"github.com/veda/agent-runtime/internal/metrics"
)

// JSONExporter implements Exporter for the JSON format.
type JSONExporter struct{}

// NewJSONExporter creates a new JSONExporter.
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

type jsonMetric struct {
	Name   string            `json:"name"`
	Help   string            `json:"help"`
	Type   string            `json:"type"`
	Series []jsonSeriesValue `json:"series"`
}

type jsonSeriesValue struct {
	Labels metrics.Labels `json:"labels"`
	Value  uint64         `json:"value"`
	Count  uint64         `json:"count,omitempty"`
}

func (j *JSONExporter) Format(snapshot []*metrics.MetricData) ([]byte, error) {
	var result []jsonMetric

	for _, md := range snapshot {
		jm := jsonMetric{
			Name:   md.Name,
			Help:   md.Help,
			Series: make([]jsonSeriesValue, 0),
		}

		switch md.Type {
		case metrics.TypeCounter:
			jm.Type = "counter"
		case metrics.TypeGauge:
			jm.Type = "gauge"
		case metrics.TypeHistogram:
			jm.Type = "histogram"
		case metrics.TypeSummary:
			jm.Type = "summary"
		}

		md.Data.Range(func(key, value any) bool {
			sv, ok := value.(*metrics.SeriesValue)
			if !ok {
				return true
			}

			jsv := jsonSeriesValue{
				Labels: sv.Labels,
				Value:  sv.Value,
			}

			if md.Type == metrics.TypeHistogram || md.Type == metrics.TypeSummary {
				jsv.Count = sv.Count
			}

			jm.Series = append(jm.Series, jsv)
			return true
		})

		result = append(result, jm)
	}

	return json.MarshalIndent(result, "", "  ")
}

func (j *JSONExporter) ContentType() string {
	return "application/json"
}
