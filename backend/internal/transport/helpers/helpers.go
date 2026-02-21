package helpers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil && logger != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
	}
}

func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func HandleError(w http.ResponseWriter, logger *slog.Logger, status int, err error) {
	WriteJSON(w, logger, status, map[string]string{"error": err.Error()})
}
