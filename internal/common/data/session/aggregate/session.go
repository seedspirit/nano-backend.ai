// Package aggregate defines a complete Session ready for persistence.
package aggregate

import (
	"time"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
)

// Session combines a lifecycle record with the finalized definition it owns.
type Session struct {
	Record     session.Session
	Definition spec.Spec
}

// New creates a pending Session from a finalized definition.
func New(definition spec.Spec) Session { //nolint:gocritic // The aggregate intentionally owns an immutable definition copy.
	sessionType := definition.Type
	if sessionType == "" {
		sessionType = session.Batch
		definition.Type = sessionType
	}
	return Session{
		Record: session.Session{
			ID:        definition.ID,
			ProjectID: definition.ProjectID,
			Type:      sessionType,
			Lifecycle: session.NewLifecycle(time.Now()),
		},
		Definition: definition,
	}
}
