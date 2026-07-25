// Package interfaces provides interface compliance tests for the kernel contracts.
package interfaces_test

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/kernel/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
	"github.com/veda/agent-runtime/internal/types/runtime"
)

// mockKernel is a minimal implementation of Kernel used to verify the interface compiles.
type mockKernel struct {
	status runtime.RuntimeStatus
}

func (m *mockKernel) Init(_ context.Context) error                             { return nil }
func (m *mockKernel) Start(_ context.Context) error                            { return nil }
func (m *mockKernel) Stop(_ context.Context) error                             { return nil }
func (m *mockKernel) Suspend(_ context.Context) error                          { return nil }
func (m *mockKernel) Resume(_ context.Context) error                           { return nil }
func (m *mockKernel) GetStatus() runtime.RuntimeStatus                         { return m.status }
func (m *mockKernel) RegisterSubsystem(_ string, _ interfaces.Subsystem) error { return nil }
func (m *mockKernel) UnregisterSubsystem(_ string) error                       { return nil }
func (m *mockKernel) GetSubsystem(_ string) (interfaces.Subsystem, error)      { return nil, nil }
func (m *mockKernel) PublishEvent(_ event.Event) error                         { return nil }
func (m *mockKernel) SubscribeToEvent(_ event.Type, _ func(event.Event)) (interfaces.SubscriptionID, error) {
	return 0, nil
}
func (m *mockKernel) UnsubscribeFromEvent(_ interfaces.SubscriptionID) error { return nil }

// mockSubsystem is a minimal implementation of Subsystem.
type mockSubsystem struct {
	initCalled  bool
	startCalled bool
	stopCalled  bool
}

func (m *mockSubsystem) Init(_ context.Context) error {
	m.initCalled = true
	return nil
}

func (m *mockSubsystem) Start(_ context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockSubsystem) Stop(_ context.Context) error {
	m.stopCalled = true
	return nil
}

// mockEventPublisher is a minimal implementation of EventPublisher.
type mockEventPublisher struct {
	published []event.Event
}

func (m *mockEventPublisher) PublishEvent(evt event.Event) error {
	m.published = append(m.published, evt)
	return nil
}

// mockEventSubscriber is a minimal implementation of EventSubscriber.
type mockEventSubscriber struct{}

func (m *mockEventSubscriber) SubscribeToEvent(_ event.Type, _ func(event.Event)) (interfaces.SubscriptionID, error) {
	return 0, nil
}
func (m *mockEventSubscriber) UnsubscribeFromEvent(_ interfaces.SubscriptionID) error { return nil }

// ---------------------------------------------------------------------------
// Interface compliance verification tests
// ---------------------------------------------------------------------------

// TestKernelInterfaceCompliance verifies that the mockKernel type satisfies
// the Kernel interface at compile time. If the interface contract changes,
// this test will fail to compile, providing immediate feedback.
func TestKernelInterfaceCompliance(t *testing.T) {
	var _ interfaces.Kernel = (*mockKernel)(nil)
}

// TestSubsystemInterfaceCompliance verifies that mockSubsystem satisfies Subsystem.
func TestSubsystemInterfaceCompliance(t *testing.T) {
	var _ interfaces.Subsystem = (*mockSubsystem)(nil)
}

// TestEventPublisherInterfaceCompliance verifies that mockEventPublisher satisfies EventPublisher.
func TestEventPublisherInterfaceCompliance(t *testing.T) {
	var _ interfaces.EventPublisher = (*mockEventPublisher)(nil)
}

// TestEventSubscriberInterfaceCompliance verifies that mockEventSubscriber satisfies EventSubscriber.
func TestEventSubscriberInterfaceCompliance(t *testing.T) {
	var _ interfaces.EventSubscriber = (*mockEventSubscriber)(nil)
}

// TestKernelAlsoImplementsEventPublisher verifies that a Kernel implementation
// naturally satisfies EventPublisher (it contains PublishEvent).
func TestKernelAlsoImplementsEventPublisher(t *testing.T) {
	var _ interfaces.EventPublisher = (*mockKernel)(nil)
}

// TestKernelAlsoImplementsEventSubscriber verifies that a Kernel implementation
// naturally satisfies EventSubscriber.
func TestKernelAlsoImplementsEventSubscriber(t *testing.T) {
	var _ interfaces.EventSubscriber = (*mockKernel)(nil)
}

// TestSubsystemLifecycleOrder verifies that Init → Start → Stop is a valid
// call sequence on the Subsystem interface.
func TestSubsystemLifecycleOrder(t *testing.T) {
	ctx := context.Background()
	sub := &mockSubsystem{}

	if err := sub.Init(ctx); err != nil {
		t.Fatalf("Init() returned unexpected error: %v", err)
	}
	if !sub.initCalled {
		t.Error("expected Init to be called")
	}

	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() returned unexpected error: %v", err)
	}
	if !sub.startCalled {
		t.Error("expected Start to be called")
	}

	if err := sub.Stop(ctx); err != nil {
		t.Fatalf("Stop() returned unexpected error: %v", err)
	}
	if !sub.stopCalled {
		t.Error("expected Stop to be called")
	}
}

// TestEventPublisherCollectsEvents verifies that published events are tracked.
func TestEventPublisherCollectsEvents(t *testing.T) {
	publisher := &mockEventPublisher{}
	evt := event.NewBaseEvent("evt-1", event.TypeAgentCreated, "test")

	if err := publisher.PublishEvent(evt); err != nil {
		t.Fatalf("PublishEvent() returned unexpected error: %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.published))
	}
	if publisher.published[0].ID() != "evt-1" {
		t.Errorf("expected event ID %q, got %q", "evt-1", publisher.published[0].ID())
	}
}

// TestKernelGetStatusReturnsInitialStatus verifies RuntimeStatus is plumbed correctly.
func TestKernelGetStatusReturnsInitialStatus(t *testing.T) {
	k := &mockKernel{status: runtime.StatusUninitialized}
	if got := k.GetStatus(); got != runtime.StatusUninitialized {
		t.Errorf("expected status %v, got %v", runtime.StatusUninitialized, got)
	}
}
