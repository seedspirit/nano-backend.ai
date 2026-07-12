package spec

import "github.com/seedspirit/nano-backend.ai/internal/common/data/session"

// ModelOptions describes the finalized base model for a Session.
type ModelOptions struct {
	BaseModel string `json:"base_model"`
}

// DataOptions describes the finalized dataset(s) used by a Session.
type DataOptions struct {
	Datasets []DatasetRef `json:"datasets"`
}

// DatasetRef identifies a finalized dataset and split.
type DatasetRef struct {
	Path  string `json:"path"`
	Split string `json:"split"`
}

// ResourceOptions specifies finalized compute resources for a Session.
type ResourceOptions struct {
	CPU     session.CPUOptions     `json:"cpu,omitempty"`
	GPU     session.GPUOptions     `json:"gpu"`
	Memory  session.MemoryOptions  `json:"memory"`
	Timeout session.TimeoutOptions `json:"timeout"`
}

// TrainingOptions holds finalized training parameters.
type TrainingOptions struct {
	Parameters map[string]any `json:"parameters,omitempty"`
}
