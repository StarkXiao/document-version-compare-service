package transport
import (
	"document-version-compare-service/internal/domain"
	"encoding/json"
	"net/http"
)
type envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id"`
}
func respond(w http.ResponseWriter, status int, code, message, trace string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Code: code, Message: message, Data: data, TraceID: trace})
}
func success(w http.ResponseWriter, r *http.Request, data any) {
	respond(w, http.StatusOK, "ok", "success", trace(r), data)
}
func created(w http.ResponseWriter, r *http.Request, data any) {
	respond(w, http.StatusCreated, "ok", "created", trace(r), data)
}
func failure(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		err = domain.ErrInvalid
	}
	status, code := http.StatusInternalServerError, "internal_error"
	switch err {
	case domain.ErrNotFound:
		status, code = http.StatusNotFound, "not_found"
	case domain.ErrInvalid:
		status, code = http.StatusBadRequest, "invalid_request"
	case domain.ErrConflict:
		status, code = http.StatusConflict, "conflict"
	case domain.ErrForbidden:
		status, code = http.StatusForbidden, "forbidden"
	}
	respond(w, status, code, err.Error(), trace(r), nil)
}
