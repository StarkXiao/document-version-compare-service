package repository
import "document-version-compare-service/internal/domain"
type Store interface {
	CreateDocument(domain.Document, domain.Version, []domain.Paragraph) error
	GetDocument(string) (domain.Document, error)
	ListDocuments() []domain.Document
	CreateVersion(domain.Version, []domain.Paragraph) error
	GetVersion(string) (domain.Version, error)
	ListVersions(string) []domain.Version
	Paragraphs(string) []domain.Paragraph
	UpdateDocument(domain.Document) error
	CreateJob(domain.ComparisonJob) error
	GetJob(string) (domain.ComparisonJob, error)
	FindJob(string, string) (domain.ComparisonJob, bool)
	UpdateJob(domain.ComparisonJob) error
	SaveResult(domain.ComparisonResult) error
	GetResult(string) (domain.ComparisonResult, error)
	CreateComment(domain.Comment) error
	GetComment(string) (domain.Comment, error)
	UpdateComment(domain.Comment) error
	ListComments(string) []domain.Comment
	CreateReply(domain.CommentReply) error
	ListReplies(string) []domain.CommentReply
	CreateRollback(domain.RollbackRecord) error
	AddAudit(domain.AuditLog) error
	ListAudits(string) []domain.AuditLog
}
