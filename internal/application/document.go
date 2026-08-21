package application
import (
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"strings"
	"sync"
	"time"
)
type DocumentService struct {
	store   repository.Store
	enqueue func(string, string, string) (domain.ComparisonJob, error)
	mu      sync.Mutex
}
func NewDocumentService(store repository.Store, enqueue func(string, string, string) (domain.ComparisonJob, error)) *DocumentService {
	return &DocumentService{store: store, enqueue: enqueue}
}
type CreateDocumentInput struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Content string `json:"content"`
	ActorID string `json:"-"`
	TraceID string `json:"-"`
}
func (s *DocumentService) Create(input CreateDocumentInput) (domain.Document, domain.Version, error) {
	input.Title, input.Content = strings.TrimSpace(input.Title), domain.Normalize(input.Content)
	if input.Title == "" || input.Content == "" {
		return domain.Document{}, domain.Version{}, domain.ErrInvalid
	}
	now := time.Now().UTC()
	docID := id("doc_")
	versionID := id("ver_")
	version := domain.Version{ID: versionID, DocumentID: docID, Number: 1, Content: input.Content, ContentHash: domain.Hash(input.Content), Message: "initial version", CreatedBy: input.ActorID, CreatedAt: now}
	doc := domain.Document{ID: docID, Title: input.Title, Summary: strings.TrimSpace(input.Summary), CreatedBy: input.ActorID, CurrentVersionID: versionID, CreatedAt: now, UpdatedAt: now}
	if err := s.store.CreateDocument(doc, version, domain.SplitParagraphs(versionID, input.Content)); err != nil {
		return domain.Document{}, domain.Version{}, err
	}
	audit(s.store, input.ActorID, "document.created", "document", doc.ID, input.TraceID, map[string]string{"version": version.ID})
	return doc, version, nil
}
type CreateVersionInput struct {
	Content         string `json:"content"`
	Message         string `json:"message"`
	ExpectedVersion int    `json:"expected_version"`
	ActorID         string `json:"-"`
	TraceID         string `json:"-"`
}
func (s *DocumentService) CreateVersion(documentID string, input CreateVersionInput) (domain.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, err := s.store.GetDocument(documentID)
	if err != nil {
		return domain.Version{}, err
	}
	current, err := s.store.GetVersion(doc.CurrentVersionID)
	if err != nil {
		return domain.Version{}, err
	}
	input.Content = domain.Normalize(input.Content)
	if input.Content == "" || input.ExpectedVersion != current.Number {
		return domain.Version{}, domain.ErrConflict
	}
	if domain.Hash(input.Content) == current.ContentHash {
		return domain.Version{}, domain.ErrConflict
	}
	now := time.Now().UTC()
	versionID := id("ver_")
	version := domain.Version{ID: versionID, DocumentID: documentID, Number: current.Number + 1, Content: input.Content, ContentHash: domain.Hash(input.Content), Message: strings.TrimSpace(input.Message), CreatedBy: input.ActorID, ParentID: current.ID, CreatedAt: now}
	doc.CurrentVersionID = version.ID
	doc.UpdatedAt = now
	paragraphs := domain.SplitParagraphs(versionID, input.Content)
	if committer, ok := s.store.(interface {
		CommitVersion(domain.Document, domain.Version, []domain.Paragraph) error
	}); ok {
		if err = committer.CommitVersion(doc, version, paragraphs); err != nil {
			return domain.Version{}, err
		}
	} else {
		if err = s.store.CreateVersion(version, paragraphs); err != nil {
			return domain.Version{}, err
		}
		if err = s.store.UpdateDocument(doc); err != nil {
			return domain.Version{}, err
		}
	}
	audit(s.store, input.ActorID, "version.created", "document", documentID, input.TraceID, map[string]string{"version": version.ID})
	_, _ = s.enqueue(current.ID, version.ID, documentID)
	return version, nil
}
func (s *DocumentService) Get(id string) (domain.Document, error) { return s.store.GetDocument(id) }
func (s *DocumentService) List() []domain.Document                { return s.store.ListDocuments() }
func (s *DocumentService) Versions(id string) ([]domain.Version, error) {
	if _, err := s.store.GetDocument(id); err != nil {
		return nil, err
	}
	return s.store.ListVersions(id), nil
}
func (s *DocumentService) Version(id string) (domain.Version, []domain.Paragraph, error) {
	v, e := s.store.GetVersion(id)
	paragraphs := s.store.Paragraphs(id)
	if len(paragraphs) > 0 {
		paragraphs[0].Content = strings.ToUpper(paragraphs[0].Content)
	}
	return v, paragraphs, e
}
