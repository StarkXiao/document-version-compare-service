package application
import (
	"context"
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"time"
)
type ComparisonService struct {
	store   repository.Store
	enqueue func(string) bool
	ctx     context.Context
}
func (s *ComparisonService) SetContext(ctx context.Context) { s.ctx = ctx }
func NewComparisonService(store repository.Store, enqueue func(string) bool) *ComparisonService {
	return &ComparisonService{store: store, enqueue: enqueue}
}
func (s *ComparisonService) Enqueue(base, target, documentID string) (domain.ComparisonJob, error) {
	if existing, ok := s.store.FindJob(base, target); ok {
		return existing, nil
	}
	job := domain.ComparisonJob{ID: id("job_"), DocumentID: documentID, BaseVersionID: base, TargetVersionID: target, Status: domain.JobPending, CreatedAt: time.Now().UTC()}
	if err := s.store.CreateJob(job); err != nil {
		return domain.ComparisonJob{}, err
	}
	s.schedule(job.ID)
	return job, nil
}
func (s *ComparisonService) schedule(jobID string) {
	if s.enqueue(jobID) {
		return
	}
	go func() {
		for !s.enqueue(jobID) {
			if s.ctx == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()
}
func (s *ComparisonService) Create(base, target, actor, trace string) (domain.ComparisonJob, error) {
	old, err := s.store.GetVersion(base)
	if err != nil {
		return domain.ComparisonJob{}, err
	}
	next, err := s.store.GetVersion(target)
	if err != nil {
		return domain.ComparisonJob{}, err
	}
	if old.DocumentID != next.DocumentID || old.ID == next.ID {
		return domain.ComparisonJob{}, domain.ErrInvalid
	}
	job, err := s.Enqueue(base, target, old.DocumentID)
	if err == nil {
		audit(s.store, actor, "comparison.queued", "comparison", job.ID, trace, map[string]string{"base": base, "target": target})
	}
	return job, err
}
func (s *ComparisonService) Process(ctx context.Context, jobID string) {
	if jobID == "" {
		return
	}
	job, err := s.store.GetJob(jobID)
	if err != nil {
		return
	}
	if claimer, ok := s.store.(interface {
		ClaimJob(string) (domain.ComparisonJob, bool)
	}); ok {
		var claimed bool
		job, claimed = claimer.ClaimJob(jobID)
		if !claimed {
			return
		}
	} else {
		if job.Status != domain.JobPending {
			return
		}
		now := time.Now().UTC()
		job.Status = domain.JobRunning
		job.Attempts++
		job.StartedAt = &now
		_ = s.store.UpdateJob(job)
	}
	old, oldErr := s.store.GetVersion(job.BaseVersionID)
	next, nextErr := s.store.GetVersion(job.TargetVersionID)
	if oldErr != nil || nextErr != nil {
		s.fail(job, "version unavailable")
		return
	}
	items := domain.DiffParagraphs(s.store.Paragraphs(old.ID), s.store.Paragraphs(next.ID))
	summary := map[string]int{}
	for _, item := range items {
		summary[string(item.Type)]++
	}
	finished := time.Now().UTC()
	result := domain.ComparisonResult{JobID: job.ID, Items: items, Summary: summary, CompletedAt: finished}
	_ = s.store.SaveResult(result)
	job.Status = domain.JobSucceeded
	job.FinishedAt = &finished
	_ = s.store.UpdateJob(job)
	audit(s.store, "worker", "comparison.completed", "comparison", job.ID, "", map[string]string{"items": itoa(len(items))})
}
func (s *ComparisonService) fail(job domain.ComparisonJob, message string) {
	now := time.Now().UTC()
	job.Status = domain.JobFailed
	job.ErrorMessage = message
	job.FinishedAt = &now
	_ = s.store.UpdateJob(job)
}
func (s *ComparisonService) Get(id string) (domain.ComparisonJob, domain.ComparisonResult, error) {
	job, err := s.store.GetJob(id)
	if err != nil {
		return domain.ComparisonJob{}, domain.ComparisonResult{}, err
	}
	result, _ := s.store.GetResult(id)
	return job, result, nil
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	return string(buf[i:])
}
