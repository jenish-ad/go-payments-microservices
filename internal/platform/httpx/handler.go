package httpx

import (
	"net/http"
	"strconv"

	"github.com/jenish-ad/go-payments-microservices/internal/platform/apperr"
)

// Handle Func is an HTTP handler that return its error instead of writing it.
// The error goes to the client through Error like every other error. Only return one before writing anything to w.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		Error(w, r, err)
	}
}

//A QueryInt returns the wuery parameter key as an int or def of its absent.
//A value that isnt a whole number is a calidaitonn error on that parameter.

func QueryInt(r *http.Request, key string, def int) (int, error) {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		var v apperr.ValidationError
		v.Add(key, "must be a whole number")
		return 0, v.Err()
	}
	return n, nil
}
