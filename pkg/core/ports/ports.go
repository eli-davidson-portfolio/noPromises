package ports

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/elleshadow/noPromises/pkg/core/ip"
)

type PortType int

const (
	TypeInput PortType = iota
	TypeOutput
)

type Port[T any] struct {
	name           string
	description    string
	required       bool
	portType       PortType
	channels       []chan *ip.IP[T]
	maxConnections int
	mu             sync.RWMutex
}

func NewInput[T any](name, description string, required bool) *Port[T] {
	return &Port[T]{
		name:        name,
		description: description,
		required:    required,
		portType:    TypeInput,
		channels:    make([]chan *ip.IP[T], 0),
	}
}

func NewOutput[T any](name, description string, required bool) *Port[T] {
	return &Port[T]{
		name:        name,
		description: description,
		required:    required,
		portType:    TypeOutput,
		channels:    make([]chan *ip.IP[T], 0),
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
	p.maxConnections = max
}

func Connect[T any](port *Port[T], ch chan *ip.IP[T]) error {
	if port == nil {
		return fmt.Errorf("nil port")
	}
	if ch == nil {
		return fmt.Errorf("nil channel")
	}

	port.mu.Lock()
	defer port.mu.Unlock()

	if port.maxConnections > 0 && len(port.channels) >= port.maxConnections {
		return fmt.Errorf("maximum connections reached")
	}

	port.channels = append(port.channels, ch)
	return nil
}

func (p *Port[T]) Send(ctx context.Context, packet *ip.IP[T]) error {
	p.mu.RLock()
	channels := make([]chan *ip.IP[T], len(p.channels))
	copy(channels, p.channels)
	p.mu.RUnlock()

	for _, ch := range channels {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- packet:
		}
	}
	return nil
}

func (p *Port[T]) Receive(ctx context.Context) (*ip.IP[T], error) {
	p.mu.RLock()
	channels := make([]chan *ip.IP[T], len(p.channels))
	copy(channels, p.channels)
	p.mu.RUnlock()

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels connected")
	}

	// Create cases for select
	cases := make([]reflect.SelectCase, len(channels)+1)
	cases[0] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	for i, ch := range channels {
		cases[i+1] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}

	// Wait for data or context cancellation
	chosen, value, ok := reflect.Select(cases)
	if chosen == 0 { // Context done
		return nil, ctx.Err()
	}
	if !ok {
		return nil, fmt.Errorf("channel closed")
	}

	packet, ok := value.Interface().(*ip.IP[T])
	if !ok {
		return nil, fmt.Errorf("invalid packet type")
	}
	return packet, nil
}
