package engine

import "context"

// Engine defines the interface for the flow execution engine
type Engine interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
