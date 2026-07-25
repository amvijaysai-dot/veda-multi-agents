// Package interfaces defines the contracts for the memory subsystem.
package interfaces

import (
	"context"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Interface compliance verification via compile-time type assertions.
// ---------------------------------------------------------------------------

type mockShortTerm struct{}

func (m *mockShortTerm) Store(_ context.Context, _, _, _, _ string) error { return nil }
func (m *mockShortTerm) Retrieve(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}
func (m *mockShortTerm) Delete(_ context.Context, _, _, _ string) error { return nil }
func (m *mockShortTerm) List(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockShortTerm) Clear(_ context.Context, _, _ string) error              { return nil }
func (m *mockShortTerm) PersistenceHint(_ context.Context, _, _, _ string) error { return nil }

type mockLongTerm struct{}

func (m *mockLongTerm) Store(_ context.Context, _, _, _ string, _ time.Duration) error { return nil }
func (m *mockLongTerm) Retrieve(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *mockLongTerm) Delete(_ context.Context, _, _ string) error { return nil }
func (m *mockLongTerm) Query(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLongTerm) Scan(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLongTerm) Forget(_ context.Context, _, _, _ string) error { return nil }

type mockConsolidation struct{}

func (m *mockConsolidation) Consolidate(_ context.Context, _, _ string) error { return nil }

type mockSharing struct{}

func (m *mockSharing) Share(_ context.Context, _, _, _ string) error { return nil }

type mockPrivacy struct{}

func (m *mockPrivacy) Scrub(_ context.Context, _ string) (string, error) { return "", nil }

// Compile-time assertions
var (
	_ ShortTermMemory      = (*mockShortTerm)(nil)
	_ LongTermMemory       = (*mockLongTerm)(nil)
	_ ConsolidationManager = (*mockConsolidation)(nil)
	_ SharingManager       = (*mockSharing)(nil)
	_ PrivacyManager       = (*mockPrivacy)(nil)
)

// Runtime tests: verify interface method behaviours on mock stubs.
func TestShortTermMemory_Compliance(t *testing.T) {
	m := &mockShortTerm{}
	ctx := context.Background()
	_ = m.Store(ctx, "a", "s", "k", "v")
	_, _ = m.Retrieve(ctx, "a", "s", "k")
	_ = m.Delete(ctx, "a", "s", "k")
	_, _ = m.List(ctx, "a", "s", "prefix")
	_ = m.Clear(ctx, "a", "s")
	_ = m.PersistenceHint(ctx, "a", "s", "k")
}

func TestLongTermMemory_Compliance(t *testing.T) {
	m := &mockLongTerm{}
	ctx := context.Background()
	_ = m.Store(ctx, "a", "k", "v", 0)
	_, _ = m.Retrieve(ctx, "a", "k")
	_ = m.Delete(ctx, "a", "k")
	_, _ = m.Query(ctx, "a", "q")
	_, _ = m.Scan(ctx, "a", "prefix")
	_ = m.Forget(ctx, "a", "k", "reason")
}

func TestConsolidationManager_Compliance(t *testing.T) {
	m := &mockConsolidation{}
	_ = m.Consolidate(context.Background(), "a", "s")
}

func TestSharingManager_Compliance(t *testing.T) {
	m := &mockSharing{}
	_ = m.Share(context.Background(), "a", "k", "target")
}

func TestPrivacyManager_Compliance(t *testing.T) {
	m := &mockPrivacy{}
	_, _ = m.Scrub(context.Background(), "text")
}
