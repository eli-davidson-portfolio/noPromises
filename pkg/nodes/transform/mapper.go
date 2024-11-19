package transform

import (
	"context"
	"fmt"
	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/nodes"
)

type Mapper[In, Out any] struct {
	*nodes.BaseNode[In, Out]
	Transform func(In) Out
}

func NewMapper[In, Out any](transform func(In) Out) *Mapper[In, Out] {
	return &Mapper[In, Out]{
		BaseNode:  nodes.NewBaseNode[In, Out]("Mapper"),
		Transform: transform,
	}
}

func (m *Mapper[In, Out]) Process(ctx context.Context) error {
	if m.Transform == nil {
		return fmt.Errorf("nil transform function")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := m.InPort.Receive(ctx)
			if err != nil {
				return err
			}

			result := m.Transform(packet.Data())
			if err := m.OutPort.Send(ctx, ip.New(result)); err != nil {
				return err
			}
		}
	}
}
