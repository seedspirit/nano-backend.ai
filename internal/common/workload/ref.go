package workload

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Ref is an opaque reference to an agent-side prepared workload, carrying enough
// identity to route follow-up calls. It is the result type a launcher port
// (defined by its consumer) returns from preparation.
type Ref struct {
	RunID           uuid.UUID `json:"run_id"`
	AgentID         agent.ID  `json:"agent_id"`
	AgentWorkloadID string    `json:"agent_workload_id"`
}

// NewRef constructs a Ref, requiring run ID, agent ID, and agent workload ID.
func NewRef(runID uuid.UUID, agentID agent.ID, agentWorkloadID string) (Ref, error) {
	if runID == uuid.Nil {
		return Ref{}, errordef.Errorf(errordef.InvalidInput, "workload ref requires a run ID")
	}
	if agentID == (agent.ID{}) {
		return Ref{}, errordef.Errorf(errordef.InvalidInput, "workload ref requires an agent ID")
	}
	if agentWorkloadID == "" {
		return Ref{}, errordef.Errorf(errordef.InvalidInput, "workload ref requires an agent workload ID")
	}
	return Ref{RunID: runID, AgentID: agentID, AgentWorkloadID: agentWorkloadID}, nil
}
