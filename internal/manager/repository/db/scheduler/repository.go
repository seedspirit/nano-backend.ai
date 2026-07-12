package scheduler

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	managerdb "github.com/seedspirit/nano-backend.ai/internal/manager/repository/db"
)

// Args configures the SQLite scheduler repository.
type Args struct {
	DBPath string
}

// Repository reads scheduling candidates from SQLite.
type Repository struct {
	db *sqlx.DB
}

// NewRepository opens, migrates, and returns a scheduler repository.
func NewRepository(ctx context.Context, args Args) (*Repository, error) {
	dbx, err := managerdb.Open(ctx, managerdb.Args(args))
	if err != nil {
		return nil, err
	}
	return &Repository{db: dbx}, nil
}

// Close releases the repository database handle.
func (r *Repository) Close() error {
	return r.db.Close()
}

// ListPendingSessions returns pending Sessions without applying scheduling policy.
func (r *Repository) ListPendingSessions(ctx context.Context) ([]session.Session, error) {
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
