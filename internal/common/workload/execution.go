package workload

import (
	"maps"
	"slices"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/run"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Execution describes what to run: the image, optional entrypoint/command,
// environment, and time budget. These values are agent-independent.
type Execution struct {
	Image      ImageRef           `json:"image"`
	Entrypoint []string           `json:"entrypoint,omitempty"`
	Command    []string           `json:"command,omitempty"`
	Env        map[string]string  `json:"env,omitempty"`
	Timeout    run.TimeoutOptions `json:"timeout"`
}

// ExecutionArgs carries the inputs to NewExecution.
type ExecutionArgs struct {
	Image      ImageRef
	Entrypoint []string
	Command    []string
	Env        map[string]string
	Timeout    run.TimeoutOptions
}

// NewExecution constructs an Execution; the image is required and slices/map
// are copied so later caller mutation cannot affect the stored value.
func NewExecution(args *ExecutionArgs) (Execution, error) {
	if args.Image == (ImageRef{}) {
		return Execution{}, errordef.Errorf(errordef.InvalidInput, "execution requires an image")
	}
	return Execution{
		Image:      args.Image,
		Entrypoint: slices.Clone(args.Entrypoint),
		Command:    slices.Clone(args.Command),
		Env:        maps.Clone(args.Env),
		Timeout:    args.Timeout,
	}, nil
}
