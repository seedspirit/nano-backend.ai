package kernel

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Identifiers references the Session, Project, and Spec a kernel belongs to.
type Identifiers struct {
	SessionID uuid.UUID `json:"session_id"`
	ProjectID uuid.UUID `json:"project_id"`
	SpecID    uuid.UUID `json:"spec_id"`
}

// NewIdentifiers constructs Identifiers, requiring all IDs to be non-nil.
func NewIdentifiers(sessionID, projectID, specID uuid.UUID) (Identifiers, error) {
	if sessionID == uuid.Nil || projectID == uuid.Nil || specID == uuid.Nil {
		return Identifiers{}, errordef.Errorf(errordef.InvalidInput, "identifiers require non-nil session, project, and spec IDs")
	}
	return Identifiers{SessionID: sessionID, ProjectID: projectID, SpecID: specID}, nil
}
