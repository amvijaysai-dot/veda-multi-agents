// Package observability provides health checking and diagnostic utilities.
package observability

import (
	"context"
	"testing"
)

func TestHealthChecker_Liveness(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterCheck("db", func(ctx context.Context) (HealthStatus, string) {
		return StatusHealthy, "OK"
	})
	hc.RegisterCheck("cache", func(ctx context.Context) (HealthStatus, string) {
		return StatusDegraded, "Slow"
	})

	alive, components := hc.CheckLiveness(context.Background())
	if !alive {
		t.Error("expected liveness to be true despite degraded cache")
	}
	if len(components) != 2 {
		t.Errorf("expected 2 components, got %d", len(components))
	}
}

func TestHealthChecker_Readiness(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterCheck("db", func(ctx context.Context) (HealthStatus, string) {
		return StatusHealthy, "OK"
	})
	hc.RegisterCheck("cache", func(ctx context.Context) (HealthStatus, string) {
		return StatusDegraded, "Slow"
	})

	ready, _ := hc.CheckReadiness(context.Background())
	if ready {
		t.Error("expected readiness to be false with degraded cache")
	}
}

func TestHealthChecker_Unhealthy(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterCheck("db", func(ctx context.Context) (HealthStatus, string) {
		return StatusUnhealthy, "Dead"
	})

	alive, _ := hc.CheckLiveness(context.Background())
	if alive {
		t.Error("expected liveness to be false when unhealthy")
	}
}

func TestCollectDiagnostics(t *testing.T) {
	snap := CollectDiagnostics()
	if snap.NumGoroutine == 0 {
		t.Error("expected > 0 goroutines")
	}
	if snap.SysBytes == 0 {
		t.Error("expected > 0 sys bytes")
	}
}
