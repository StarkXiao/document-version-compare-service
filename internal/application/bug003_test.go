package application

import (
	"document-version-compare-service/internal/infrastructure"
	"testing"
)

func TestBug003VersionReadMutatesStoredParagraph(t *testing.T) {
	s := infrastructure.NewMemoryStore()
	services := NewServices(s, func(string) bool { return true })
	_, v, err := services.Documents.Create(CreateDocumentInput{Title: "doc", Content: "Mixed Case"})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = services.Documents.Version(v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Paragraphs(v.ID)[0].Content; got != "Mixed Case" {
		t.Fatalf("stored paragraph = %q", got)
	}
}
