package transport
import (
	"document-version-compare-service/internal/application"
	"document-version-compare-service/internal/domain"
	"encoding/json"
	"net/http"
)
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	success(w, r, map[string]string{"status": "ready"})
}
func (s *Server) listDocuments(w http.ResponseWriter, r *http.Request) {
	success(w, r, s.services.Documents.List())
}
func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in application.CreateDocumentInput
	if err := decode(w, r, &in); err != nil {
		failure(w, r, err)
		return
	}
	in.ActorID, in.TraceID = actor(r), trace(r)
	doc, version, err := s.services.Documents.Create(in)
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, map[string]any{"document": doc, "version": version})
}
func (s *Server) getDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := s.services.Documents.Get(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, doc)
}
func (s *Server) listVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.services.Documents.Versions(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, versions)
}
func (s *Server) createVersion(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "editor", "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in application.CreateVersionInput
	if err := decode(w, r, &in); err != nil {
		failure(w, r, err)
		return
	}
	in.ActorID, in.TraceID = actor(r), trace(r)
	version, err := s.services.Documents.CreateVersion(r.PathValue("id"), in)
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, version)
}
func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	version, paragraphs, err := s.services.Documents.Version(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, map[string]any{"version": version, "paragraphs": paragraphs})
}
func (s *Server) rollback(w http.ResponseWriter, r *http.Request) {
	if !s.useKey(r) {
		failure(w, r, domain.ErrConflict)
		return
	}
	if !requireRole(r, "admin") {
		failure(w, r, domain.ErrForbidden)
		return
	}
	var in struct {
		SourceVersionID string `json:"source_version_id"`
		Reason          string `json:"reason"`
	}
	if err := decode(w, r, &in); err != nil {
		failure(w, r, err)
		return
	}
	version, err := s.services.Rollbacks.Rollback(r.PathValue("id"), in.SourceVersionID, in.Reason, actor(r), trace(r))
	if err != nil {
		failure(w, r, err)
		return
	}
	created(w, r, version)
}
func (s *Server) listAudits(w http.ResponseWriter, r *http.Request) {
	success(w, r, s.audit.List(r.PathValue("id")))
}
func (s *Server) activityReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.audit.Report(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, report)
}
func (s *Server) changeReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.audit.Changes(r.PathValue("id"))
	if err != nil {
		failure(w, r, err)
		return
	}
	success(w, r, report)
}
