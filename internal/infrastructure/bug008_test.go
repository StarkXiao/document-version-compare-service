package infrastructure

import (
	"context"
	"testing"
	"time"
)

func TestBug008ClosedQueueDoesNotDispatchEmptyJobs(t *testing.T) {
	q := NewQueue(1)
	q.Close()
	called := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.Run(ctx, 1, func(_ context.Context, id string) { called <- id })
	select {
	case id := <-called:
		t.Fatalf("empty job dispatched: %q", id)
	case <-time.After(20 * time.Millisecond):
	}
}
