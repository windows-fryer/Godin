package responder

import (
	"encoding/json"
	"errors"
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

func Respond(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)

		if err != nil {
			return err
		}
	}

	return nil
}

func RespondError(w http.ResponseWriter, err error) {
	var e *Error

	if errors.As(err, &e) {
		if err := Respond(w, e.Code, e); err != nil {
			panic(err)
		}
	} else {
		if err := Respond(w, http.StatusInternalServerError, NewError(http.StatusInternalServerError, "Internal Server Error")); err != nil {
			panic(err)
		}
	}
}
