package responder

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func RespondError(w http.ResponseWriter, err error) {
	if e, ok := err.(*Error); ok {
		Respond(w, e.Code, e)
	} else {
		Respond(w, http.StatusInternalServerError, &Error{Message: err.Error()})
	}
}
