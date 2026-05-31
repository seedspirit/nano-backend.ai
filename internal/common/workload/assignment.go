package workload

import (
	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Assignment is the provisioner's placement decision for a workload.
type Assignment struct {
	AgentID agent.ID `json:"agent_id"`
}

// NewAssignment constructs an Assignment, requiring a non-zero agent ID.
func NewAssignment(agentID agent.ID) (Assignment, error) {
	if agentID == (agent.ID{}) {
		return Assignment{}, errordef.Errorf(errordef.InvalidInput, "assignment requires an agent ID")
	}
	return Assignment{AgentID: agentID}, nil
}
