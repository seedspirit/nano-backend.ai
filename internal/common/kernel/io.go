package kernel

import "slices"

// Mount binds an agent-side source path into the kernel container.
type Mount struct {
	Source   AgentPath `json:"source"`
	Target   string    `json:"target"`
	ReadOnly bool      `json:"read_only,omitempty"`
}

// IOBindings groups the kernel's mounts and agent-side artifact/log/config paths.
type IOBindings struct {
	Mounts       []Mount   `json:"mounts,omitempty"`
	ArtifactPath AgentPath `json:"artifact_path"`
	LogPath      AgentPath `json:"log_path"`
	ConfigPath   AgentPath `json:"config_path"`
}

// IOArgs carries the inputs to NewIOBindings.
type IOArgs struct {
	Mounts       []Mount
	ArtifactPath AgentPath
	LogPath      AgentPath
	ConfigPath   AgentPath
}

// NewIOBindings constructs IOBindings, copying the Mounts slice.
func NewIOBindings(args *IOArgs) IOBindings {
	return IOBindings{
		Mounts:       slices.Clone(args.Mounts),
		ArtifactPath: args.ArtifactPath,
		LogPath:      args.LogPath,
		ConfigPath:   args.ConfigPath,
	}
}
