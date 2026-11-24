package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jenish-ad/go-payments-microservices/internal/platform/apperr"
)

// errorBody is the shape of every error response from every service
type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// classify maps an error to an HTTP status and a stable error code.
// Clients should use the code to decide how to handle the error.
func classify(err error) (int, string) {
	var reqErr *requestError
	var valErr *apperr.ValidationError
	var ruleErr *apperr.RuleError
	switch {
	case errors.As(err, &reqErr):
		return reqErr.status, reqErr.code
	case errors.As(err, &valErr):
		return http.StatusUnprocessableEntity, "invalid_input"
	case errors.As(err, &ruleErr):
		return http.StatusUnprocessableEntity, ruleErr.Code
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, apperr.ErrUnauthenticated):
		return http.StatusUnauthorized, "unauthenticated"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

// Error sends err to the clients as a JSON error response. Client errors show
// the error's message. Server errors are logged and get a generic response.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, code := classify(err)
	d := errorDetail{Code: code, Message: err.Error()}
	if valErr, ok := errors.AsType[*apperr.ValidationError](err); ok {
		d.Fields = valErr.Fields
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "err", err)
		d.Message = "internal server error"
	}
	writeError(w, r, status, d)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, d errorDetail) {
	d.RequestID = RequestIDFrom(r.Context())
	// The response contains only strings and a map of strings, which always encode.
	body, _ := json.Marshal(errorBody{Error: d})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// NotFound answers requests that match no route with a JSON 404, instead of
// net/http's plain-text one.
func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, r, fmt.Errorf("route %w", apperr.ErrNotFound))
}
