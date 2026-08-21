package infrastructure

import (
	"document-version-compare-service/internal/domain"
	"fmt"
	"sync"
	"testing"
)

func TestBug001ConcurrentCommentWrites(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 1000000; i++ { _ = s.state() }
	}()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.CreateComment(domain.Comment{ID: fmt.Sprintf("c%d", i), DocumentID: "d"})
		}(i)
	}
	close(start)
	wg.Wait()
}
