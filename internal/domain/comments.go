package domain
import "time"
type CommentStatus string
const (
	CommentOpen     CommentStatus = "open"
	CommentResolved CommentStatus = "resolved"
	CommentReopened CommentStatus = "reopened"
)
type Comment struct {
	ID               string        `json:"id"`
	DocumentID       string        `json:"document_id"`
	VersionID        string        `json:"version_id"`
	ParagraphID      string        `json:"paragraph_id,omitempty"`
	ComparisonItemID string        `json:"comparison_item_id,omitempty"`
	Content          string        `json:"content"`
	AuthorID         string        `json:"author_id"`
	Status           CommentStatus `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}
type CommentReply struct {
	ID        string    `json:"id"`
	CommentID string    `json:"comment_id"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}
type RollbackRecord struct {
	ID               string    `json:"id"`
	DocumentID       string    `json:"document_id"`
	SourceVersionID  string    `json:"source_version_id"`
	CreatedVersionID string    `json:"created_version_id"`
	Reason           string    `json:"reason"`
	ActorID          string    `json:"actor_id"`
	CreatedAt        time.Time `json:"created_at"`
}
type AuditLog struct {
	ID         string            `json:"id"`
	ActorID    string            `json:"actor_id"`
	Action     string            `json:"action"`
	EntityType string            `json:"entity_type"`
	EntityID   string            `json:"entity_id"`
	TraceID    string            `json:"trace_id"`
	Detail     map[string]string `json:"detail"`
	CreatedAt  time.Time         `json:"created_at"`
}
