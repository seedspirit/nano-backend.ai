package session

import "github.com/google/uuid"

// Definition is the immutable execution definition owned by a Session.
type Definition struct {
	ID              uuid.UUID       `json:"id"`
	ProjectID       uuid.UUID       `json:"project_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Type            Type            `json:"type"`
	PresetRefs      PresetRefs      `json:"preset_refs,omitempty"`
	ModelOptions    ModelOptions    `json:"model_options"`
	DataOptions     DataOptions     `json:"data_options"`
	ResourceOptions ResourceOptions `json:"resource_options"`
	TrainingOptions TrainingOptions `json:"training_options"`
}

// PresetRefs records the presets used to finalize a Session definition.
type PresetRefs struct {
	Trainer  *uuid.UUID `json:"trainer,omitempty"`
	Resource *uuid.UUID `json:"resource,omitempty"`
	Output   *uuid.UUID `json:"output,omitempty"`
}

// ModelOptions describes the finalized base model for a Session.
type ModelOptions struct {
	BaseModel string `json:"base_model"`
}

// DataOptions describes the finalized datasets used by a Session.
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
	CPU     CPUOptions     `json:"cpu,omitempty"`
	GPU     GPUOptions     `json:"gpu"`
	Memory  MemoryOptions  `json:"memory"`
	Timeout TimeoutOptions `json:"timeout"`
}

// TrainingOptions holds finalized training parameters.
type TrainingOptions struct {
	Parameters map[string]any `json:"parameters,omitempty"`
}
