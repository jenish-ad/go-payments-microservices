package httpx

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"

	"github.com/jenish-ad/go-payments-microservices/internal/platform/logging"
)

// RequestIDHeader is the HTTP header used to pass a request ID between services and return it to the client.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// RequestID preserves a valid incoming request ID or generates a new one.
// It adds the ID to the request context for logging and sets the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID(id) {
			id = rand.Text()
		}
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		ctx = logging.With(ctx, slog.String("request_id", id))
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom returns the request ID in ctx, or "" if absent.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// validRequestID accepts 1 to 64 ASCII letters, digits, hyphens, and underscores. The header comes from
// the client and we copy it into our logs and responses.
func validRequestID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, c := range id {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}
