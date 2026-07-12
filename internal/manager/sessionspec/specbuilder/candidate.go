package specbuilder

import (
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
)

// Candidate is the validation target created from user draft and resolved presets.
type Candidate struct {
	Draft   *draft.Draft
	Presets preset.Presets
}
