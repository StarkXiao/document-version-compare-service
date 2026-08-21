package application
import (
	"document-version-compare-service/internal/domain"
	"document-version-compare-service/internal/repository"
	"strings"
	"time"
)
type CommentService struct{ store repository.Store }
func NewCommentService(store repository.Store) *CommentService { return &CommentService{store: store} }
type CreateCommentInput struct {
	DocumentID  string `json:"document_id"`
	VersionID   string `json:"version_id"`
	ParagraphID string `json:"paragraph_id"`
	Content     string `json:"content"`
	ActorID     string `json:"-"`
	TraceID     string `json:"-"`
}
func (s *CommentService) Create(in CreateCommentInput) (domain.Comment, error) {
	in.Content = strings.TrimSpace(in.Content)
	version, err := s.store.GetVersion(in.VersionID)
	if err != nil {
		return domain.Comment{}, err
	}
	if version.DocumentID != in.DocumentID || in.Content == "" {
		return domain.Comment{}, domain.ErrInvalid
	}
	if in.ParagraphID != "" {
		found := false
		for _, p := range s.store.Paragraphs(in.VersionID) {
			if p.ID == in.ParagraphID {
				found = true
			}
		}
		if !found {
			return domain.Comment{}, domain.ErrNotFound
		}
	}
	now := time.Now().UTC()
	comment := domain.Comment{ID: id("comment_"), DocumentID: in.DocumentID, VersionID: in.VersionID, ParagraphID: in.ParagraphID, Content: in.Content, AuthorID: in.ActorID, Status: domain.CommentOpen, CreatedAt: now, UpdatedAt: now}
	if err = s.store.CreateComment(comment); err != nil {
		return domain.Comment{}, err
	}
	audit(s.store, in.ActorID, "comment.created", "document", in.DocumentID, in.TraceID, map[string]string{"comment": comment.ID})
	return comment, nil
}
func (s *CommentService) Reply(commentID, content, actor, trace string) (domain.CommentReply, error) {
	if _, err := s.store.GetComment(commentID); err != nil {
		return domain.CommentReply{}, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return domain.CommentReply{}, domain.ErrInvalid
	}
	reply := domain.CommentReply{ID: id("reply_"), CommentID: commentID, Content: content, AuthorID: actor, CreatedAt: time.Now().UTC()}
	if err := s.store.CreateReply(reply); err != nil {
		return domain.CommentReply{}, err
	}
	audit(s.store, actor, "comment.replied", "comment", commentID, trace, nil)
	return reply, nil
}
func (s *CommentService) SetStatus(commentID string, status domain.CommentStatus, actor, trace string) (domain.Comment, error) {
	comment, err := s.store.GetComment(commentID)
	if err != nil {
		return domain.Comment{}, err
	}
	if status != domain.CommentResolved && status != domain.CommentReopened && status != domain.CommentOpen {
		return domain.Comment{}, domain.ErrInvalid
	}
	comment.Status = status
	comment.UpdatedAt = time.Now().UTC()
	if err = s.store.UpdateComment(comment); err != nil {
		return domain.Comment{}, err
	}
	audit(s.store, actor, "comment.status_changed", "comment", commentID, trace, map[string]string{"status": string(status)})
	return comment, nil
}
func (s *CommentService) List(documentID string) []domain.Comment {
	return s.store.ListComments(documentID)
}
func (s *CommentService) Replies(id string) []domain.CommentReply { return s.store.ListReplies(id) }
