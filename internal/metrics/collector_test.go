// Package metrics provides metrics collection and exposition.
package metrics

import (
	"sync"
	"testing"
)

func TestCollector_Counter(t *testing.T) {
	c := NewCollector()
	counter := c.RegisterCounter("test_req", "Test requests")

	lbls := Labels{"method": "GET"}
	counter.Inc(lbls)
	counter.Add(5, lbls)

	snap := c.GatherSnapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(snap))
	}

	val, ok := snap[0].Data.Load(hashLabels(lbls))
	if !ok {
		t.Fatal("expected to find series")
	}

	sv := val.(*SeriesValue)
	if sv.Value != 6 {
		t.Errorf("expected value 6, got %d", sv.Value)
	}
}

func TestCollector_Gauge(t *testing.T) {
	c := NewCollector()
	gauge := c.RegisterGauge("test_mem", "Test memory")

	lbls := Labels{"node": "1"}
	gauge.Set(100, lbls)
	gauge.Inc(lbls)
	gauge.Sub(50, lbls)

	snap := c.GatherSnapshot()
	val, _ := snap[0].Data.Load(hashLabels(lbls))
	sv := val.(*SeriesValue)

	if sv.Value != 51 {
		t.Errorf("expected value 51, got %d", sv.Value)
	}
}

func TestCollector_Concurrency(t *testing.T) {
	c := NewCollector()
	counter := c.RegisterCounter("concurrent_requests", "concurrent test")

	var wg sync.WaitGroup
	lbls := Labels{"route": "/api"}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc(lbls)
		}()
	}

	wg.Wait()

	snap := c.GatherSnapshot()
	val, _ := snap[0].Data.Load(hashLabels(lbls))
	sv := val.(*SeriesValue)

	if sv.Value != 1000 {
		t.Errorf("expected value 1000, got %d", sv.Value)
	}
}
