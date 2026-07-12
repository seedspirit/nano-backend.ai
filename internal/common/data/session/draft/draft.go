package draft

import (
	"github.com/google/uuid"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
)

// Draft is the persisted input used by a spec builder to create a finalized Spec.
type Draft struct {
	ID              uuid.UUID          `json:"id"`
	ProjectID       uuid.UUID          `json:"project_id"`
	Name            string             `json:"name"`
	Description     string             `json:"description,omitempty"`
	Type            session.Type       `json:"type"`
	PresetRefs      session.PresetRefs `json:"preset_refs,omitempty"`
	ModelOptions    ModelOptionsReq    `json:"model_options"`
	DataOptions     DataOptionsReq     `json:"data_options"`
	ResourceOptions ResourceOptionsReq `json:"resource_options"`
	TrainingOptions TrainingOptionsReq `json:"training_options"`
}
