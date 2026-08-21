package transport

import (
	"document-version-compare-service/internal/domain"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBug004WrappedNotFoundLosesClassification(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	failure(w, r, errors.New("load: "+domain.ErrNotFound.Error()))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d", w.Code)
	}
}
