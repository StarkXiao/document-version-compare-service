package application

import (
	"document-version-compare-service/internal/infrastructure"
	"testing"
)

func TestCreateVersionCreatesComparisonJob(t *testing.T) {
	store := infrastructure.NewMemoryStore()
	services := NewServices(store, func(string) bool { return true })
	doc, first, err := services.Documents.Create(CreateDocumentInput{Title: "设计", Content: "旧内容", ActorID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := services.Documents.CreateVersion(doc.ID, CreateVersionInput{Content: "新内容", ExpectedVersion: first.Number, ActorID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Number != 2 {
		t.Fatalf("got version %d", second.Number)
	}
	if _, ok := store.FindJob(first.ID, second.ID); !ok {
		t.Fatal("comparison job was not created")
	}
}

func TestRejectsDuplicateVersionContent(t *testing.T) {
	store := infrastructure.NewMemoryStore()
	services := NewServices(store, func(string) bool { return true })
	doc, first, err := services.Documents.Create(CreateDocumentInput{Title: "设计", Content: "内容", ActorID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = services.Documents.CreateVersion(doc.ID, CreateVersionInput{Content: "内容", ExpectedVersion: first.Number, ActorID: "u1"}); err == nil {
		t.Fatal("expected duplicate content error")
	}
}
