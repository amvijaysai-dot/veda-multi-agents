// Package observability provides health checking and diagnostic utilities.
package observability

import (
	"context"
	"sync"
)

// HealthStatus represents the status of a subsystem.
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// HealthComponent represents a single subsystem being monitored.
type HealthComponent struct {
	Name    string
	Status  HealthStatus
	Message string
}

// HealthChecker manages liveness and readiness probes.
type HealthChecker struct {
	mu     sync.RWMutex
	checks map[string]func(context.Context) (HealthStatus, string)
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func(context.Context) (HealthStatus, string)),
	}
}

// RegisterCheck adds a health check function for a specific component.
func (h *HealthChecker) RegisterCheck(name string, checkFn func(context.Context) (HealthStatus, string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = checkFn
}

// CheckLiveness evaluates all registered checks. If any check is Unhealthy,
// the overall system is considered Unhealthy (not alive).
func (h *HealthChecker) CheckLiveness(ctx context.Context) (bool, []HealthComponent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	alive := true
	components := make([]HealthComponent, 0, len(h.checks))

	for name, fn := range h.checks {
		status, msg := fn(ctx)
		components = append(components, HealthComponent{
			Name:    name,
			Status:  status,
			Message: msg,
		})
		if status == StatusUnhealthy {
			alive = false
		}
	}

	return alive, components
}

// CheckReadiness evaluates all registered checks. It demands full Health
// (no degraded or unhealthy) to be considered ready for traffic.
func (h *HealthChecker) CheckReadiness(ctx context.Context) (bool, []HealthComponent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ready := true
	components := make([]HealthComponent, 0, len(h.checks))

	for name, fn := range h.checks {
		status, msg := fn(ctx)
		components = append(components, HealthComponent{
			Name:    name,
			Status:  status,
			Message: msg,
		})
		if status != StatusHealthy {
			ready = false
		}
	}

	return ready, components
}
