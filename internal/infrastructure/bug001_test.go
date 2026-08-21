package infrastructure

import (
	"document-version-compare-service/internal/domain"
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
			_ = s.CreateComment(domain.Comment{ID: "c" + itoa(i), DocumentID: "d"})
		}(i)
	}
	wg.Wait()
}
