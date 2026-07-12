package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
	"github.com/seedspirit/nano-backend.ai/internal/manager/repository/db/entity"
)

// SchedulerArgs configures the SQLite scheduler repository.
type SchedulerArgs struct {
	DBPath string
}

// SchedulerRepository reads scheduling candidates from SQLite.
type SchedulerRepository struct {
	db *sqlx.DB
}

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

// NewSchedulerRepository opens, migrates, and returns a scheduler repository.
func NewSchedulerRepository(ctx context.Context, args SchedulerArgs) (*SchedulerRepository, error) {
	dbx, err := Open(ctx, Args(args))
	if err != nil {
		return nil, err
	}
	return &SchedulerRepository{db: dbx}, nil
}

// Close releases the repository database handle.
func (r *SchedulerRepository) Close() error {
	return r.db.Close()
}

// ListPendingSessions returns pending Sessions without applying scheduling policy.
func (r *SchedulerRepository) ListPendingSessions(ctx context.Context) ([]session.Session, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin pending sessions read transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var rows []pendingSessionRow
	if err := tx.SelectContext(ctx, &rows, `
		SELECT id, project_id, type, name, description, model_base_model,
			resource_cpu_cores, resource_gpu_count, resource_memory_limit_bytes,
			resource_timeout_duration_seconds, idempotency_key, status, result,
			failure_reason, created_at, started_at, finished_at,
			COALESCE((
				SELECT json_group_array(json_object(
					'category', category,
					'preset_id', preset_id
				))
				FROM session_preset_refs
				WHERE session_id = sessions.id
			), '[]') AS preset_refs_json,
			COALESCE((
				SELECT json_group_array(json_object(
					'Ordinal', ordinal,
					'DatasetRef', dataset_ref,
					'SplitName', split_name
				))
				FROM (
					SELECT ordinal, dataset_ref, split_name
					FROM session_datasets
					WHERE session_id = sessions.id
					ORDER BY ordinal
				)
			), '[]') AS datasets_json,
			COALESCE((
				SELECT json_group_array(json_object(
					'Key', key,
					'Value', value
				))
				FROM session_training_parameters
				WHERE session_id = sessions.id
			), '[]') AS training_parameters_json
		FROM sessions
		WHERE status = ?
	`, string(session.Pending)); err != nil {
		return nil, fmt.Errorf("list pending sessions: %w", err)
	}

	result := make([]session.Session, 0, len(rows))
	for i := range rows {
		item, err := rows[i].ToData()
		if err != nil {
			return nil, fmt.Errorf("map pending session: %w", err)
		}
		definition, err := rows[i].toDefinition()
		if err != nil {
			return nil, fmt.Errorf("map definition for pending session %s: %w", item.ID, err)
		}
		item.Definition = definition
		result = append(result, item)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pending sessions read transaction: %w", err)
	}
	return result, nil
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
