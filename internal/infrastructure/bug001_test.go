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
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.CreateComment(domain.Comment{ID: fmt.Sprintf("c%d", i), DocumentID: "d"})
		}(i)
	}
	wg.Wait()
}
