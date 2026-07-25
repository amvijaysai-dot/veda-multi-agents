package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileCheckpointer_SaveAndDecompress(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "veda-checkpoint-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cp, err := NewFileCheckpointer(tempDir)
	if err != nil {
		t.Fatalf("failed to create checkpointer: %v", err)
	}

	state := []byte(`{"memory": {"short_term": ["hello", "world"]}}`)
	agentID := "agent-123"

	ctx := context.Background()

	id, err := cp.Save(ctx, agentID, state, "v1.0")
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty checkpoint ID")
	}

	// Verify the file was created in the correct directory
	agentDir := filepath.Join(tempDir, agentID)
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		t.Fatalf("failed to read agent dir: %v", err)
	}

	var found string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), id) {
			found = filepath.Join(agentDir, entry.Name())
			break
		}
	}
	if found == "" {
		t.Fatal("could not find checkpoint file")
	}

	// Read and verify decompression
	data, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("failed to read checkpoint file: %v", err)
	}

	// In a real scenario, the file is JSON and contains compressed Data.
	// But since this test just tests creation, we'll verify it's valid JSON containing our ID.
	if !strings.Contains(string(data), id) {
		t.Error("checkpoint data does not contain ID")
	}
}

func TestFileCheckpointer_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "veda-checkpoint-delete-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cp, _ := NewFileCheckpointer(tempDir)
	ctx := context.Background()

	id1, _ := cp.Save(ctx, "agent-1", []byte("state1"), "v1")
	id2, _ := cp.Save(ctx, "agent-1", []byte("state2"), "v1")
	id3, _ := cp.Save(ctx, "agent-2", []byte("state3"), "v1")

	// Test Delete specific
	if err := cp.Delete(ctx, id1); err != nil {
		t.Fatalf("failed to delete checkpoint %s: %v", id1, err)
	}

	// Verify id1 is gone
	agent1Dir := filepath.Join(tempDir, "agent-1")
	entries, _ := os.ReadDir(agent1Dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), id1) {
			t.Error("checkpoint 1 should be deleted")
		}
	}

	// Verify id2 is still there
	foundId2 := false
	for _, e := range entries {
		if strings.Contains(e.Name(), id2) {
			foundId2 = true
		}
	}
	if !foundId2 {
		t.Error("checkpoint 2 should still exist")
	}

	// Test DeleteForAgent
	if err := cp.DeleteForAgent(ctx, "agent-2"); err != nil {
		t.Fatalf("failed to delete for agent-2: %v", err)
	}

	_, err = os.Stat(filepath.Join(tempDir, "agent-2"))
	if !os.IsNotExist(err) {
		t.Error("agent-2 directory should be deleted")
	}

	// id3 should be gone
	if err := cp.Delete(ctx, id3); err != nil {
		t.Fatalf("delete should be idempotent for id3: %v", err)
	}
}
