package transport
import (
	"document-version-compare-service/internal/application"
	"document-version-compare-service/internal/domain"
	"encoding/json"
	"net/http"
)
func (s *Server) createComparison(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in struct {
		BaseVersionID   string `json:"base_version_id"`
		TargetVersionID string `json:"target_version_id"`
	}
	if err := decode(w, r, &in); err != nil {
		failure(w, r, err)
		return
	}
	job, err := s.services.Comparisons.Create(in.BaseVersionID, in.TargetVersionID, actor(r), trace(r))
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, job)
}
func (s *Server) getComparison(w http.ResponseWriter, r *http.Request) {
	job, result, err := s.services.Comparisons.Get(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, map[string]any{"job": job, "result": result})
}
func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in application.CreateCommentInput
	if err := decode(w, r, &in); err != nil {
		failure(w, r, err)
		return
	}
	in.ActorID, in.TraceID = actor(r), trace(r)
	comment, err := s.services.Comments.Create(in)
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, comment)
}
func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	success(w, r, s.services.Comments.List(r.PathValue("id")))
}
func (s *Server) replyComment(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		failure(w, r, err)
		return
	}
	reply, err := s.services.Comments.Reply(r.PathValue("id"), in.Content, actor(r), trace(r))
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, reply)
}
func (s *Server) updateComment(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in struct {
		Status domain.CommentStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		failure(w, r, err)
		return
	}
	comment, err := s.services.Comments.SetStatus(r.PathValue("id"), in.Status, actor(r), trace(r))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, comment)
}
