package kernel

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// ID identifies an agent-side kernel and carries enough context to route
// lifecycle calls without exposing a container ID as a domain concept.
type ID struct {
	SessionID uuid.UUID `json:"session_id"`
	AgentID   agent.ID  `json:"agent_id"`
	KernelID  string    `json:"kernel_id"`
}

// NewID constructs an ID, requiring session, agent, and kernel IDs.
func NewID(sessionID uuid.UUID, agentID agent.ID, kernelID string) (ID, error) {
	if sessionID == uuid.Nil {
		return ID{}, errordef.Errorf(errordef.InvalidInput, "kernel ID requires a session ID")
	}
	if agentID == (agent.ID{}) {
		return ID{}, errordef.Errorf(errordef.InvalidInput, "kernel ID requires an agent ID")
	}
	if kernelID == "" {
		return ID{}, errordef.Errorf(errordef.InvalidInput, "kernel ID is empty")
	}
	return ID{SessionID: sessionID, AgentID: agentID, KernelID: kernelID}, nil
}
