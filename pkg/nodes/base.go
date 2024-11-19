package nodes

import (
	"context"
	"sync"

	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/elleshadow/noPromises/pkg/core/process"
)

// BaseNode provides common functionality for all nodes
type BaseNode[In, Out any] struct {
	process.BaseProcess
	InPort  *ports.Port[In]
	OutPort *ports.Port[Out]
	Config  map[string]interface{}
	mu      sync.RWMutex
}

// NewBaseNode creates a new base node with the given name
func NewBaseNode[In, Out any](name string) *BaseNode[In, Out] {
	return &BaseNode[In, Out]{
		BaseProcess: process.NewBaseProcess(name),
		InPort:      ports.NewInput[In]("in", "Input port", true),
		OutPort:     ports.NewOutput[Out]("out", "Output port", true),
		Config:      make(map[string]interface{}),
	}
}

// Initialize prepares the node for execution
func (n *BaseNode[In, Out]) Initialize(ctx context.Context) error {
	return n.BaseProcess.Initialize(ctx)
}

// Process implements the process.Process interface
func (n *BaseNode[In, Out]) Process(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown cleans up node resources
func (n *BaseNode[In, Out]) Shutdown(ctx context.Context) error {
	return n.BaseProcess.Shutdown(ctx)
}

// GetConfig returns the node configuration
func (n *BaseNode[In, Out]) GetConfig() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Config
}

// SetConfig updates the node configuration
func (n *BaseNode[In, Out]) SetConfig(config map[string]interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Config = config
}
