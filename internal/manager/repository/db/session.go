package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
	"github.com/seedspirit/nano-backend.ai/internal/manager/repository/db/entity"
)

// Args configures the SQLite session repository.
type Args struct {
	DBPath string
}

// SessionRepository reads session ledger data from SQLite.
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository opens, migrates, and returns a SQLite session repository.
func NewSessionRepository(ctx context.Context, args Args) (*SessionRepository, error) {
	dbx, err := Open(ctx, args)
	if err != nil {
		return nil, err
	}
	return &SessionRepository{db: dbx}, nil
}

// Close releases the repository database handle.
func (r *SessionRepository) Close() error {
	return r.db.Close()
}

// GetSpec returns the finalized spec for a session.
func (r *SessionRepository) GetSpec(ctx context.Context, sessionID uuid.UUID) (spec.Spec, error) {
	var row entity.Spec
	err := r.db.GetContext(ctx, &row, `
		SELECT id, project_id, type, name, description, model_base_model,
			resource_cpu_cores, resource_gpu_count, resource_memory_limit_bytes,
			resource_timeout_duration_seconds, created_at
		FROM sessions
		WHERE id = ?
	`, sessionID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return spec.Spec{}, errordef.ErrNotFound
	}
	if err != nil {
		return spec.Spec{}, fmt.Errorf("get spec for session %s: %w", sessionID, err)
	}
	specID, err := uuid.Parse(row.ID)
	if err != nil {
		return spec.Spec{}, fmt.Errorf("parse spec id %q: %w", row.ID, err)
	}
	refs, err := r.getSpecPresetRefs(ctx, specID)
	if err != nil {
		return spec.Spec{}, err
	}
	row.PresetRefs = refs

	datasets, err := r.getSpecDatasets(ctx, specID)
	if err != nil {
		return spec.Spec{}, err
	}
	row.Datasets = datasets

	parameters, err := r.getSpecTrainingParameters(ctx, specID)
	if err != nil {
		return spec.Spec{}, err
	}
	row.TrainingParameters = parameters

	return row.ToData()
}

// ListProjectSessions returns the most recent sessions for a project.
func (r *SessionRepository) ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error) {
	var exists int
	err := r.db.GetContext(ctx, &exists, `
		SELECT 1
		FROM projects
		WHERE id = ?
	`, projectID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errordef.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("check project %s exists: %w", projectID, err)
	}

	var rows []entity.Session
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT id, project_id, type, idempotency_key, status, result, failure_reason,
			created_at, started_at, finished_at
		FROM sessions
		WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, projectID.String(), limit); err != nil {
		return nil, fmt.Errorf("list sessions for project %s: %w", projectID, err)
	}

	sessions := make([]session.Session, 0, len(rows))
	for i := range rows {
		item, err := rows[i].ToData()
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, item)
	}
	return sessions, nil
}

// ProjectExists returns nil when the project exists, or errordef.ErrNotFound otherwise.
func (r *SessionRepository) ProjectExists(ctx context.Context, projectID uuid.UUID) error {
	var exists int
	err := r.db.GetContext(ctx, &exists, `SELECT 1 FROM projects WHERE id = ?`, projectID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return errordef.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("check project %s exists: %w", projectID, err)
	}
	return nil
}

// CreateSession persists a spec and a pending session in a single transaction.
func (r *SessionRepository) CreateSession(ctx context.Context, target *session.Session) error {
	if target == nil {
		return fmt.Errorf("session is required")
	}
	createdAt := encoding.FormatTime(target.Lifecycle.CreatedAt)

	specEntity, err := entity.FromData(&target.Definition, createdAt)
	if err != nil {
		return fmt.Errorf("convert spec: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create-session tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := insertSession(ctx, tx, &specEntity, target, createdAt); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create-session tx: %w", err)
	}
	return nil
}

func insertSession(ctx context.Context, tx *sqlx.Tx, e *entity.Spec, r *session.Session, createdAt string) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sessions (
			id, project_id, type, name, description, model_base_model,
			resource_cpu_cores, resource_gpu_count, resource_memory_limit_bytes,
			resource_timeout_duration_seconds, idempotency_key, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.ID.String(), r.ProjectID.String(), string(r.Type), e.Name, e.Description,
		e.ModelBaseModel, e.ResourceCPUCores, e.ResourceGPUCount,
		e.ResourceMemoryLimitBytes, e.ResourceTimeoutDurationSeconds,
		r.IdempotencyKey, string(r.Lifecycle.Status), createdAt); err != nil {
		return fmt.Errorf("insert session %s: %w", r.ID, err)
	}
	for _, ds := range e.Datasets {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_datasets (session_id, ordinal, dataset_ref, split_name)
			VALUES (?, ?, ?, ?)
		`, r.ID.String(), ds.Ordinal, ds.DatasetRef, ds.SplitName); err != nil {
			return fmt.Errorf("insert session dataset %s/%d: %w", r.ID, ds.Ordinal, err)
		}
	}
	for _, p := range e.TrainingParameters {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_training_parameters (session_id, key, value)
			VALUES (?, ?, ?)
		`, r.ID.String(), p.Key, p.Value); err != nil {
			return fmt.Errorf("insert session parameter %s/%s: %w", r.ID, p.Key, err)
		}
	}
	for _, ref := range []struct {
		category string
		id       *uuid.UUID
	}{
		{string(preset.TrainerPreset), e.PresetRefs.Trainer},
		{string(preset.ResourcePreset), e.PresetRefs.Resource},
		{string(preset.OutputPreset), e.PresetRefs.Output},
	} {
		if ref.id == nil || *ref.id == uuid.Nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_preset_refs (session_id, category, preset_id)
			VALUES (?, ?, ?)
		`, r.ID.String(), ref.category, ref.id.String()); err != nil {
			return fmt.Errorf("insert session preset ref %s/%s: %w", r.ID, ref.category, err)
		}
	}
	return nil
}

func (r *SessionRepository) getSpecDatasets(ctx context.Context, specID uuid.UUID) ([]entity.SpecDataset, error) {
	var rows []entity.SpecDataset
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT ordinal, dataset_ref, split_name
		FROM session_datasets
		WHERE session_id = ?
		ORDER BY ordinal
	`, specID.String()); err != nil {
		return nil, fmt.Errorf("get spec datasets %s: %w", specID, err)
	}
	return rows, nil
}

func (r *SessionRepository) getSpecTrainingParameters(ctx context.Context, specID uuid.UUID) ([]entity.SpecTrainingParameter, error) {
	var rows []entity.SpecTrainingParameter
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT key, value
		FROM session_training_parameters
		WHERE session_id = ?
	`, specID.String()); err != nil {
		return nil, fmt.Errorf("get spec training parameters %s: %w", specID, err)
	}
	return rows, nil
}

func (r *SessionRepository) getSpecPresetRefs(ctx context.Context, specID uuid.UUID) (session.PresetRefs, error) {
	var rows []struct {
		Category string `db:"category"`
		PresetID string `db:"preset_id"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT category, preset_id
		FROM session_preset_refs
		WHERE session_id = ?
	`, specID.String()); err != nil {
		return session.PresetRefs{}, fmt.Errorf("get spec preset refs %s: %w", specID, err)
	}

	var refs session.PresetRefs
	for _, row := range rows {
		id, err := uuid.Parse(row.PresetID)
		if err != nil {
			return session.PresetRefs{}, fmt.Errorf("parse preset id %q: %w", row.PresetID, err)
		}
		switch preset.Category(row.Category) {
		case preset.TrainerPreset:
			refs.Trainer = &id
		case preset.ResourcePreset:
			refs.Resource = &id
		case preset.OutputPreset:
			refs.Output = &id
		default:
			return session.PresetRefs{}, fmt.Errorf("unknown preset category %q", row.Category)
		}
	}
	return refs, nil
}
