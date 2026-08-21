package infrastructure

import (
	"document-version-compare-service/internal/domain"
	"testing"
)

func TestBug002CommentIndexNilMap(t *testing.T) {
	s := NewMemoryStore()
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil map panic")
		}
	}()
	_ = s.CreateComment(domain.Comment{ID: "c1", DocumentID: "d1"})
}
