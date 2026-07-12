package kernel

import (
	"slices"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Allocation is the provisioner's placement decision for a kernel.
type Allocation struct {
	AgentID agent.ID   `json:"agent_id"`
	GPUs    []GPUIndex `json:"gpus,omitempty"`
}

// NewAllocation constructs an Allocation, requiring a non-zero agent ID.
func NewAllocation(agentID agent.ID, gpus []GPUIndex) (Allocation, error) {
	if agentID == (agent.ID{}) {
		return Allocation{}, errordef.Errorf(errordef.InvalidInput, "allocation requires an agent ID")
	}
	return Allocation{AgentID: agentID, GPUs: slices.Clone(gpus)}, nil
}
