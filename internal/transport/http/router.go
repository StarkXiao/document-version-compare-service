package transport
import (
	"document-version-compare-service/internal/application"
	"net/http"
	"sync"
)
type Server struct {
	services *application.Services
	audit    *application.AuditService
	mux      *http.ServeMux
	keys     map[string]struct{}
	keyMu    sync.Mutex
}
func NewServer(services *application.Services) *Server {
	s := &Server{services: services, audit: application.NewAuditService(services.Store), mux: http.NewServeMux(), keys: map[string]struct{}{}}
	s.routes()
	return s
}
func (s *Server) useKey(r *http.Request) bool {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		return true
	}
	key = actor(r) + ":" + r.Method + ":" + r.URL.Path + ":" + key
	s.keyMu.Lock()
	defer s.keyMu.Unlock()
	if _, ok := s.keys[key]; ok {
		return false
	}
	s.keys[key] = struct{}{}
	return true
}
func (s *Server) Handler() http.Handler { return withMiddleware(s.mux) }
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /readyz", s.health)
	s.mux.HandleFunc("GET /api/v1/documents", s.listDocuments)
	s.mux.HandleFunc("POST /api/v1/documents", s.createDocument)
	s.mux.HandleFunc("GET /api/v1/documents/{id}", s.getDocument)
	s.mux.HandleFunc("GET /api/v1/documents/{id}/versions", s.listVersions)
	s.mux.HandleFunc("POST /api/v1/documents/{id}/versions", s.createVersion)
	s.mux.HandleFunc("POST /api/v1/documents/{id}/rollbacks", s.rollback)
	s.mux.HandleFunc("GET /api/v1/documents/{id}/audit-logs", s.listAudits)
	s.mux.HandleFunc("GET /api/v1/documents/{id}/activity-report", s.activityReport)
	s.mux.HandleFunc("GET /api/v1/documents/{id}/change-report", s.changeReport)
	s.mux.HandleFunc("GET /api/v1/versions/{id}", s.getVersion)
	s.mux.HandleFunc("POST /api/v1/comparisons", s.createComparison)
	s.mux.HandleFunc("GET /api/v1/comparisons/{id}", s.getComparison)
	s.mux.HandleFunc("POST /api/v1/comments", s.createComment)
	s.mux.HandleFunc("GET /api/v1/documents/{id}/comments", s.listComments)
	s.mux.HandleFunc("POST /api/v1/comments/{id}/replies", s.replyComment)
	s.mux.HandleFunc("PATCH /api/v1/comments/{id}/status", s.updateComment)
}
