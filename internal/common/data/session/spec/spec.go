package spec

import (
	"github.com/google/uuid"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
)

// Spec is the immutable, finalized input used to create a Session.
//
// A Spec is finalized before submission and persisted as part of exactly one
// Session. It captures the immutable execution definition owned by that Session.
type Spec = session.Definition

// New creates a finalized Spec with a fresh ID and the given identifying fields.
func New(projectID uuid.UUID, name string) Spec {
	return Spec{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
	}
}
