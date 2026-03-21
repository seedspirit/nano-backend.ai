package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ── Types ──────────────────────────────────────────────

// KernelID is a UUID-based unique identifier for a kernel instance.
// The zero value is invalid; use NewKernelID or ParseKernelID to construct.
type KernelID struct {
	uuid uuid.UUID
}

// NewKernelID generates a new random KernelID.
func NewKernelID() KernelID {
	return KernelID{uuid: uuid.New()}
}

// ParseKernelID parses a string into a KernelID.
// Returns ErrInvalidKernelID if the string is not a valid UUID.
func ParseKernelID(s string) (KernelID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return KernelID{}, ErrInvalidKernelID
	}
	return KernelID{uuid: id}, nil
}

// String returns the UUID string representation.
func (id KernelID) String() string {
	return id.uuid.String()
}

// IsZero reports whether the KernelID is the zero value (uninitialized).
func (id KernelID) IsZero() bool {
	return id.uuid == uuid.Nil
}

// MarshalJSON implements json.Marshaler.
func (id KernelID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.uuid.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *KernelID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return ErrInvalidKernelID
	}
	id.uuid = parsed
	return nil
}

// KernelSpec describes how to create a new kernel.
type KernelSpec struct {
	Command []string `json:"command"`
}

// KernelStatusType represents the kind of kernel status.
type KernelStatusType string

// Kernel status constants.
const (
	StatusRunning KernelStatusType = "running"
	StatusExited  KernelStatusType = "exited"
	StatusFailed  KernelStatusType = "failed"
)

// KernelStatus represents the current status of a kernel.
type KernelStatus struct {
	Type   KernelStatusType `json:"type"`
	Code   int              `json:"code,omitempty"`
	Reason string           `json:"reason,omitempty"`
}

// Running returns a KernelStatus indicating the kernel is running.
func Running() KernelStatus {
	return KernelStatus{Type: StatusRunning}
}

// Exited returns a KernelStatus indicating the kernel exited with a code.
func Exited(code int) KernelStatus {
	return KernelStatus{Type: StatusExited, Code: code}
}

// Failed returns a KernelStatus indicating the kernel failed.
func Failed(reason string) KernelStatus {
	return KernelStatus{Type: StatusFailed, Reason: reason}
}

// ── Errors ─────────────────────────────────────────────

var (
	// ErrKernelNotFound indicates the requested kernel does not exist.
	ErrKernelNotFound = errors.New("kernel not found")
	// ErrKernelAlreadyExists indicates a kernel with the given ID already exists.
	ErrKernelAlreadyExists = errors.New("kernel already exists")
	// ErrKernelRuntime indicates a runtime error in kernel operations.
	ErrKernelRuntime = errors.New("kernel runtime error")
	// ErrInvalidKernelID indicates the string is not a valid UUID.
	ErrInvalidKernelID = errors.New("invalid kernel ID: must be a valid UUID")
)

// KernelError provides context for kernel operation failures.
type KernelError struct {
	Op  string
	ID  KernelID
	Err error
}

func (e *KernelError) Error() string {
	if !e.ID.IsZero() {
		return fmt.Sprintf("kernel %s %s: %v", e.Op, e.ID, e.Err)
	}
	return fmt.Sprintf("kernel %s: %v", e.Op, e.Err)
}

func (e *KernelError) Unwrap() error {
	return e.Err
}

// ── Interface ──────────────────────────────────────────

// KernelRuntime abstracts kernel lifecycle management.
// Implementations handle the actual creation, destruction, and status
// querying of kernel processes (e.g., local process, Docker, K8s).
type KernelRuntime interface {
	Create(ctx context.Context, spec KernelSpec) (KernelID, error)
	Destroy(ctx context.Context, id KernelID) error
	Status(ctx context.Context, id KernelID) (KernelStatus, error)
}
