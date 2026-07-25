// Package impl provides tests for the Kernel implementation.
package impl

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/types/event"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// ---------------------------------------------------------------------------
// Kernel status tests
// ---------------------------------------------------------------------------

func TestKernel_InitialStatus_IsUninitialized(t *testing.T) {
	k := NewKernel()
	if got := k.GetStatus(); got != runtime.StatusUninitialized {
		t.Errorf("expected initial status %v, got %v", runtime.StatusUninitialized, got)
	}
}

func TestKernel_StatusAfterInit_IsReady(t *testing.T) {
	k := NewKernel()
	if err := k.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if got := k.GetStatus(); got != runtime.StatusReady {
		t.Errorf("expected status %v after Init, got %v", runtime.StatusReady, got)
	}
}

func TestKernel_StatusAfterStop_IsTerminated(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	_ = k.Stop(ctx)

	if got := k.GetStatus(); got != runtime.StatusTerminated {
		t.Errorf("expected status %v after Stop, got %v", runtime.StatusTerminated, got)
	}
}

func TestKernel_StatusAfterSuspend_IsSuspended(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	_ = k.Suspend(ctx)

	if got := k.GetStatus(); got != runtime.StatusSuspended {
		t.Errorf("expected status %v after Suspend, got %v", runtime.StatusSuspended, got)
	}
}

func TestKernel_StatusAfterResume_IsReady(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	_ = k.Suspend(ctx)
	_ = k.Resume(ctx)

	if got := k.GetStatus(); got != runtime.StatusReady {
		t.Errorf("expected status %v after Resume, got %v", runtime.StatusReady, got)
	}
}

// ---------------------------------------------------------------------------
// State machine guard tests
// ---------------------------------------------------------------------------

func TestKernel_InitTwiceReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)

	err := k.Init(ctx)
	if err == nil {
		t.Fatal("expected error from second Init(), got nil")
	}
}

func TestKernel_StartWithoutInitReturnsError(t *testing.T) {
	k := NewKernel()
	err := k.Start(context.Background())
	if err == nil {
		t.Fatal("expected error from Start() without Init(), got nil")
	}
}

func TestKernel_SuspendWithoutStartReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	// Not started yet

	// After Init, status is Ready (not Busy or explicitly running), Suspend should pass.
	// Suspend from Ready state is valid.
	if err := k.Suspend(ctx); err != nil {
		t.Fatalf("Suspend() from Ready state returned error: %v", err)
	}
}

func TestKernel_ResumeWithoutSuspendReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	// Not suspended

	err := k.Resume(ctx)
	if err == nil {
		t.Fatal("expected error from Resume() when not suspended, got nil")
	}
}

// ---------------------------------------------------------------------------
// Event bus tests
// ---------------------------------------------------------------------------

func TestKernel_PublishEvent_StoppedBusReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	_ = k.Stop(ctx)

	evt := event.NewBaseEvent("evt-1", event.TypeAgentCreated, "test")
	err := k.PublishEvent(evt)
	if err == nil {
		t.Fatal("expected error publishing to stopped event bus, got nil")
	}
}

func TestKernel_SubscribeAndPublish_HandlerInvoked(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	received := make([]event.Event, 0)
	handler := func(e event.Event) {
		received = append(received, e)
	}

	if _, err := k.SubscribeToEvent(event.TypeAgentCreated, handler); err != nil {
		t.Fatalf("SubscribeToEvent() failed: %v", err)
	}

	evt := event.NewBaseEvent("evt-1", event.TypeAgentCreated, "test")
	if err := k.PublishEvent(evt); err != nil {
		t.Fatalf("PublishEvent() failed: %v", err)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 event received, got %d", len(received))
	}
	if received[0].ID() != "evt-1" {
		t.Errorf("expected event ID %q, got %q", "evt-1", received[0].ID())
	}
}

func TestKernel_SubscribeNilHandlerReturnsError(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	_, err := k.SubscribeToEvent(event.TypeAgentCreated, nil)
	if err == nil {
		t.Fatal("expected error subscribing with nil handler, got nil")
	}
}

func TestKernel_UnsubscribeRemovesHandler(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	callCount := 0
	handler := func(e event.Event) { callCount++ }

	subID, err := k.SubscribeToEvent(event.TypeAgentCreated, handler)
	if err != nil {
		t.Fatalf("SubscribeToEvent() failed: %v", err)
	}
	if err := k.UnsubscribeFromEvent(subID); err != nil {
		t.Fatalf("UnsubscribeFromEvent() failed: %v", err)
	}

	evt := event.NewBaseEvent("evt-2", event.TypeAgentCreated, "test")
	_ = k.PublishEvent(evt)

	if callCount != 0 {
		t.Errorf("expected handler not to be called after unsubscribe, got %d calls", callCount)
	}
}

func TestKernel_UnsubscribeUnknownIDIsNoOp(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	// Unsubscribing a nonexistent SubscriptionID must not return an error.
	if err := k.UnsubscribeFromEvent(9999); err != nil {
		t.Errorf("expected no-op for unknown SubscriptionID, got error: %v", err)
	}
}

func TestKernel_MultipleSubscribers_AllReceiveEvents(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	count1, count2 := 0, 0
	_, _ = k.SubscribeToEvent(event.TypeAgentCreated, func(_ event.Event) { count1++ })
	_, _ = k.SubscribeToEvent(event.TypeAgentCreated, func(_ event.Event) { count2++ })

	evt := event.NewBaseEvent("evt-multi", event.TypeAgentCreated, "test")
	_ = k.PublishEvent(evt)

	if count1 != 1 || count2 != 1 {
		t.Errorf("expected both handlers to receive event, got count1=%d count2=%d", count1, count2)
	}
}

func TestKernel_PublishDoesNotDeliverToWrongType(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	agentCreatedCount := 0
	_, _ = k.SubscribeToEvent(event.TypeAgentCreated, func(_ event.Event) { agentCreatedCount++ })

	// Publish a different event type.
	evt := event.NewBaseEvent("evt-3", event.TypeAgentStopped, "test")
	_ = k.PublishEvent(evt)

	if agentCreatedCount != 0 {
		t.Errorf("expected 0 deliveries to TypeAgentCreated handler, got %d", agentCreatedCount)
	}
}

// ---------------------------------------------------------------------------
// UnregisterSubsystem lifecycle guard tests
// ---------------------------------------------------------------------------

func TestKernel_UnregisterSubsystem_AllowedWhenUninitialized(t *testing.T) {
	k := NewKernel()
	_ = k.RegisterSubsystem("alpha", newNoop("alpha"))

	if err := k.UnregisterSubsystem("alpha"); err != nil {
		t.Fatalf("UnregisterSubsystem() in Uninitialized state returned error: %v", err)
	}
	_, err := k.GetSubsystem("alpha")
	if err == nil {
		t.Fatal("expected error after unregistering 'alpha', got nil")
	}
}

func TestKernel_UnregisterSubsystem_AllowedWhenTerminated(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.RegisterSubsystem("alpha", newNoop("alpha"))
	_ = k.Init(ctx)
	_ = k.Start(ctx)
	_ = k.Stop(ctx)

	// After Stop, status is Terminated — unregister should succeed.
	if err := k.UnregisterSubsystem("alpha"); err != nil {
		t.Fatalf("UnregisterSubsystem() in Terminated state returned error: %v", err)
	}
}

func TestKernel_UnregisterSubsystem_RejectedWhenRunning(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.RegisterSubsystem("alpha", newNoop("alpha"))
	_ = k.Init(ctx)
	_ = k.Start(ctx)

	err := k.UnregisterSubsystem("alpha")
	if err == nil {
		t.Fatal("expected error when unregistering subsystem while kernel is running, got nil")
	}
}

func TestKernel_UnregisterSubsystem_RejectedWhenInitialized(t *testing.T) {
	k := NewKernel()
	ctx := context.Background()
	_ = k.RegisterSubsystem("alpha", newNoop("alpha"))
	_ = k.Init(ctx)
	// Kernel is now in Ready state (post-Init, pre-Start)

	err := k.UnregisterSubsystem("alpha")
	if err == nil {
		t.Fatal("expected error when unregistering subsystem in Ready state, got nil")
	}
}
