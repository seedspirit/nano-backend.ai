package workload

import (
	"slices"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/run"
)

// GPUIndex is an agent-local GPU device index, distinct from the logical GPU
// count in run.GPUOptions.
type GPUIndex int

// Resources is the compute allocated to a workload: CPU and Memory copied from
// the Spec, and the agent-local GPU device indices a provisioner assigns.
type Resources struct {
	CPU    run.CPUOptions    `json:"cpu"`
	Memory run.MemoryOptions `json:"memory"`
	GPUs   []GPUIndex        `json:"gpus,omitempty"`
}

// NewResources constructs Resources, copying the GPUs slice.
func NewResources(cpu run.CPUOptions, memory run.MemoryOptions, gpus []GPUIndex) Resources {
	return Resources{CPU: cpu, Memory: memory, GPUs: slices.Clone(gpus)}
}
