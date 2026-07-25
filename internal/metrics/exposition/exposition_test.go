// Package exposition provides formatting for metrics data.
package exposition

import (
	"strings"
	"testing"

	"github.com/veda/agent-runtime/internal/metrics"
)

func TestPrometheusExporter(t *testing.T) {
	c := metrics.NewCollector()
	c.RegisterCounter("http_requests_total", "Total HTTP requests").Inc(metrics.Labels{"method": "GET"})

	snap := c.GatherSnapshot()

	exporter := NewPrometheusExporter()
	out, err := exporter.Format(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res := string(out)
	if !strings.Contains(res, "# HELP http_requests_total Total HTTP requests") {
		t.Error("missing help")
	}
	if !strings.Contains(res, "# TYPE http_requests_total counter") {
		t.Error("missing type")
	}
	if !strings.Contains(res, `http_requests_total{method="GET"} 1`) {
		t.Errorf("missing series, got: %s", res)
	}
}

func TestJSONExporter(t *testing.T) {
	c := metrics.NewCollector()
	c.RegisterHistogram("db_duration", "DB query duration").Observe(150, metrics.Labels{"query": "select"})

	snap := c.GatherSnapshot()

	exporter := NewJSONExporter()
	out, err := exporter.Format(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res := string(out)
	if !strings.Contains(res, `"name": "db_duration"`) {
		t.Error("missing name")
	}
	if !strings.Contains(res, `"type": "histogram"`) {
		t.Error("missing type")
	}
	if !strings.Contains(res, `"value": 150`) {
		t.Error("missing sum/value")
	}
	if !strings.Contains(res, `"count": 1`) {
		t.Error("missing count")
	}
}
