package debug

import (
	"context"
	"log"

	"github.com/elleshadow/noPromises/pkg/core/ip"
	"github.com/elleshadow/noPromises/pkg/nodes"
)

// Logger logs incoming data packets
type Logger[T any] struct {
	*nodes.BaseNode[T, T]
	LogPrefix string
}

// NewLogger creates a new logger node
func NewLogger[T any](prefix string) *Logger[T] {
	return &Logger[T]{
		BaseNode:  nodes.NewBaseNode[T, T]("Logger"),
		LogPrefix: prefix,
	}
}

// Process implements the processing logic
func (l *Logger[T]) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packet, err := l.InPort.Receive(ctx)
			if err != nil {
				return err
			}

			log.Printf("%s: %v", l.LogPrefix, packet.Data())

			if err := l.OutPort.Send(ctx, ip.New(packet.Data())); err != nil {
				return err
			}
		}
	}
}
