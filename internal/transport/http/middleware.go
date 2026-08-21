package transport
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)
type ctxKey string
const traceKey ctxKey = "trace"
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tid := r.Header.Get("X-Trace-ID")
		if tid == "" {
			bytes := make([]byte, 6)
			_, _ = rand.Read(bytes)
			tid = hex.EncodeToString(bytes)
		}
		w.Header().Set("X-Trace-ID", tid)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceKey, tid)))
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
func trace(r *http.Request) string { value, _ := r.Context().Value(traceKey).(string); return value }
func actor(r *http.Request) string {
	if user := r.Header.Get("X-User-ID"); user != "" {
		return user
	}
	return "anonymous"
}
func role(r *http.Request) string { return r.Header.Get("X-Role") }
func requireRole(r *http.Request, allowed ...string) bool {
	if actor(r) == "anonymous" {
		return false
	}
	for _, value := range allowed {
		if role(r) == value {
			return true
		}
	}
	return false
}
