package ports

import (
	"context"
	"fmt"
	"sync"

	"github.com/elleshadow/noPromises/pkg/core/ip"
)

type PortType int

const (
	TypeInput PortType = iota
	TypeOutput
)

type Port[T any] struct {
	name        string
	description string
	required    bool
	portType    PortType
	channels    []chan *ip.IP[T]
	maxConns    int
	mu          sync.RWMutex
}

func NewInput[T any](name, description string, required bool) *Port[T] {
	return &Port[T]{
		name:        name,
		description: description,
		required:    required,
		portType:    TypeInput,
		channels:    make([]chan *ip.IP[T], 0),
		maxConns:    1, // Default to 1 connection
	}
}

func NewOutput[T any](name, description string, required bool) *Port[T] {
	return &Port[T]{
		name:        name,
		description: description,
		required:    required,
		portType:    TypeOutput,
		channels:    make([]chan *ip.IP[T], 0),
		maxConns:    1, // Default to 1 connection
	}
}

func (p *Port[T]) Name() string {
	return p.name
}

func (p *Port[T]) Description() string {
	return p.description
}

func (p *Port[T]) Required() bool {
	return p.required
}

func (p *Port[T]) Type() PortType {
	return p.portType
}

func (p *Port[T]) SetMaxConnections(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxConns = max
}

func (p *Port[T]) Connect(ch chan *ip.IP[T]) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.channels) >= p.maxConns {
		return fmt.Errorf("port %s: maximum connections (%d) exceeded", p.name, p.maxConns)
	}

	p.channels = append(p.channels, ch)
	return nil
}

func (p *Port[T]) Send(ctx context.Context, packet *ip.IP[T]) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.channels) == 0 {
		return fmt.Errorf("port %s: no connected channels", p.name)
	}

	// For output ports, send to all connected channels
	for _, ch := range p.channels {
		select {
		case ch <- packet:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (p *Port[T]) Receive(ctx context.Context) (*ip.IP[T], error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.channels) == 0 {
		return nil, fmt.Errorf("port %s: no connected channels", p.name)
	}

	// For input ports, receive from first connected channel
	select {
	case packet := <-p.channels[0]:
		return packet, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
