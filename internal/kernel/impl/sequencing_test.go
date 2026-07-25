// Package impl provides tests for sequencing and shutdown logic.
package impl

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// trackedSubsystem records the order of Init/Start/Stop calls for sequencing tests.
type trackedSubsystem struct {
	name        string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	initErr     error
	startErr    error
	stopErr     error

	// callLog records the sequence of lifecycle method calls (for ordering tests).
	callLog *[]string
}

func newTracked(name string, log *[]string) *trackedSubsystem {
	return &trackedSubsystem{name: name, callLog: log}
}

func (t *trackedSubsystem) Init(_ context.Context) error {
	t.initCalled = true
	if t.callLog != nil {
		*t.callLog = append(*t.callLog, fmt.Sprintf("init:%s", t.name))
	}
	return t.initErr
}

func (t *trackedSubsystem) Start(_ context.Context) error {
	t.startCalled = true
	if t.callLog != nil {
		*t.callLog = append(*t.callLog, fmt.Sprintf("start:%s", t.name))
	}
	return t.startErr
}

func (t *trackedSubsystem) Stop(_ context.Context) error {
	t.stopCalled = true
	if t.callLog != nil {
		*t.callLog = append(*t.callLog, fmt.Sprintf("stop:%s", t.name))
	}
	return t.stopErr
}

// buildRegistry creates a registry with the given tracked subsystems already registered.
func buildRegistry(subs ...*trackedSubsystem) *Registry {
	r := newRegistry()
	for _, s := range subs {
		_ = r.Register(s.name, s)
	}
	return r
}

// ---------------------------------------------------------------------------
// Sequencer.InitAll tests
// ---------------------------------------------------------------------------

func TestSequencer_InitAll_Success(t *testing.T) {
	log := []string{}
	a := newTracked("a", &log)
	b := newTracked("b", &log)
	c := newTracked("c", &log)

	r := buildRegistry(a, b, c)
	seq := newSequencer(r)

	if err := seq.InitAll(context.Background()); err != nil {
		t.Fatalf("InitAll() unexpected error: %v", err)
	}

	// Verify all subsystems were initialized.
	if !a.initCalled || !b.initCalled || !c.initCalled {
		t.Error("expected all subsystems to be initialized")
	}

	// Verify initialization order (a → b → c).
	want := []string{"init:a", "init:b", "init:c"}
	if len(log) != len(want) {
		t.Fatalf("expected %d calls, got %d: %v", len(want), len(log), log)
	}
	for i, w := range want {
		if log[i] != w {
			t.Errorf("call[%d] = %q, want %q", i, log[i], w)
		}
	}
}

func TestSequencer_InitAll_FailureRollsBackPreviousSubsystems(t *testing.T) {
	log := []string{}
	a := newTracked("a", &log)
	b := newTracked("b", &log)
	b.initErr = errors.New("b init failed")
	c := newTracked("c", &log) // should never be reached

	r := buildRegistry(a, b, c)
	seq := newSequencer(r)

	err := seq.InitAll(context.Background())
	if err == nil {
		t.Fatal("expected error from InitAll, got nil")
	}

	// a was initialized before b failed; it must be rolled back (stopped).
	if !a.stopCalled {
		t.Error("expected previously-initialized subsystem 'a' to be stopped during rollback")
	}

	// c should never have been initialized.
	if c.initCalled {
		t.Error("subsystem 'c' should not have been initialized after 'b' failed")
	}
}

func TestSequencer_InitAll_EmptyRegistrySucceeds(t *testing.T) {
	r := newRegistry()
	seq := newSequencer(r)

	if err := seq.InitAll(context.Background()); err != nil {
		t.Fatalf("InitAll() on empty registry returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Sequencer.StartAll tests
// ---------------------------------------------------------------------------

func TestSequencer_StartAll_Success(t *testing.T) {
	log := []string{}
	a := newTracked("a", &log)
	b := newTracked("b", &log)

	r := buildRegistry(a, b)
	seq := newSequencer(r)

	if err := seq.StartAll(context.Background()); err != nil {
		t.Fatalf("StartAll() unexpected error: %v", err)
	}

	want := []string{"start:a", "start:b"}
	for i, w := range want {
		if log[i] != w {
			t.Errorf("call[%d] = %q, want %q", i, log[i], w)
		}
	}
}

func TestSequencer_StartAll_FailureRollsBack(t *testing.T) {
	log := []string{}
	a := newTracked("a", &log)
	b := newTracked("b", &log)
	b.startErr = errors.New("b start failed")

	r := buildRegistry(a, b)
	seq := newSequencer(r)

	err := seq.StartAll(context.Background())
	if err == nil {
		t.Fatal("expected error from StartAll, got nil")
	}

	// 'a' was started before 'b' failed; it must be rolled back.
	if !a.stopCalled {
		t.Error("expected 'a' to be stopped during rollback")
	}
}

// ---------------------------------------------------------------------------
// Sequencer.StopAll tests
// ---------------------------------------------------------------------------

func TestSequencer_StopAll_StopsInReverseOrder(t *testing.T) {
	log := []string{}
	a := newTracked("a", &log)
	b := newTracked("b", &log)
	c := newTracked("c", &log)

	r := buildRegistry(a, b, c)
	seq := newSequencer(r)

	if err := seq.StopAll(context.Background()); err != nil {
		t.Fatalf("StopAll() unexpected error: %v", err)
	}

	// Reverse of registration order: c → b → a
	want := []string{"stop:c", "stop:b", "stop:a"}
	if len(log) != len(want) {
		t.Fatalf("expected %d calls, got %d: %v", len(want), len(log), log)
	}
	for i, w := range want {
		if log[i] != w {
			t.Errorf("call[%d] = %q, want %q", i, log[i], w)
		}
	}
}

func TestSequencer_StopAll_CollectsAllErrors(t *testing.T) {
	a := newTracked("a", nil)
	b := newTracked("b", nil)
	c := newTracked("c", nil)
	a.stopErr = errors.New("a stop failed")
	c.stopErr = errors.New("c stop failed")

	r := buildRegistry(a, b, c)
	seq := newSequencer(r)

	err := seq.StopAll(context.Background())
	if err == nil {
		t.Fatal("expected combined stop error, got nil")
	}

	// All subsystems should still have been stopped.
	if !a.stopCalled {
		t.Error("expected 'a' Stop to be called even though it fails")
	}
	if !b.stopCalled {
		t.Error("expected 'b' Stop to be called")
	}
	if !c.stopCalled {
		t.Error("expected 'c' Stop to be called even though it fails")
	}
}

func TestSequencer_StopAll_EmptyRegistrySucceeds(t *testing.T) {
	r := newRegistry()
	seq := newSequencer(r)

	if err := seq.StopAll(context.Background()); err != nil {
		t.Fatalf("StopAll() on empty registry returned unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DoubleInitGuard tests
// ---------------------------------------------------------------------------

func TestDoubleInitGuard_FirstCheckSucceeds(t *testing.T) {
	g := newDoubleInitGuard()
	if err := g.Check("sub-a"); err != nil {
		t.Fatalf("Check() first call returned unexpected error: %v", err)
	}
}

func TestDoubleInitGuard_SecondCheckFails(t *testing.T) {
	g := newDoubleInitGuard()
	_ = g.Check("sub-a")

	if err := g.Check("sub-a"); err == nil {
		t.Fatal("expected error on second Check() for same name, got nil")
	}
}

func TestDoubleInitGuard_ResetAllowsReinit(t *testing.T) {
	g := newDoubleInitGuard()
	_ = g.Check("sub-a")
	g.Reset("sub-a")

	if err := g.Check("sub-a"); err != nil {
		t.Fatalf("Check() after Reset() returned unexpected error: %v", err)
	}
}

func TestDoubleInitGuard_IsInitialized(t *testing.T) {
	g := newDoubleInitGuard()
	if g.IsInitialized("sub-a") {
		t.Error("expected IsInitialized() to be false before Check()")
	}
	_ = g.Check("sub-a")
	if !g.IsInitialized("sub-a") {
		t.Error("expected IsInitialized() to be true after Check()")
	}
}

// ---------------------------------------------------------------------------
// Kernel lifecycle integration tests
// (uses the full Kernel → Sequencer pipeline)
// ---------------------------------------------------------------------------

func TestKernel_FullLifecycle_InitStartStop(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()

	a := newTracked("a", nil)
	b := newTracked("b", nil)

	_ = k.RegisterSubsystem("a", a)
	_ = k.RegisterSubsystem("b", b)

	if err := k.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if err := k.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if err := k.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if !a.initCalled || !b.initCalled {
		t.Error("expected both subsystems to be initialized")
	}
	if !a.startCalled || !b.startCalled {
		t.Error("expected both subsystems to be started")
	}
	if !a.stopCalled || !b.stopCalled {
		t.Error("expected both subsystems to be stopped")
	}
}

func TestKernel_StopIsIdempotent(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()

	if err := k.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if err := k.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// First Stop should succeed.
	if err := k.Stop(ctx); err != nil {
		t.Fatalf("first Stop() failed: %v", err)
	}
	// Second Stop should be a no-op, not an error.
	if err := k.Stop(ctx); err != nil {
		t.Fatalf("second Stop() (idempotent) failed: %v", err)
	}
}

func TestKernel_SuspendResume(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()

	if err := k.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if err := k.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	if err := k.Suspend(ctx); err != nil {
		t.Fatalf("Suspend() failed: %v", err)
	}
	if err := k.Resume(ctx); err != nil {
		t.Fatalf("Resume() failed: %v", err)
	}
	if err := k.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}

func TestKernel_InitFailureRollsBack(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()

	good := newTracked("good", nil)
	bad := newTracked("bad", nil)
	bad.initErr = errors.New("deliberate init failure")

	_ = k.RegisterSubsystem("good", good)
	_ = k.RegisterSubsystem("bad", bad)

	err := k.Init(ctx)
	if err == nil {
		t.Fatal("expected error from Init(), got nil")
	}

	// The good subsystem should have been rolled back (stopped).
	if !good.stopCalled {
		t.Error("expected 'good' subsystem to be stopped during rollback")
	}
}

func TestKernel_RegisterAfterInitReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()

	if err := k.Init(ctx); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	err := k.RegisterSubsystem("late", newNoop("late"))
	if err == nil {
		t.Fatal("expected error when registering subsystem after Init(), got nil")
	}
}

func TestKernel_GetSubsystem_Success(t *testing.T) {
	k := NewKernel()
	sub := newNoop("alpha")
	_ = k.RegisterSubsystem("alpha", sub)

	got, err := k.GetSubsystem("alpha")
	if err != nil {
		t.Fatalf("GetSubsystem() unexpected error: %v", err)
	}
	if got != sub {
		t.Error("GetSubsystem() returned different subsystem than what was registered")
	}
}

func TestKernel_GetSubsystem_UnknownReturnsError(t *testing.T) {
	k := NewKernel()
	_, err := k.GetSubsystem("unknown")
	if err == nil {
		t.Fatal("expected error for unknown subsystem, got nil")
	}
}

func TestKernel_UnregisterSubsystem(t *testing.T) {
	k := NewKernel()
	_ = k.RegisterSubsystem("alpha", newNoop("alpha"))

	if err := k.UnregisterSubsystem("alpha"); err != nil {
		t.Fatalf("UnregisterSubsystem() unexpected error: %v", err)
	}

	_, err := k.GetSubsystem("alpha")
	if err == nil {
		t.Fatal("expected error after unregistering 'alpha', got nil")
	}
}
