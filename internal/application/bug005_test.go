package application

import (
	"context"
	"document-version-compare-service/internal/infrastructure"
	"testing"
)

func TestBug005CancelledComparisonStillWritesResult(t *testing.T) {
	s := infrastructure.NewMemoryStore()
	services := NewServices(s, func(string) bool { return true })
	_, first, err := services.Documents.Create(CreateDocumentInput{Title: "doc", Content: "one"})
	if err != nil { t.Fatal(err) }
	second, err := services.Documents.CreateVersion(first.DocumentID, CreateVersionInput{Content: "two", ExpectedVersion: 1})
	if err != nil { t.Fatal(err) }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	services.Comparisons.SetContext(ctx)
	job, err := services.Comparisons.Enqueue(first.ID, second.ID, first.DocumentID)
	if err != nil { t.Fatal(err) }
	services.Comparisons.Process(ctx, job.ID)
	if _, err := s.GetResult(job.ID); err == nil { t.Fatal("cancelled job produced result") }
}
