package kernel

import "github.com/seedspirit/nano-backend.ai/internal/common/data/session"

// GPUIndex is an agent-local GPU device index, distinct from the logical GPU
// count in session.GPUOptions.
type GPUIndex int

// ResourceSlots is the compute capacity requested by a session. Physical GPU
// device indices belong to Allocation, which is produced by the provisioner.
type ResourceSlots struct {
	CPU    session.CPUOptions    `json:"cpu"`
	Memory session.MemoryOptions `json:"memory"`
}

// NewResourceSlots constructs ResourceSlots.
func NewResourceSlots(cpu session.CPUOptions, memory session.MemoryOptions) ResourceSlots {
	return ResourceSlots{CPU: cpu, Memory: memory}
}
