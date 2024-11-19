package process

import "context"

// Process represents a component that can process data
type Process interface {
	// Name returns the process name
	Name() string

	// Process starts the main processing loop
	Process(ctx context.Context) error

	// Initialize prepares the process for execution
	Initialize(ctx context.Context) error

	// Shutdown cleans up process resources
	Shutdown(ctx context.Context) error

	// IsInitialized returns whether the process has been initialized
	IsInitialized() bool
}
