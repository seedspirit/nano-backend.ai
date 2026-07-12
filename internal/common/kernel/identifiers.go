package kernel

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Identifiers references the Session and Project a kernel belongs to.
type Identifiers struct {
	SessionID uuid.UUID `json:"session_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// NewIdentifiers constructs Identifiers, requiring both IDs to be non-nil.
func NewIdentifiers(sessionID, projectID uuid.UUID) (Identifiers, error) {
	if sessionID == uuid.Nil || projectID == uuid.Nil {
		return Identifiers{}, errordef.Errorf(errordef.InvalidInput, "identifiers require non-nil session and project IDs")
	}
	return Identifiers{SessionID: sessionID, ProjectID: projectID}, nil
}
