// Package session defines Session lifecycle types: Session, Status, primitive resource
// shapes, and the artifact index produced by a completed Session.
package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a single execution and owns its finalized definition.
type Session struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Type           Type      `json:"type"`
	IdempotencyKey *string   `json:"idempotency_key,omitempty"`
	Lifecycle      Lifecycle `json:"lifecycle"`
}

// Transition applies a lifecycle transition to the Session.
func (r *Session) Transition(transition Transition, at time.Time) error {
	return r.Lifecycle.Transition(transition, at)
}
