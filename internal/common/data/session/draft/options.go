package draft

import "github.com/seedspirit/nano-backend.ai/internal/common/data/session"

// ModelOptionsReq describes the requested base model for a Draft.
type ModelOptionsReq struct {
	BaseModel string `json:"base_model"`
}

// DataOptionsReq describes the requested dataset(s) for a Draft.
type DataOptionsReq struct {
	Datasets []DatasetRefReq `json:"datasets"`
}

// DatasetRefReq identifies a requested dataset and split.
type DatasetRefReq struct {
	Path  string `json:"path"`
	Split string `json:"split"`
}

// ResourceOptionsReq specifies requested compute resources for a Draft.
type ResourceOptionsReq struct {
	CPU     session.CPUOptions     `json:"cpu,omitempty"`
	GPU     session.GPUOptions     `json:"gpu"`
	Memory  session.MemoryOptions  `json:"memory"`
	Timeout session.TimeoutOptions `json:"timeout"`
}

// TrainingOptionsReq holds user-provided training parameters.
type TrainingOptionsReq struct {
	Parameters map[string]any `json:"parameters,omitempty"`
}
