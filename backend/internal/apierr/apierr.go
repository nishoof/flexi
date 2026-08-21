package apierr

import (
	"encoding/json"
	"log"
	"net/http"
)

// Error is an HTTP-aware error. Message is safe for clients; Err is internal.
type Error struct {
	Status  int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func BadRequest(msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Message: msg}
}

func NotFound(msg string) *Error {
	return &Error{Status: http.StatusNotFound, Message: msg}
}

func Unauthorized(msg string) *Error {
	return &Error{Status: http.StatusUnauthorized, Message: msg}
}

func Internal(err error) *Error {
	return &Error{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Err:     err,
	}
}

func (e *Error) WriteTo(w http.ResponseWriter) {
	if e == nil {
		return
	}
	if e.Status >= 500 {
		log.Printf("error (http %d): %v", e.Status, e.Err)
	}

	w.WriteHeader(e.Status)

	response := struct {
		Error string `json:"error"`
	}{
		Error: e.Message,
	}
	_ = json.NewEncoder(w).Encode(response)
}
