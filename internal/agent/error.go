package agent

import "errors"

// Agent-level errors for startup and operation.
var (
	// ErrStartup indicates the agent failed during startup.
	ErrStartup = errors.New("agent startup failed")
	// ErrOperation indicates the agent encountered an operational error.
	ErrOperation = errors.New("agent operation error")
)
