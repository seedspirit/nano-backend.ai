// Package runtime defines the runtime abstraction and its data types.
//
// A runtime instance represents an executable unit (e.g., a local process, a container)
// whose lifecycle is managed through the Runtime interface.
package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ── Types ──────────────────────────────────────────────

// ID is a UUID-based unique identifier for a runtime instance.
// The zero value is invalid; use NewID or ParseID to construct.
type ID struct {
	uuid uuid.UUID
}

// NewID generates a new random runtime ID.
func NewID() ID {
	return ID{uuid: uuid.New()}
}

// ParseID parses a string into an ID.
// Returns ErrInvalidID if the string is not a valid UUID.
func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, ErrInvalidID
	}
	return ID{uuid: id}, nil
}

// String returns the UUID string representation.
func (id ID) String() string {
	return id.uuid.String()
}

// IsZero reports whether the ID is the zero value (uninitialized).
func (id ID) IsZero() bool {
	return id.uuid == uuid.Nil
}

// MarshalJSON implements json.Marshaler.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.uuid.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return ErrInvalidID
	}
	id.uuid = parsed
	return nil
}

// Spec describes how to create a new runtime.
type Spec struct {
	Command []string `json:"command"`
}

// StatusType represents the kind of runtime status.
type StatusType string

// Runtime status constants.
const (
	StatusRunning StatusType = "running"
	StatusExited  StatusType = "exited"
	StatusFailed  StatusType = "failed"
)

// Status represents the current status of a runtime.
type Status struct {
	Type   StatusType `json:"type"`
	Code   int        `json:"code,omitempty"`
	Reason string     `json:"reason,omitempty"`
}

// Running returns a Status indicating the runtime is running.
func Running() Status {
	return Status{Type: StatusRunning}
}

// Exited returns a Status indicating the runtime exited with a code.
func Exited(code int) Status {
	return Status{Type: StatusExited, Code: code}
}

// Failed returns a Status indicating the runtime failed.
func Failed(reason string) Status {
	return Status{Type: StatusFailed, Reason: reason}
}

// ── Errors ─────────────────────────────────────────────

var (
	// ErrNotFound indicates the requested runtime does not exist.
	ErrNotFound = errors.New("runtime not found")
	// ErrAlreadyExists indicates a runtime instance with the given ID already exists.
	ErrAlreadyExists = errors.New("runtime already exists")
	// ErrRuntime indicates a runtime error in runtime operations.
	ErrRuntime = errors.New("runtime error")
	// ErrInvalidID indicates the string is not a valid UUID.
	ErrInvalidID = errors.New("invalid runtime ID: must be a valid UUID")
)

// Error provides context for runtime operation failures.
type Error struct {
	Op  string
	ID  ID
	Err error
}

func (e *Error) Error() string {
	if !e.ID.IsZero() {
		return fmt.Sprintf("runtime %s %s: %v", e.Op, e.ID, e.Err)
	}
	return fmt.Sprintf("runtime %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// ── Interface ──────────────────────────────────────────

// Runtime abstracts runtime lifecycle management.
// Implementations handle the actual creation, destruction, and status
// querying of runtime processes (e.g., local process, Docker, K8s).
type Runtime interface {
	Create(ctx context.Context, spec Spec) (ID, error)
	Destroy(ctx context.Context, id ID) error
	Status(ctx context.Context, id ID) (Status, error)
}
