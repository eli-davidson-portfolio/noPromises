package process

import (
	"context"
	"sync"
)

// BaseProcess provides common functionality for all processes
type BaseProcess struct {
	name         string
	mu           sync.RWMutex
	initialized  bool
	shutdownOnce sync.Once
	isShutdown   bool
}

// NewBaseProcess creates a new base process with the given name
func NewBaseProcess(name string) BaseProcess {
	return BaseProcess{
		name:        name,
		initialized: false,
		isShutdown:  false,
	}
}

// Name returns the process name
func (p *BaseProcess) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

// Process implements the basic process behavior
func (p *BaseProcess) Process(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Initialize prepares the process for execution
func (p *BaseProcess) Initialize(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isShutdown {
		return ErrProcessShutdown
	}

	p.initialized = true
	return nil
}

// Shutdown cleans up process resources
func (p *BaseProcess) Shutdown(_ context.Context) error {
	p.shutdownOnce.Do(func() {
		p.mu.Lock()
		p.initialized = false
		p.isShutdown = true
		p.mu.Unlock()
	})
	return nil
}

// IsInitialized returns whether the process has been initialized
func (p *BaseProcess) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized && !p.isShutdown
}
