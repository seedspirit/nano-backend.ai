package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
	"github.com/seedspirit/nano-backend.ai/internal/manager/repository/db/entity"
)

type pendingSessionRow struct {
	entity.Session
	Name                           string `db:"name"`
	Description                    string `db:"description"`
	ModelBaseModel                 string `db:"model_base_model"`
	ResourceCPUCores               int    `db:"resource_cpu_cores"`
	ResourceGPUCount               int    `db:"resource_gpu_count"`
	ResourceMemoryLimitBytes       int64  `db:"resource_memory_limit_bytes"`
	ResourceTimeoutDurationSeconds int64  `db:"resource_timeout_duration_seconds"`
	PresetRefsJSON                 string `db:"preset_refs_json"`
	DatasetsJSON                   string `db:"datasets_json"`
	TrainingParametersJSON         string `db:"training_parameters_json"`
}

type presetRefRow struct {
	Category string `json:"category"`
	PresetID string `json:"preset_id"`
}

func (r *pendingSessionRow) toDefinition() (session.Definition, error) {
	row := entity.Spec{
		ID:                             r.ID,
		ProjectID:                      r.ProjectID,
		Type:                           r.Type,
		Name:                           r.Name,
		Description:                    r.Description,
		ModelBaseModel:                 r.ModelBaseModel,
		ResourceCPUCores:               r.ResourceCPUCores,
		ResourceGPUCount:               r.ResourceGPUCount,
		ResourceMemoryLimitBytes:       r.ResourceMemoryLimitBytes,
		ResourceTimeoutDurationSeconds: r.ResourceTimeoutDurationSeconds,
		CreatedAt:                      r.CreatedAt,
	}

	var presetRows []presetRefRow
	if err := json.Unmarshal([]byte(r.PresetRefsJSON), &presetRows); err != nil {
		return session.Definition{}, fmt.Errorf("decode preset refs: %w", err)
	}
	for _, presetRow := range presetRows {
		id, err := uuid.Parse(presetRow.PresetID)
		if err != nil {
			return session.Definition{}, fmt.Errorf("parse preset id %q: %w", presetRow.PresetID, err)
		}
		switch preset.Category(presetRow.Category) {
		case preset.TrainerPreset:
			row.PresetRefs.Trainer = &id
		case preset.ResourcePreset:
			row.PresetRefs.Resource = &id
		case preset.OutputPreset:
			row.PresetRefs.Output = &id
		default:
			return session.Definition{}, fmt.Errorf("unknown preset category %q", presetRow.Category)
		}
	}

	if err := json.Unmarshal([]byte(r.DatasetsJSON), &row.Datasets); err != nil {
		return session.Definition{}, fmt.Errorf("decode datasets: %w", err)
	}
	if err := json.Unmarshal([]byte(r.TrainingParametersJSON), &row.TrainingParameters); err != nil {
		return session.Definition{}, fmt.Errorf("decode training parameters: %w", err)
	}

	return row.ToData()
}
