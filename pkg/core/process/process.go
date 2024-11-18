package process

import (
	"context"
	"sync"
)

type BaseProcess struct {
	initialized bool
	mu          sync.RWMutex
}

func (p *BaseProcess) Initialize(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initialized = true
	return nil
}

func (p *BaseProcess) Shutdown(_ context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.initialized = false
	return nil
}

func (p *BaseProcess) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// Process is the interface that must be implemented by all processes
type Process interface {
	Initialize(ctx context.Context) error
	Process(ctx context.Context) error
	Shutdown(ctx context.Context) error
	IsInitialized() bool
}
