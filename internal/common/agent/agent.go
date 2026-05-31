// Package agent defines shared agent identity types used across manager and
// agent boundaries.
package agent

import (
	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// ID uniquely identifies an agent. The zero value is invalid; use NewID or
// ParseID.
type ID struct {
	uuid uuid.UUID
}

// NewID generates a new random agent ID.
func NewID() ID { return ID{uuid: uuid.New()} }

// ParseID parses a UUID string into an ID.
func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, errordef.Errorf(errordef.InvalidInput, "invalid agent ID: %q", s)
	}
	return ID{uuid: id}, nil
}

// String returns the UUID string form.
func (a ID) String() string { return a.uuid.String() }

// MarshalJSON encodes the ID as its UUID string.
func (a ID) MarshalJSON() ([]byte, error) {
	s, err := encoding.MarshalJSON(a.uuid.String())
	return []byte(s), err
}

// UnmarshalJSON decodes an ID from a UUID string.
func (a *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := encoding.UnmarshalJSON(string(data), &s); err != nil {
		return err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return errordef.Errorf(errordef.InvalidInput, "invalid agent ID: %q", s)
	}
	a.uuid = id
	return nil
}
