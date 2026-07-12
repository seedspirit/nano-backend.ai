package entity

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
)

// Session is the database record shape for a session row.
type Session struct {
	ID             string         `db:"id"`
	ProjectID      string         `db:"project_id"`
	Type           string         `db:"type"`
	IdempotencyKey sql.NullString `db:"idempotency_key"`
	Status         string         `db:"status"`
	Result         string         `db:"result"`
	FailureReason  sql.NullString `db:"failure_reason"`
	CreatedAt      string         `db:"created_at"`
	StartedAt      sql.NullString `db:"started_at"`
	FinishedAt     sql.NullString `db:"finished_at"`
}

// ToData converts the database record into the public session type.
func (r *Session) ToData() (session.Session, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return session.Session{}, fmt.Errorf("parse run id %q: %w", r.ID, err)
	}
	projectID, err := uuid.Parse(r.ProjectID)
	if err != nil {
		return session.Session{}, fmt.Errorf("parse project id %q: %w", r.ProjectID, err)
	}
	createdAt, err := encoding.ParseTime(r.CreatedAt)
	if err != nil {
		return session.Session{}, err
	}

	lifecycle := session.Lifecycle{
		Status:    session.Status(r.Status),
		Result:    session.Result(r.Result),
		CreatedAt: createdAt,
	}
	if r.FailureReason.Valid {
		reason := session.FailureReason(r.FailureReason.String)
		lifecycle.FailureReason = &reason
	}
	if r.StartedAt.Valid {
		startedAt, err := encoding.ParseTime(r.StartedAt.String)
		if err != nil {
			return session.Session{}, err
		}
		lifecycle.StartedAt = &startedAt
	}
	if r.FinishedAt.Valid {
		finishedAt, err := encoding.ParseTime(r.FinishedAt.String)
		if err != nil {
			return session.Session{}, err
		}
		lifecycle.FinishedAt = &finishedAt
	}

	result := session.Session{
		Definition: session.Definition{
			ID:        id,
			ProjectID: projectID,
			Type:      session.Type(r.Type),
		},
		Lifecycle: lifecycle,
	}
	if r.IdempotencyKey.Valid {
		key := r.IdempotencyKey.String
		result.IdempotencyKey = &key
	}
	return result, nil
}
