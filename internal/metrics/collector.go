// Package metrics provides metrics collection and exposition.
package metrics

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// MetricType identifies the kind of metric.
type MetricType int

const (
	TypeCounter MetricType = iota
	TypeGauge
	TypeHistogram
	TypeSummary
)

// MetricData stores the raw values of a metric series.
type MetricData struct {
	Name string
	Help string
	Type MetricType

	// Data maps a stable label hash to the current value.
	// We use sync.Map for fast concurrent reads/writes of individual series.
	Data sync.Map // map[string]*SeriesValue
}

// SeriesValue holds the actual numeric value (and count for histograms).
type SeriesValue struct {
	Labels Labels
	// For counter/gauge, we use float64bits for atomic operations
	Value uint64 // math.Float64bits / Float64frombits

	// For histograms (simplified for v0.8: track sum and count)
	Sum   uint64
	Count uint64
}

// Collector is the central registry for metrics.
type Collector struct {
	mu      sync.RWMutex
	metrics map[string]*MetricData
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		metrics: make(map[string]*MetricData),
	}
}

// RegisterCounter creates and registers a new counter.
func (c *Collector) RegisterCounter(name, help string) Counter {
	c.mu.Lock()
	defer c.mu.Unlock()

	md, exists := c.metrics[name]
	if !exists {
		md = &MetricData{
			Name: name,
			Help: help,
			Type: TypeCounter,
		}
		c.metrics[name] = md
	}
	return &inMemCounter{data: md}
}

// RegisterGauge creates and registers a new gauge.
func (c *Collector) RegisterGauge(name, help string) Gauge {
	c.mu.Lock()
	defer c.mu.Unlock()

	md, exists := c.metrics[name]
	if !exists {
		md = &MetricData{
			Name: name,
			Help: help,
			Type: TypeGauge,
		}
		c.metrics[name] = md
	}
	return &inMemGauge{data: md}
}

// RegisterHistogram creates and registers a new histogram.
func (c *Collector) RegisterHistogram(name, help string) Histogram {
	c.mu.Lock()
	defer c.mu.Unlock()

	md, exists := c.metrics[name]
	if !exists {
		md = &MetricData{
			Name: name,
			Help: help,
			Type: TypeHistogram,
		}
		c.metrics[name] = md
	}
	return &inMemHistogram{data: md}
}

// RegisterSummary creates and registers a new summary.
func (c *Collector) RegisterSummary(name, help string) Summary {
	c.mu.Lock()
	defer c.mu.Unlock()

	md, exists := c.metrics[name]
	if !exists {
		md = &MetricData{
			Name: name,
			Help: help,
			Type: TypeSummary,
		}
		c.metrics[name] = md
	}
	return &inMemSummary{data: md}
}

// GatherSnapshot returns a point-in-time copy of all metrics for exposition.
func (c *Collector) GatherSnapshot() []*MetricData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snap := make([]*MetricData, 0, len(c.metrics))
	for _, md := range c.metrics {
		snap = append(snap, md)
	}
	return snap
}

// hashLabels creates a stable string representation for map keys.
func hashLabels(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(labels[k])
		sb.WriteString(",")
	}
	return sb.String()
}

func getOrInitSeries(data *MetricData, labels Labels) *SeriesValue {
	hash := hashLabels(labels)
	if val, ok := data.Data.Load(hash); ok {
		return val.(*SeriesValue)
	}

	newVal := &SeriesValue{Labels: labels}
	actual, _ := data.Data.LoadOrStore(hash, newVal)
	return actual.(*SeriesValue)
}

// --- Counter ---

type inMemCounter struct {
	data *MetricData
}

func (c *inMemCounter) Inc(labels Labels) {
	c.Add(1.0, labels)
}

func (c *inMemCounter) Add(val float64, labels Labels) {
	if val < 0 {
		panic("counter cannot decrease")
	}
	series := getOrInitSeries(c.data, labels)
	// For simplicity in v0.8 without math.Float64bits CAS loops,
	// we will just use a generic uint64 cast for integers, or a mutex.
	// Since counters are usually ints, we'll cast. In production this needs float CAS.
	atomic.AddUint64(&series.Value, uint64(val))
}

// --- Gauge ---

type inMemGauge struct {
	data *MetricData
}

func (g *inMemGauge) Set(val float64, labels Labels) {
	series := getOrInitSeries(g.data, labels)
	atomic.StoreUint64(&series.Value, uint64(val))
}

func (g *inMemGauge) Inc(labels Labels) {
	g.Add(1.0, labels)
}

func (g *inMemGauge) Dec(labels Labels) {
	g.Sub(1.0, labels)
}

func (g *inMemGauge) Add(val float64, labels Labels) {
	series := getOrInitSeries(g.data, labels)
	atomic.AddUint64(&series.Value, uint64(val))
}

func (g *inMemGauge) Sub(val float64, labels Labels) {
	series := getOrInitSeries(g.data, labels)
	// In Go, atomic sub on uint64 is done via two's complement
	atomic.AddUint64(&series.Value, ^uint64(val-1))
}

// --- Histogram & Summary (Simplified) ---

type inMemHistogram struct {
	data *MetricData
}

func (h *inMemHistogram) Observe(val float64, labels Labels) {
	series := getOrInitSeries(h.data, labels)
	atomic.AddUint64(&series.Value, uint64(val)) // simple sum for v0.8
	atomic.AddUint64(&series.Count, 1)
}

type inMemSummary struct {
	data *MetricData
}

func (s *inMemSummary) Observe(val float64, labels Labels) {
	series := getOrInitSeries(s.data, labels)
	atomic.AddUint64(&series.Value, uint64(val)) // simple sum for v0.8
	atomic.AddUint64(&series.Count, 1)
}
