// Package session defines Session lifecycle types: Session, Status, primitive resource
// shapes, and the artifact index produced by a completed Session.
package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a single execution instance of a Spec.
//
// A Session owns its identity and lifecycle metadata (timestamps, status) and
// references the Project + Spec it was created from. The same Spec may spawn
// multiple Sessions (e.g., reproducibility retries), each distinguished by its
// own id and optional client-provided idempotency key.
type Session struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	SpecID         uuid.UUID `json:"spec_id"`
	Type           Type      `json:"type"`
	IdempotencyKey *string   `json:"idempotency_key,omitempty"`
	Lifecycle      Lifecycle `json:"lifecycle"`
}

// NewWithSpec creates a batch Session for the given project and spec in Pending status.
func NewWithSpec(specID, projectID uuid.UUID, sessionType Type) Session {
	if sessionType == "" {
		sessionType = Batch
	}
	return Session{
		ID:        uuid.New(),
		ProjectID: projectID,
		SpecID:    specID,
		Type:      sessionType,
		Lifecycle: NewLifecycle(time.Now()),
	}
}

// Transition applies a lifecycle transition to the Session.
func (r *Session) Transition(transition Transition, at time.Time) error {
	return r.Lifecycle.Transition(transition, at)
}
