// Package consolidation provides mechanisms to move data from short-term memory to long-term memory.
package consolidation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockCandidateProvider struct {
	candidates map[string]string
}

func (m *mockCandidateProvider) GetConsolidationCandidates(_, _ string) map[string]string {
	return m.candidates
}

type mockLongTerm struct {
	stored map[string]string
	err    error
}

func (m *mockLongTerm) Store(_ context.Context, agentID, key, value string, _ time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.stored[key] = value
	return nil
}
func (m *mockLongTerm) Retrieve(_ context.Context, _, _ string) (string, error) { return "", nil }
func (m *mockLongTerm) Delete(_ context.Context, _, _ string) error             { return nil }
func (m *mockLongTerm) Query(_ context.Context, _, _ string) ([]string, error)  { return nil, nil }
func (m *mockLongTerm) Scan(_ context.Context, _, _ string) ([]string, error)   { return nil, nil }
func (m *mockLongTerm) Forget(_ context.Context, _, _, _ string) error          { return nil }

type mockPrivacy struct {
	failScrub bool
}

func (m *mockPrivacy) Scrub(_ context.Context, text string) (string, error) {
	if m.failScrub {
		return "", errors.New("scrub failed")
	}
	if text == "secret" {
		return "[REDACTED]", nil
	}
	return text, nil
}

func TestNewBasicConsolidator_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil dependencies")
		}
	}()
	NewBasicConsolidator(nil, nil, nil, 0)
}

func TestBasicConsolidator_Consolidate(t *testing.T) {
	shortTerm := &mockCandidateProvider{
		candidates: map[string]string{
			"key1": "val1",
			"key2": "secret",
		},
	}
	longTerm := &mockLongTerm{stored: make(map[string]string)}
	privacy := &mockPrivacy{}

	c := NewBasicConsolidator(shortTerm, longTerm, privacy, 0)
	err := c.Consolidate(context.Background(), "agent1", "session1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(longTerm.stored) != 2 {
		t.Fatalf("expected 2 stored items, got %d", len(longTerm.stored))
	}
	if longTerm.stored["key1"] != "val1" {
		t.Errorf("expected val1, got %q", longTerm.stored["key1"])
	}
	if longTerm.stored["key2"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", longTerm.stored["key2"])
	}
}

func TestBasicConsolidator_Consolidate_EmptyCandidates(t *testing.T) {
	shortTerm := &mockCandidateProvider{candidates: make(map[string]string)}
	longTerm := &mockLongTerm{stored: make(map[string]string)}
	privacy := &mockPrivacy{}

	c := NewBasicConsolidator(shortTerm, longTerm, privacy, 0)
	err := c.Consolidate(context.Background(), "agent1", "session1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(longTerm.stored) != 0 {
		t.Error("expected no items stored")
	}
}

func TestBasicConsolidator_Consolidate_EmptyArgs(t *testing.T) {
	shortTerm := &mockCandidateProvider{}
	longTerm := &mockLongTerm{}
	privacy := &mockPrivacy{}

	c := NewBasicConsolidator(shortTerm, longTerm, privacy, 0)
	err := c.Consolidate(context.Background(), "", "session1")
	if err == nil {
		t.Error("expected error for empty agentID")
	}
}

func TestBasicConsolidator_Consolidate_ScrubFailureSkipsItem(t *testing.T) {
	shortTerm := &mockCandidateProvider{
		candidates: map[string]string{
			"key1": "val1",
		},
	}
	longTerm := &mockLongTerm{stored: make(map[string]string)}
	privacy := &mockPrivacy{failScrub: true}

	c := NewBasicConsolidator(shortTerm, longTerm, privacy, 0)
	err := c.Consolidate(context.Background(), "agent1", "session1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(longTerm.stored) != 0 {
		t.Error("expected item to be skipped due to scrub failure")
	}
}

func TestBasicConsolidator_Consolidate_LongTermStoreError(t *testing.T) {
	shortTerm := &mockCandidateProvider{
		candidates: map[string]string{
			"key1": "val1",
		},
	}
	longTerm := &mockLongTerm{
		stored: make(map[string]string),
		err:    errors.New("store failed"),
	}
	privacy := &mockPrivacy{}

	c := NewBasicConsolidator(shortTerm, longTerm, privacy, 0)
	err := c.Consolidate(context.Background(), "agent1", "session1")
	if err == nil {
		t.Error("expected error from longTerm.Store, got nil")
	}
}
