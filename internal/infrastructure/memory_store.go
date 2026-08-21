package infrastructure
import (
	"document-version-compare-service/internal/domain"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)
type MemoryStore struct {
	mu                 sync.RWMutex
	documents          map[string]domain.Document
	versions           map[string]domain.Version
	byDocument         map[string][]string
	paragraphs         map[string][]domain.Paragraph
	jobs               map[string]domain.ComparisonJob
	results            map[string]domain.ComparisonResult
	comments           map[string]domain.Comment
	commentsByDocument map[string][]string
	replies            map[string][]domain.CommentReply
	audits             map[string][]domain.AuditLog
	rollbacks          []domain.RollbackRecord
	path               string
	loadErr            error
}
func NewMemoryStore(paths ...string) *MemoryStore {
	s := &MemoryStore{documents: map[string]domain.Document{}, versions: map[string]domain.Version{}, byDocument: map[string][]string{}, paragraphs: map[string][]domain.Paragraph{}, jobs: map[string]domain.ComparisonJob{}, results: map[string]domain.ComparisonResult{}, comments: map[string]domain.Comment{}, commentsByDocument: map[string][]string{}, replies: map[string][]domain.CommentReply{}, audits: map[string][]domain.AuditLog{}}
	if len(paths) > 0 {
		s.path = paths[0]
		s.loadErr = s.load()
	}
	return s
}
func (s *MemoryStore) LoadError() error { return s.loadErr }
type snapshot struct {
	Documents          map[string]domain.Document
	Versions           map[string]domain.Version
	ByDocument         map[string][]string
	Paragraphs         map[string][]domain.Paragraph
	Jobs               map[string]domain.ComparisonJob
	Results            map[string]domain.ComparisonResult
	Comments           map[string]domain.Comment
	CommentsByDocument map[string][]string
	Replies            map[string][]domain.CommentReply
	Audits             map[string][]domain.AuditLog
	Rollbacks          []domain.RollbackRecord
}
func (s *MemoryStore) state() snapshot {
	return snapshot{s.documents, s.versions, s.byDocument, s.paragraphs, s.jobs, s.results, s.comments, s.commentsByDocument, s.replies, s.audits, s.rollbacks}
}
func (s *MemoryStore) load() error {
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var v snapshot
	if err = json.Unmarshal(raw, &v); err != nil {
		return err
	}
	if v.Documents != nil {
		s.documents = v.Documents
	}
	if v.Versions != nil {
		s.versions = v.Versions
	}
	if v.ByDocument != nil {
		s.byDocument = v.ByDocument
	}
	if v.Paragraphs != nil {
		s.paragraphs = v.Paragraphs
	}
	if v.Jobs != nil {
		s.jobs = v.Jobs
	}
	if v.Results != nil {
		s.results = v.Results
	}
	if v.Comments != nil {
		s.comments = v.Comments
	}
	if v.CommentsByDocument != nil {
		s.commentsByDocument = v.CommentsByDocument
	}
	if v.Replies != nil {
		s.replies = v.Replies
	}
	if v.Audits != nil {
		s.audits = v.Audits
	}
	s.rollbacks = v.Rollbacks
	return nil
}
func (s *MemoryStore) persist() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	raw, err := json.Marshal(s.state())
	if err != nil {
		return err
	}
	temp := s.path + ".tmp"
	if err = os.WriteFile(temp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(temp, s.path)
}
func (s *MemoryStore) CreateDocument(doc domain.Document, version domain.Version, paragraphs []domain.Paragraph) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.documents[doc.ID]; exists {
		return domain.ErrConflict
	}
	s.documents[doc.ID], s.versions[version.ID], s.paragraphs[version.ID] = doc, version, paragraphs
	s.byDocument[doc.ID] = []string{version.ID}
	return s.persist()
}
func (s *MemoryStore) GetDocument(id string) (domain.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.documents[id]
	if !ok {
		return domain.Document{}, domain.ErrNotFound
	}
	return item, nil
}
func (s *MemoryStore) ListDocuments() []domain.Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Document, 0, len(s.documents))
	for _, d := range s.documents {
		out = append(out, d)
	}
	return out
}
func (s *MemoryStore) CreateVersion(version domain.Version, paragraphs []domain.Paragraph) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.versions[version.ID]; exists {
		return domain.ErrConflict
	}
	if _, exists := s.documents[version.DocumentID]; !exists {
		return domain.ErrNotFound
	}
	for _, id := range s.byDocument[version.DocumentID] {
		if s.versions[id].Number == version.Number {
			return domain.ErrConflict
		}
	}
	s.versions[version.ID], s.paragraphs[version.ID] = version, paragraphs
	s.byDocument[version.DocumentID] = append(s.byDocument[version.DocumentID], version.ID)
	return s.persist()
}
func (s *MemoryStore) CommitVersion(doc domain.Document, version domain.Version, paragraphs []domain.Paragraph) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[doc.ID]; !ok {
		return domain.ErrNotFound
	}
	for _, id := range s.byDocument[doc.ID] {
		if s.versions[id].Number == version.Number {
			return domain.ErrConflict
		}
	}
	s.versions[version.ID], s.paragraphs[version.ID] = version, paragraphs
	s.byDocument[doc.ID] = append(s.byDocument[doc.ID], version.ID)
	s.documents[doc.ID] = doc
	return s.persist()
}
func (s *MemoryStore) GetVersion(id string) (domain.Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.versions[id]
	if !ok {
		return domain.Version{}, domain.ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) ListVersions(docID string) []domain.Version {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.byDocument[docID]
	out := make([]domain.Version, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.versions[id])
	}
	return out
}
func (s *MemoryStore) Paragraphs(id string) []domain.Paragraph {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paragraphs[id]
}
func (s *MemoryStore) UpdateDocument(doc domain.Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.documents[doc.ID]; !ok {
		return domain.ErrNotFound
	}
	s.documents[doc.ID] = doc
	return s.persist()
}
func (s *MemoryStore) CreateJob(job domain.ComparisonJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[job.ID]; ok {
		return domain.ErrConflict
	}
	s.jobs[job.ID] = job
	return s.persist()
}
func (s *MemoryStore) GetJob(id string) (domain.ComparisonJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.jobs[id]
	if !ok {
		return domain.ComparisonJob{}, domain.ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) ClaimJob(id string) (domain.ComparisonJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok || job.Status != domain.JobPending {
		return domain.ComparisonJob{}, false
	}
	now := time.Now().UTC()
	job.Status = domain.JobRunning
	job.Attempts++
	job.StartedAt = &now
	s.jobs[id] = job
	_ = s.persist()
	return job, true
}
func (s *MemoryStore) FindJob(base, target string) (domain.ComparisonJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.jobs {
		if v.BaseVersionID == base && v.TargetVersionID == target {
			return v, true
		}
	}
	return domain.ComparisonJob{}, false
}
func (s *MemoryStore) UpdateJob(job domain.ComparisonJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[job.ID]; !ok {
		return domain.ErrNotFound
	}
	s.jobs[job.ID] = job
	return s.persist()
}
func (s *MemoryStore) SaveResult(result domain.ComparisonResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[result.JobID] = result
	return s.persist()
}
func (s *MemoryStore) GetResult(id string) (domain.ComparisonResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.results[id]
	if !ok {
		return domain.ComparisonResult{}, domain.ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) CreateComment(c domain.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comments[c.ID] = c
	s.commentsByDocument[c.DocumentID] = append(s.commentsByDocument[c.DocumentID], c.ID)
	return s.persist()
}
func (s *MemoryStore) GetComment(id string) (domain.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.comments[id]
	if !ok {
		return domain.Comment{}, domain.ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) UpdateComment(c domain.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.comments[c.ID]; !ok {
		return domain.ErrNotFound
	}
	s.comments[c.ID] = c
	return s.persist()
}
func (s *MemoryStore) ListComments(id string) []domain.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Comment{}
	for _, id := range s.commentsByDocument[id] {
		out = append(out, s.comments[id])
	}
	return out
}
func (s *MemoryStore) CreateReply(r domain.CommentReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replies[r.CommentID] = append(s.replies[r.CommentID], r)
	return s.persist()
}
func (s *MemoryStore) ListReplies(id string) []domain.CommentReply {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.CommentReply(nil), s.replies[id]...)
}
func (s *MemoryStore) CreateRollback(r domain.RollbackRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollbacks = append(s.rollbacks, r)
	return s.persist()
}
func (s *MemoryStore) AddAudit(a domain.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audits[a.EntityID] = append(s.audits[a.EntityID], a)
	return s.persist()
}
func (s *MemoryStore) ListAudits(id string) []domain.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.AuditLog(nil), s.audits[id]...)
}
func (s *MemoryStore) ResumeJobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0)
	for id, job := range s.jobs {
		if job.Status == domain.JobPending || job.Status == domain.JobRunning {
			job.Status = domain.JobPending
			job.StartedAt = nil
			s.jobs[id] = job
			ids = append(ids, id)
		}
	}
	_ = s.persist()
	return ids
}
