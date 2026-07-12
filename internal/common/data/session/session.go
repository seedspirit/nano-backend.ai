// Package session defines Session lifecycle types: Session, Status, primitive resource
// shapes, and the artifact index produced by a completed Session.
package session

import "time"

// Session represents a single execution and owns its finalized definition.
type Session struct {
	Definition
	IdempotencyKey *string   `json:"idempotency_key,omitempty"`
	Lifecycle      Lifecycle `json:"lifecycle"`
}

// NewPending creates a Session from its finalized definition.
func NewPending(definition Definition) Session { //nolint:gocritic // The Session intentionally owns an immutable definition copy.
	if definition.Type == "" {
		definition.Type = Batch
	}
	return Session{
		Definition: definition,
		Lifecycle:  NewLifecycle(time.Now()),
	}
}

// Transition applies a lifecycle transition to the Session.
func (r *Session) Transition(transition Transition, at time.Time) error {
	return r.Lifecycle.Transition(transition, at)
}
