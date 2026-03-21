package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ── Types ──────────────────────────────────────────────

// KernelID is a unique identifier for a kernel instance.
type KernelID string

// NewKernelID generates a new random KernelID.
func NewKernelID() KernelID {
	return KernelID(uuid.New().String())
}

// String returns the string representation of a KernelID.
func (id KernelID) String() string {
	return string(id)
}

// KernelSpec describes how to create a new kernel.
type KernelSpec struct {
	Command []string `json:"command"`
}

// KernelStatusType represents the kind of kernel status.
type KernelStatusType int

const (
	// StatusRunning indicates the kernel is currently running.
	StatusRunning KernelStatusType = iota
	// StatusExited indicates the kernel exited with a code.
	StatusExited
	// StatusFailed indicates the kernel failed with a reason.
	StatusFailed
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
)

// KernelError provides context for kernel operation failures.
type KernelError struct {
	Op  string
	ID  KernelID
	Err error
}

func (e *KernelError) Error() string {
	if e.ID != "" {
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
