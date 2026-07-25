// Package execcontext provides the prototype ContextManager implementation for
// the VEDA Agent Runtime execution engine.
//
// The ContextManager is responsible for building the system prompt that is sent
// to the reasoning engine at the start of each turn, and for maintaining the
// running observation history within a turn.
//
// In v0.4 all memory retrieval is stubbed. The real memory integration
// (ShortTermMemory + LongTermMemory) will be wired in Milestone v0.5.
//
// Package name "execcontext" avoids collision with the standard "context" package.
//
// Dependency rule: this package imports execution/interfaces only.
package execcontext

import (
	"context"
	"fmt"
	"strings"

	"github.com/veda/agent-runtime/internal/execution/interfaces"
)

// PrototypeContextManager builds prompts and maintains observation history
// for a single agent turn. It is not safe to share across concurrent turns;
// each turn should use its own instance or call Reset between turns.
type PrototypeContextManager struct {
	// agentDescription is a static description of the agent's role, incorporated
	// into every system prompt.
	agentDescription string

	// maxHistoryEntries is the maximum number of observation entries retained.
	// Older entries are dropped when the limit is exceeded (FIFO).
	maxHistoryEntries int
}

// NewPrototypeContextManager creates a PrototypeContextManager with the given
// agent description and a maximum history size.
//
// If maxHistoryEntries is ≤ 0 it is set to the default of 50.
func NewPrototypeContextManager(agentDescription string, maxHistoryEntries int) *PrototypeContextManager {
	if maxHistoryEntries <= 0 {
		maxHistoryEntries = 50
	}
	return &PrototypeContextManager{
		agentDescription:  agentDescription,
		maxHistoryEntries: maxHistoryEntries,
	}
}

// BuildPrompt constructs the system prompt for the given agent and session.
// It incorporates the static agent description, available tool names, and a
// stub memory section (v0.4: always empty).
//
// BuildPrompt respects ctx cancellation.
//
// BuildPrompt implements interfaces.ContextManager.
func (c *PrototypeContextManager) BuildPrompt(
	ctx context.Context,
	agentID, sessionID string,
	tools []string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context_manager: context cancelled: %w", err)
	}
	if agentID == "" {
		return "", fmt.Errorf("context_manager: agentID must not be empty")
	}

	var b strings.Builder

	// Agent identity header.
	b.WriteString(fmt.Sprintf("You are agent %q (session: %s).\n", agentID, sessionID))

	// Optional agent description.
	if desc := strings.TrimSpace(c.agentDescription); desc != "" {
		b.WriteString(desc)
		b.WriteString("\n")
	}

	// Available tools section.
	if len(tools) > 0 {
		b.WriteString("\nAvailable tools:\n")
		for _, t := range tools {
			b.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}

	// Memory section (stubbed in v0.4).
	b.WriteString("\n[Memory: not available in v0.4]\n")

	return b.String(), nil
}

// AppendObservation appends a formatted tool result to the history string and
// returns the updated history. When the number of newline-delimited entries
// in history would exceed maxHistoryEntries, the oldest entry is trimmed.
//
// AppendObservation implements interfaces.ContextManager.
func (c *PrototypeContextManager) AppendObservation(history string, result interfaces.ToolResult) string {
	var entry string
	if result.Err != nil {
		entry = fmt.Sprintf("[Tool: %s | Error: %v]", result.ToolName, result.Err)
	} else {
		entry = fmt.Sprintf("[Tool: %s | Output: %s]", result.ToolName, result.Output)
	}

	if history == "" {
		return entry
	}

	combined := history + "\n" + entry

	// Enforce history length limit.
	lines := strings.Split(combined, "\n")
	if len(lines) > c.maxHistoryEntries {
		lines = lines[len(lines)-c.maxHistoryEntries:]
	}

	return strings.Join(lines, "\n")
}
