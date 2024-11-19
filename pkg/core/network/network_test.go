package network

import (
	"context"
	"testing"
	"time"

	"github.com/elleshadow/noPromises/pkg/core/process"
)

type testProcess struct {
	process.BaseProcess
	processed chan struct{}
}

func newTestProcess(name string) *testProcess {
	return &testProcess{
		BaseProcess: process.NewBaseProcess(name),
		processed:   make(chan struct{}, 1),
	}
}

func (p *testProcess) Process(ctx context.Context) error {
	p.processed <- struct{}{}
	<-ctx.Done()
	return ctx.Err()
}

func TestNetwork(t *testing.T) {
	n := New()
	p1 := newTestProcess("p1")
	p2 := newTestProcess("p2")

	n.AddProcess(p1)
	n.AddProcess(p2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		if err := n.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	select {
	case <-p1.processed:
		select {
		case <-p2.processed:
			cancel()
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for p2")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for p1")
	}
}
