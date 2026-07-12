package session

// Type describes how a session is operated, following Backend.AI session types.
type Type string

const (
	// Interactive is a user-driven session such as Jupyter or a development shell.
	Interactive Type = "interactive"
	// Batch is a finite command execution such as training or evaluation.
	Batch Type = "batch"
	// Inference hosts a model-serving runtime.
	Inference Type = "inference"
	// System is reserved for platform-internal computation.
	System Type = "system"
)
