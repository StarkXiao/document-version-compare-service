package application
import (
	"crypto/rand"
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"encoding/hex"
	"time"
)
type Services struct {
	Documents   *DocumentService
	Comparisons *ComparisonService
	Comments    *CommentService
	Rollbacks   *RollbackService
	Store       repository.Store
}
func NewServices(store repository.Store, enqueue func(string) bool) *Services {
	comparison := NewComparisonService(store, enqueue)
	return &Services{Store: store, Documents: NewDocumentService(store, comparison.Enqueue), Comparisons: comparison, Comments: NewCommentService(store), Rollbacks: NewRollbackService(store, comparison.Enqueue)}
}
func id(prefix string) string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return prefix + time.Now().Format("20060102150405")
	}
	return prefix + hex.EncodeToString(bytes)
}
func audit(store repository.Store, actor, action, entity, entityID, trace string, detail map[string]string) {
	_ = store.AddAudit(domain.AuditLog{ID: id("audit_"), ActorID: actor, Action: action, EntityType: entity, EntityID: entityID, TraceID: trace, Detail: detail, CreatedAt: time.Now().UTC()})
}
