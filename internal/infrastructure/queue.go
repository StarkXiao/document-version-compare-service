package infrastructure
import (
	"context"
	"sync"
)
type Queue struct {
	jobs chan string
	once sync.Once
}
func NewQueue(size int) *Queue { return &Queue{jobs: make(chan string, size)} }
func (q *Queue) Enqueue(id string) bool {
	select {
	case q.jobs <- id:
		return true
	default:
		return false
	}
}
func (q *Queue) Run(ctx context.Context, workers int, handle func(context.Context, string)) {
	q.once.Do(func() {
		for i := 0; i < workers; i++ {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case id := <-q.jobs:
						handle(ctx, id)
					}
				}
			}()
		}
	})
}
