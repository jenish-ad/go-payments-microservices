package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxBodyBytes caps request bodies. Our requests are tiny and the cap stops a
// client from making us read a huge body into memory.
const maxBodyBytes = 1 << 20 // 1MiB

// WriteJSON sends v as JSON with the given status. It encodes v before writing
// anything so an encoding failure can return a 500.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		Error(w, r, fmt.Errorf("encode response: %w", err))
		return
	}
	writeBody(w, status, body)
}
func writeBody(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(body, '\n'))
}

// DecodeJSON reads the request body into dst. It is strict on purpose: one
// JSON value of at most 1 MiB, with no unknown fields. Error maps decoding failures to HTTP 400 or 413.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := dec.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			return decodeError(err)
		}
		return badRequest("body must contain a single JSON value")
	}
	return nil
}

// requestError describes an invalid HTTP request, such as a malformed JSON body.
type requestError struct {
	status int
	code   string
	msg    string
}

func (e *requestError) Error() string { return e.msg }

func badRequest(msg string) error {
	return &requestError{http.StatusBadRequest, "bad_request", msg}
}

// NewError returns an error that Error sends with exactly this status and
// code. It's for problems with the request itself, like a missing header.
func NewError(status int, code, msg string) error {
	return &requestError{status, code, msg}
}

// decodeError turns encoding/json's errors into messages a client can act on.
func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var sizeErr *http.MaxBytesError
	switch {
	case errors.Is(err, io.EOF):
		return badRequest("body must not be empty")
	case errors.As(err, &syntaxErr):
		return badRequest(fmt.Sprintf("malformed JSON at byte %d", syntaxErr.Offset))
	case errors.Is(err, io.ErrUnexpectedEOF):
		return badRequest("malformed JSON")
	case errors.As(err, &typeErr):
		return badRequest(fmt.Sprintf("field %q must be of type %s", typeErr.Field, typeErr.Type))
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		return badRequest(strings.TrimPrefix(err.Error(), "json: "))
	case errors.As(err, &sizeErr):
		msg := fmt.Sprintf("body must be at most %d bytes", sizeErr.Limit)
		return &requestError{http.StatusRequestEntityTooLarge, "body_too_large", msg}
	default:
		return badRequest("invalid JSON body")
	}
}
