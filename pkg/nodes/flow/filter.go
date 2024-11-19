package flow

import (
	"context"
	"fmt"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/nodes"
)

// Filter filters packets based on a predicate function
type Filter[T any] struct {
	*nodes.BaseNode[T, T]
	Predicate func(T) bool
}

// NewFilter creates a new filter node
func NewFilter[T any](predicate func(T) bool) *Filter[T] {
	return &Filter[T]{
		BaseNode:  nodes.NewBaseNode[T, T]("Filter"),
		Predicate: predicate,
	}
}

// Process implements the processing logic
func (f *Filter[T]) Process(ctx context.Context) error {
	if f.Predicate == nil {
		return fmt.Errorf("nil predicate function")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := f.InPort.Receive(ctx)
			if err != nil {
				return err
			}

			if f.Predicate(packet.Data()) {
				if err := f.OutPort.Send(ctx, ip.New(packet.Data())); err != nil {
					return err
				}
			}
		}
	}
}
