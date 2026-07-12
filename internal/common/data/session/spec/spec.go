package spec

import (
	"github.com/google/uuid"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
)

// Spec is the immutable, finalized input used to create a Session.
//
// A Spec is finalized before submission and persisted as part of exactly one
// Session. It captures the immutable execution definition owned by that Session.
type Spec struct {
	ID              uuid.UUID       `json:"id"`
	ProjectID       uuid.UUID       `json:"project_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Type            session.Type    `json:"type"`
	PresetRefs      preset.Refs     `json:"preset_refs,omitempty"`
	ModelOptions    ModelOptions    `json:"model_options"`
	DataOptions     DataOptions     `json:"data_options"`
	ResourceOptions ResourceOptions `json:"resource_options"`
	TrainingOptions TrainingOptions `json:"training_options"`
}

// New creates a finalized Spec with a fresh ID and the given identifying fields.
func New(projectID uuid.UUID, name string) Spec {
	return Spec{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
	}
}
