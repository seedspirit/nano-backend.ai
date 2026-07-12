package session

import (
	"errors"
	"time"
)

// Status represents the lifecycle stage of a Session.
//
//	pending → preparing → running → terminated
type Status string

// Session status constants.
const (
	Pending    Status = "pending"
	Preparing  Status = "preparing"
	Running    Status = "running"
	Terminated Status = "terminated"
)

// Result records the outcome independently from lifecycle status.
type Result string

const (
	// ResultUndefined indicates that a session has not reached a terminal outcome.
	ResultUndefined Result = "undefined"
	// ResultSuccess indicates successful execution.
	ResultSuccess Result = "success"
	// ResultFailure indicates unsuccessful execution.
	ResultFailure Result = "failure"
)

// FailureReason is the machine-readable reason a Session failed.
type FailureReason string

var (
	// ErrInvalidTransition indicates a status transition is not allowed.
	ErrInvalidTransition = errors.New("invalid session status transition")
	// ErrFailureReasonRequired indicates a failed transition lacks a reason.
	ErrFailureReasonRequired = errors.New("failure reason is required")
)

// Transition describes a requested lifecycle transition.
type Transition struct {
	next          Status
	result        Result
	failureReason *FailureReason
}

// Next creates an ordinary lifecycle transition.
func Next(status Status) Transition {
	return Transition{next: status}
}

// Fail creates a failed lifecycle transition with a machine-readable reason.
func Fail(reason FailureReason) Transition {
	return Transition{
		next:          Terminated,
		result:        ResultFailure,
		failureReason: &reason,
	}
}

// Succeed creates a successful terminal transition.
func Succeed() Transition { return Transition{next: Terminated, result: ResultSuccess} }

// Lifecycle groups mutable execution state for a Session.
type Lifecycle struct {
	Status        Status         `json:"status"`
	Result        Result         `json:"result"`
	FailureReason *FailureReason `json:"failure_reason,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	FinishedAt    *time.Time     `json:"finished_at,omitempty"`
}

// NewLifecycle creates a pending lifecycle with the given creation time.
func NewLifecycle(createdAt time.Time) Lifecycle {
	return Lifecycle{
		Status:    Pending,
		Result:    ResultUndefined,
		CreatedAt: createdAt,
	}
}

// CanTransitionTo reports whether moving from s to next is a valid state
func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case Pending:
		return next == Preparing
	case Preparing:
		return next == Running || next == Terminated
	case Running:
		return next == Terminated
	case Terminated:
		return false
	default:
		return false
	}
}

// Transition applies the requested lifecycle transition.
func (l *Lifecycle) Transition(transition Transition, at time.Time) error {
	next := transition.next
	if !l.Status.CanTransitionTo(next) {
		return ErrInvalidTransition
	}
	if next == Terminated && transition.result == ResultFailure {
		if transition.failureReason == nil || *transition.failureReason == "" {
			return ErrFailureReasonRequired
		}
		reasonCopy := *transition.failureReason
		l.FailureReason = &reasonCopy
	} else {
		l.FailureReason = nil
	}
	if next == Terminated {
		if transition.result != ResultSuccess && transition.result != ResultFailure {
			return ErrInvalidTransition
		}
		l.Result = transition.result
	}

	l.Status = next
	switch next {
	case Running:
		if l.StartedAt == nil {
			startedAt := at
			l.StartedAt = &startedAt
		}
	case Terminated:
		finishedAt := at
		l.FinishedAt = &finishedAt
	}

	return nil
}
