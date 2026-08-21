package application

import (
	"document-version-compare-service/internal/infrastructure"
	"testing"
)

func TestBug010NilEnqueueCallbackPanicsOnVersionCreate(t *testing.T) {
	s := infrastructure.NewMemoryStore()
	services := NewServices(s, nil)
	_, first, err := services.Documents.Create(CreateDocumentInput{Title: "doc", Content: "one"})
	if err != nil { t.Fatal(err) }
	defer func() { if recover() != nil { t.Fatal("nil enqueue callback panicked") } }()
	_, _ = services.Documents.CreateVersion(first.DocumentID, CreateVersionInput{Content: "two", ExpectedVersion: 1})
}
