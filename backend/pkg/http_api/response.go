package http_api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Error string `json:"error"` // The error message
}

func newErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

func WriteJSONWithLogs(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil && logger != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
	}
}

func HandleError(w http.ResponseWriter, logger *slog.Logger, status int, err error) {
	WriteJSONWithLogs(w, logger, status, newErrorResponse(err))
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Message *string `json:"message,omitempty"`
	}{Message: &msg})
}

const (
	CannotDecodeRequestBody = "cannot decode request body"
)
