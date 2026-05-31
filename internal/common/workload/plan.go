package workload

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Plan is the immutable contract a manager hands to an agent to prepare and
// start a workload. It is a thin container over focused value groups.
type Plan struct {
	Identifiers Identifiers `json:"identifiers"`
	Execution   Execution   `json:"execution"`
	Resources   Resources   `json:"resources"`
	Assignment  Assignment  `json:"assignment"`
	IO          IOBindings  `json:"io"`
}

// PlanArgs carries the validated parts assembled into a Plan.
type PlanArgs struct {
	Identifiers Identifiers
	Execution   Execution
	Resources   Resources
	Assignment  Assignment
	IO          IOBindings
}

// NewPlan assembles a Plan, guarding the required identifiers, image, and agent ID.
func NewPlan(args *PlanArgs) (Plan, error) {
	id := args.Identifiers
	if id.RunID == uuid.Nil || id.ProjectID == uuid.Nil || id.SpecID == uuid.Nil {
		return Plan{}, errordef.Errorf(errordef.InvalidInput, "workload plan identifiers are incomplete")
	}
	if args.Execution.Image == (ImageRef{}) {
		return Plan{}, errordef.Errorf(errordef.InvalidInput, "workload plan execution image is required")
	}
	if args.Assignment.AgentID == (agent.ID{}) {
		return Plan{}, errordef.Errorf(errordef.InvalidInput, "workload plan assignment agent ID is required")
	}
	return Plan{
		Identifiers: args.Identifiers,
		Execution:   args.Execution,
		Resources:   args.Resources,
		Assignment:  args.Assignment,
		IO:          args.IO,
	}, nil
}
