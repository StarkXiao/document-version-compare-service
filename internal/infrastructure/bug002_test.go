package infrastructure

import (
	"document-version-compare-service/internal/domain"
	"testing"
)

func TestBug002CommentIndexNilMap(t *testing.T) {
	s := NewMemoryStore()
	defer func() {
		if recover() != nil {
			t.Fatal("comment creation panicked")
		}
	}()
	if err := s.CreateComment(domain.Comment{ID: "c1", DocumentID: "d1"}); err != nil { t.Fatal(err) }
}
