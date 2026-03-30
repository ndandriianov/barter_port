package http_api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Error string `json:"error"` // The error message
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteErrorStr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Message *string `json:"message,omitempty"`
	}{Message: &msg})
}

func WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Message *string `json:"message,omitempty"`
	}{Message: new(err.Error())})
}

var ErrCannotDecodeRequestBody = errors.New("cannot decode request body")
