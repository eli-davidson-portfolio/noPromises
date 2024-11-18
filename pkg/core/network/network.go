package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/core/ports"
	"github.com/elleshadow/noPromises/pkg/core/process"
)

type Network struct {
	processes map[string]process.Process
	mu        sync.RWMutex
}

func New() *Network {
	return &Network{
		processes: make(map[string]process.Process),
	}
}

func (n *Network) ProcessCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.processes)
}

func (n *Network) AddProcess(id string, p process.Process) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.processes[id]; exists {
		return fmt.Errorf("process with id %s already exists", id)
	}

	n.processes[id] = p
	return nil
}

func (n *Network) Connect(fromID, fromPort, toID, toPort string) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	fromProc, exists := n.processes[fromID]
	if !exists {
		return fmt.Errorf("process %s not found", fromID)
	}

	toProc, exists := n.processes[toID]
	if !exists {
		return fmt.Errorf("process %s not found", toID)
	}

	// Type assertion to get the actual ports
	fromP, ok := fromProc.(interface{ GetPort(string) interface{} })
	if !ok {
		return fmt.Errorf("process %s does not support port access", fromID)
	}

	toP, ok := toProc.(interface{ GetPort(string) interface{} })
	if !ok {
		return fmt.Errorf("process %s does not support port access", toID)
	}

	outPort := fromP.GetPort(fromPort)
	if outPort == nil {
		return fmt.Errorf("port %s not found on process %s", fromPort, fromID)
	}

	inPort := toP.GetPort(toPort)
	if inPort == nil {
		return fmt.Errorf("port %s not found on process %s", toPort, toID)
	}

	// Type assert the ports and connect them
	outPortTyped, ok := outPort.(*ports.Port[string])
	if !ok {
		return fmt.Errorf("invalid port type for %s on process %s", fromPort, fromID)
	}

	inPortTyped, ok := inPort.(*ports.Port[string])
	if !ok {
		return fmt.Errorf("invalid port type for %s on process %s", toPort, toID)
	}

	// Create channel and connect ports
	ch := make(chan *ip.IP[string], 1) // Buffer size of 1 for basic flow control
	if err := outPortTyped.Connect(ch); err != nil {
		return fmt.Errorf("connecting output port: %w", err)
	}
	if err := inPortTyped.Connect(ch); err != nil {
		return fmt.Errorf("connecting input port: %w", err)
	}

	return nil
}

func (n *Network) Run(ctx context.Context) error {
	n.mu.RLock()
	processes := make([]process.Process, 0, len(n.processes))
	for _, p := range n.processes {
		processes = append(processes, p)
	}
	n.mu.RUnlock()

	// Initialize all processes
	for _, p := range processes {
		if err := p.Initialize(ctx); err != nil {
			// Clean up any already initialized processes
			for _, cleanup := range processes {
				if cleanup.IsInitialized() {
					_ = cleanup.Shutdown(ctx)
				}
			}
			return fmt.Errorf("initializing process: %w", err)
		}
	}

	// Create error channel for collecting process errors
	errCh := make(chan error, len(processes))

	// Start all processes
	var wg sync.WaitGroup
	for _, p := range processes {
		wg.Add(1)
		go func(p process.Process) {
			defer wg.Done()
			if err := p.Process(ctx); err != nil {
				errCh <- err
			}
		}(p)
	}

	// Wait for completion or error
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(errCh)
	}()

	// Wait for either context cancellation or process error
	var runErr error
	select {
	case <-ctx.Done():
		runErr = ctx.Err()
	case err := <-errCh:
		if err != nil {
			runErr = err
		}
	case <-done:
		// Normal completion
	}

	// Always attempt to shut down all processes
	for _, p := range processes {
		if err := p.Shutdown(ctx); err != nil && runErr == nil {
			runErr = fmt.Errorf("shutting down process: %w", err)
		}
	}

	return runErr
}
