package domain
import "time"
type Document struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Summary          string    `json:"summary"`
	CreatedBy        string    `json:"created_by"`
	CurrentVersionID string    `json:"current_version_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type Version struct {
	ID             string    `json:"id"`
	DocumentID     string    `json:"document_id"`
	Number         int       `json:"number"`
	Content        string    `json:"content"`
	ContentHash    string    `json:"content_hash"`
	Message        string    `json:"message"`
	CreatedBy      string    `json:"created_by"`
	ParentID       string    `json:"parent_id,omitempty"`
	RollbackFromID string    `json:"rollback_from_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
type Paragraph struct {
	ID        string `json:"id"`
	VersionID string `json:"version_id"`
	Position  int    `json:"position"`
	Content   string `json:"content"`
	Hash      string `json:"hash"`
}
type JobStatus string
const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
)
type ComparisonJob struct {
	ID              string     `json:"id"`
	DocumentID      string     `json:"document_id"`
	BaseVersionID   string     `json:"base_version_id"`
	TargetVersionID string     `json:"target_version_id"`
	Status          JobStatus  `json:"status"`
	Attempts        int        `json:"attempts"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}
type ChangeType string
const (
	Unchanged ChangeType = "unchanged"
	Added     ChangeType = "added"
	Deleted   ChangeType = "deleted"
	Modified  ChangeType = "modified"
	Moved     ChangeType = "moved"
)
type ComparisonItem struct {
	ID          string     `json:"id"`
	Type        ChangeType `json:"type"`
	OldPosition int        `json:"old_position,omitempty"`
	NewPosition int        `json:"new_position,omitempty"`
	OldContent  string     `json:"old_content,omitempty"`
	NewContent  string     `json:"new_content,omitempty"`
}
type ComparisonResult struct {
	JobID       string           `json:"job_id"`
	Items       []ComparisonItem `json:"items"`
	Summary     map[string]int   `json:"summary"`
	CompletedAt time.Time        `json:"completed_at"`
}
