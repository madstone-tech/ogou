package engine

import "context"

// Driver is the transport implementation.
type Driver interface {
	// Name returns a human-readable protocol name, e.g. "http".
	Name() string

	// Execute runs a single Step within a StepContext.
	Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error)

	// Close tears down any persistent connections.
	Close() error
}
