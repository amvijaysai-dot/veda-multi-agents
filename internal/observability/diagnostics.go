// Package observability provides health checking and diagnostic utilities.
package observability

import (
	"runtime"
)

// DiagnosticsSnapshot holds a point-in-time view of system metrics.
type DiagnosticsSnapshot struct {
	NumGoroutine int
	AllocBytes   uint64
	TotalAlloc   uint64
	SysBytes     uint64
	NumGC        uint32
}

// CollectDiagnostics captures standard Go runtime diagnostics.
func CollectDiagnostics() DiagnosticsSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return DiagnosticsSnapshot{
		NumGoroutine: runtime.NumGoroutine(),
		AllocBytes:   memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		SysBytes:     memStats.Sys,
		NumGC:        memStats.NumGC,
	}
}
