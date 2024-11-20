package server

import (
	"context"
	"sync"
)

// Process represents a runnable process in the server
type Process interface {
	Start(context.Context) error
	Stop(context.Context) error
	ID() string
}

// ProcessFactory creates new process instances
type ProcessFactory interface {
	Create(config map[string]interface{}) (Process, error)
}

// ProcessRegistry manages process type registrations
type ProcessRegistry struct {
	processes map[string]ProcessFactory
	mu        sync.RWMutex
}

func newProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		processes: make(map[string]ProcessFactory),
	}
}

// Register adds a new process factory
func (r *ProcessRegistry) Register(typeName string, factory ProcessFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processes[typeName] = factory
}

// Get retrieves a process factory by type name
func (r *ProcessRegistry) Get(typeName string) (ProcessFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, exists := r.processes[typeName]
	return factory, exists
}
