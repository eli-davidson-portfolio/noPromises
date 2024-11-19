package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/elleshadow/noPromises/pkg/core/process"
)

// Network represents a collection of connected processes
type Network struct {
	processes map[string]process.Process
	mu        sync.RWMutex
}

// New creates a new empty network
func New() *Network {
	return &Network{
		processes: make(map[string]process.Process),
	}
}

// AddProcess adds a process to the network
func (n *Network) AddProcess(p process.Process) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.processes[p.Name()] = p
}

// GetProcess retrieves a process by name
func (n *Network) GetProcess(name string) process.Process {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.processes[name]
}

// ProcessCount returns the number of processes in the network
func (n *Network) ProcessCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.processes)
}

// Start starts all processes in the network
func (n *Network) Start(ctx context.Context) error {
	n.mu.RLock()
	processes := make([]process.Process, 0, len(n.processes))
	for _, p := range n.processes {
		processes = append(processes, p)
	}
	n.mu.RUnlock()

	// Initialize all processes
	for _, p := range processes {
		if err := p.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize process %s: %w", p.Name(), err)
		}
	}

	// Start all processes
	errCh := make(chan error, len(processes))
	var wg sync.WaitGroup

	for _, p := range processes {
		wg.Add(1)
		go func(p process.Process) {
			defer wg.Done()
			if err := p.Process(ctx); err != nil && err != context.Canceled {
				errCh <- fmt.Errorf("process %s failed: %w", p.Name(), err)
			}
		}(p)
	}

	// Wait for completion or error
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return first error if any
	for err := range errCh {
		return err
	}

	return nil
}

// Stop stops all processes in the network
func (n *Network) Stop(ctx context.Context) error {
	n.mu.RLock()
	processes := make([]process.Process, 0, len(n.processes))
	for _, p := range n.processes {
		processes = append(processes, p)
	}
	n.mu.RUnlock()

	var lastErr error
	for _, p := range processes {
		if err := p.Shutdown(ctx); err != nil {
			lastErr = fmt.Errorf("failed to stop process %s: %w", p.Name(), err)
		}
	}
	return lastErr
}
