package application
import (
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"strings"
	"sync"
	"time"
)
type RollbackService struct {
	store   repository.Store
	enqueue func(string, string, string) (domain.ComparisonJob, error)
	mu      sync.Mutex
}
func NewRollbackService(store repository.Store, enqueue func(string, string, string) (domain.ComparisonJob, error)) *RollbackService {
	return &RollbackService{store: store, enqueue: enqueue}
}
func (s *RollbackService) Rollback(documentID, sourceID, reason, actor, trace string) (domain.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, err := s.store.GetDocument(documentID)
	if err != nil {
		return domain.Version{}, err
	}
	source, err := s.store.GetVersion(sourceID)
	if err != nil {
		return domain.Version{}, err
	}
	if source.DocumentID != documentID || strings.TrimSpace(reason) == "" {
		return domain.Version{}, domain.ErrInvalid
	}
	current, err := s.store.GetVersion(doc.CurrentVersionID)
	if err != nil {
		return domain.Version{}, err
	}
	now := time.Now().UTC()
	created := domain.Version{ID: id("ver_"), DocumentID: documentID, Number: current.Number + 1, Content: source.Content, ContentHash: source.ContentHash, Message: "rollback: " + strings.TrimSpace(reason), CreatedBy: actor, ParentID: current.ID, RollbackFromID: source.ID, CreatedAt: now}
	doc.CurrentVersionID = created.ID
	doc.UpdatedAt = now
	paragraphs := domain.SplitParagraphs(created.ID, created.Content)
	if committer, ok := s.store.(interface {
		CommitVersion(domain.Document, domain.Version, []domain.Paragraph) error
	}); ok {
		if err = committer.CommitVersion(doc, created, paragraphs); err != nil {
			return domain.Version{}, err
		}
	} else {
		if err = s.store.CreateVersion(created, paragraphs); err != nil {
			return domain.Version{}, err
		}
		if err = s.store.UpdateDocument(doc); err != nil {
			return domain.Version{}, err
		}
	}
	record := domain.RollbackRecord{ID: id("rollback_"), DocumentID: documentID, SourceVersionID: source.ID, CreatedVersionID: created.ID, Reason: reason, ActorID: actor, CreatedAt: now}
	_ = s.store.CreateRollback(record)
	audit(s.store, actor, "document.rolled_back", "document", documentID, trace, map[string]string{"source": source.ID, "created": created.ID})
	_, _ = s.enqueue(current.ID, created.ID, documentID)
	return created, nil
}
