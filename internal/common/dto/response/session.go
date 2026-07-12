package response

import (
	"time"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
)

// SessionSummary is the stable list item shape for session navigation responses.
type SessionSummary struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"project_id"`
	SpecID         uuid.UUID              `json:"spec_id"`
	Type           session.Type           `json:"type"`
	IdempotencyKey *string                `json:"idempotency_key,omitempty"`
	Status         session.Status         `json:"status"`
	Result         session.Result         `json:"result"`
	FailureReason  *session.FailureReason `json:"failure_reason,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	FinishedAt     *time.Time             `json:"finished_at,omitempty"`
}

// ProjectSessionsData is the response data payload for project session lists.
type ProjectSessionsData struct {
	Sessions []SessionSummary `json:"sessions"`
	Limit    int              `json:"limit"`
}

// NewSessionSummary converts application session data into the external summary DTO.
func NewSessionSummary(source *session.Session) SessionSummary {
	return SessionSummary{
		ID:             source.ID,
		ProjectID:      source.ProjectID,
		SpecID:         source.SpecID,
		Type:           source.Type,
		IdempotencyKey: source.IdempotencyKey,
		Status:         source.Lifecycle.Status,
		Result:         source.Lifecycle.Result,
		FailureReason:  source.Lifecycle.FailureReason,
		CreatedAt:      source.Lifecycle.CreatedAt,
		StartedAt:      source.Lifecycle.StartedAt,
		FinishedAt:     source.Lifecycle.FinishedAt,
	}
}

// NewSessionSummaries converts application session data into external summary DTOs.
func NewSessionSummaries(source []session.Session) []SessionSummary {
	summaries := make([]SessionSummary, 0, len(source))
	for i := range source {
		summaries = append(summaries, NewSessionSummary(&source[i]))
	}
	return summaries
}
