package workload

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Identifiers references the Run, Project, and Spec a workload belongs to.
type Identifiers struct {
	RunID     uuid.UUID `json:"run_id"`
	ProjectID uuid.UUID `json:"project_id"`
	SpecID    uuid.UUID `json:"spec_id"`
}

// NewIdentifiers constructs Identifiers, requiring all IDs to be non-nil.
func NewIdentifiers(runID, projectID, specID uuid.UUID) (Identifiers, error) {
	if runID == uuid.Nil || projectID == uuid.Nil || specID == uuid.Nil {
		return Identifiers{}, errordef.Errorf(errordef.InvalidInput, "identifiers require non-nil run, project, and spec IDs")
	}
	return Identifiers{RunID: runID, ProjectID: projectID, SpecID: specID}, nil
}
