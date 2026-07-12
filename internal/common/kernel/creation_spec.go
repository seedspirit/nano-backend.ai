package kernel

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// CreationSpec is the immutable contract a manager hands to an agent to create
// a kernel. It is a thin container over focused value groups.
type CreationSpec struct {
	Identifiers   Identifiers   `json:"identifiers"`
	Execution     Execution     `json:"execution"`
	ResourceSlots ResourceSlots `json:"resource_slots"`
	Allocation    Allocation    `json:"allocation"`
	IO            IOBindings    `json:"io"`
}

// CreationSpecArgs carries the validated parts assembled into a CreationSpec.
type CreationSpecArgs struct {
	Identifiers   Identifiers
	Execution     Execution
	ResourceSlots ResourceSlots
	Allocation    Allocation
	IO            IOBindings
}

// NewCreationSpec assembles a CreationSpec, guarding required identity and execution fields.
func NewCreationSpec(args *CreationSpecArgs) (CreationSpec, error) {
	id := args.Identifiers
	if id.SessionID == uuid.Nil || id.ProjectID == uuid.Nil || id.SpecID == uuid.Nil {
		return CreationSpec{}, errordef.Errorf(errordef.InvalidInput, "kernel creation spec identifiers are incomplete")
	}
	if args.Execution.Image == (ImageRef{}) {
		return CreationSpec{}, errordef.Errorf(errordef.InvalidInput, "kernel creation spec image is required")
	}
	if args.Allocation.AgentID == (agent.ID{}) {
		return CreationSpec{}, errordef.Errorf(errordef.InvalidInput, "kernel allocation agent ID is required")
	}
	return CreationSpec{
		Identifiers:   args.Identifiers,
		Execution:     args.Execution,
		ResourceSlots: args.ResourceSlots,
		Allocation:    args.Allocation,
		IO:            args.IO,
	}, nil
}
