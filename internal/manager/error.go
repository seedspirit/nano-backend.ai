package manager

import "errors"

// Manager-level errors for startup and operation.
var (
	// ErrBind indicates the server failed to bind to the address.
	ErrBind = errors.New("failed to bind address")
	// ErrServe indicates the server encountered an error while serving.
	ErrServe = errors.New("server error")
)
