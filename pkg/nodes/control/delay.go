package control

import (
	"context"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/nodes"
)

type Delay[T any] struct {
	*nodes.BaseNode[T, T]
	Duration time.Duration
}

func NewDelay[T any](duration time.Duration) *Delay[T] {
	return &Delay[T]{
		BaseNode: nodes.NewBaseNode[T, T]("Delay"),
		Duration: duration,
	}
}

func (d *Delay[T]) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := d.InPort.Receive(ctx)
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d.Duration):
				if err := d.OutPort.Send(ctx, ip.New(packet.Data())); err != nil {
					return err
				}
			}
		}
	}
}
