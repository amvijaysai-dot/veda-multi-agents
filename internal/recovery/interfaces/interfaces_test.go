package interfaces_test

import (
	"context"
	"testing"

	"github.com/veda/agent-runtime/internal/recovery/interfaces"
	"github.com/veda/agent-runtime/internal/types/event"
)

// mockCheckpointer ensures that we can implement Checkpointer.
type mockCheckpointer struct{}

func (m *mockCheckpointer) Save(ctx context.Context, agentID string, state []byte, version string) (string, error) {
	return "", nil
}
func (m *mockCheckpointer) Delete(ctx context.Context, checkpointID string) error {
	return nil
}
func (m *mockCheckpointer) DeleteForAgent(ctx context.Context, agentID string) error {
	return nil
}

// mockRestorer ensures that we can implement Restorer.
type mockRestorer struct{}

func (m *mockRestorer) Load(ctx context.Context, checkpointID string) (*interfaces.Checkpoint, error) {
	return nil, nil
}
func (m *mockRestorer) LoadLatest(ctx context.Context, agentID string) (*interfaces.Checkpoint, error) {
	return nil, nil
}
func (m *mockRestorer) List(ctx context.Context, agentID string) ([]*interfaces.Checkpoint, error) {
	return nil, nil
}

// mockValidator ensures that we can implement Validator.
type mockValidator struct{}

func (m *mockValidator) Validate(ctx context.Context, cp *interfaces.Checkpoint) error {
	return nil
}

// mockReplayer ensures that we can implement Replayer.
type mockReplayer struct{}

func (m *mockReplayer) Replay(ctx context.Context, baseline *interfaces.Checkpoint, events []event.Event) ([]byte, error) {
	return nil, nil
}

// mockCoordinator ensures that we can implement RecoveryCoordinator.
type mockCoordinator struct{}

func (m *mockCoordinator) Recover(ctx context.Context, agentID string) ([]byte, error) {
	return nil, nil
}
func (m *mockCoordinator) CreateCheckpoint(ctx context.Context, agentID string, state []byte, version string) (string, error) {
	return "", nil
}

func TestInterfaceCompliance(t *testing.T) {
	var _ interfaces.Checkpointer = (*mockCheckpointer)(nil)
	var _ interfaces.Restorer = (*mockRestorer)(nil)
	var _ interfaces.Validator = (*mockValidator)(nil)
	var _ interfaces.Replayer = (*mockReplayer)(nil)
	var _ interfaces.RecoveryCoordinator = (*mockCoordinator)(nil)
}
