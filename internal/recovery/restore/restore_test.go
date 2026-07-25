package restore

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/veda/agent-runtime/internal/recovery/checkpoint"
)

func TestFileRestorer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "veda-restore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cp, err := checkpoint.NewFileCheckpointer(tempDir)
	if err != nil {
		t.Fatalf("failed to create checkpointer: %v", err)
	}

	restorer := NewFileRestorer(tempDir)
	ctx := context.Background()

	agentID := "agent-123"

	// List should be empty
	list, err := restorer.List(ctx, agentID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 checkpoints, got %d", len(list))
	}

	// Create checkpoints
	id1, _ := cp.Save(ctx, agentID, []byte("state1"), "v1")
	time.Sleep(10 * time.Millisecond) // Ensure distinct timestamps
	id2, _ := cp.Save(ctx, agentID, []byte("state2"), "v1")

	// Load specific
	loadedCp, err := restorer.Load(ctx, id1)
	if err != nil {
		t.Fatalf("failed to load checkpoint %s: %v", id1, err)
	}

	// Decompress and verify
	data, _ := checkpoint.DecompressState(loadedCp.Data)
	if string(data) != "state1" {
		t.Errorf("expected state1, got %s", string(data))
	}

	// LoadLatest
	latest, err := restorer.LoadLatest(ctx, agentID)
	if err != nil {
		t.Fatalf("failed to load latest: %v", err)
	}
	if latest.ID != id2 {
		t.Errorf("expected latest to be %s, got %s", id2, latest.ID)
	}

	latestData, _ := checkpoint.DecompressState(latest.Data)
	if string(latestData) != "state2" {
		t.Errorf("expected latest state to be state2, got %s", string(latestData))
	}

	// List
	list, err = restorer.List(ctx, agentID)
	if err != nil {
		t.Fatalf("failed to list checkpoints: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(list))
	}
	if list[0].ID != id1 || list[1].ID != id2 {
		t.Errorf("list order incorrect")
	}

	// Delete and verify List changes
	cp.Delete(ctx, id1)

	list, _ = restorer.List(ctx, agentID)
	if len(list) != 1 || list[0].ID != id2 {
		t.Errorf("expected only id2 in list after delete")
	}
}
