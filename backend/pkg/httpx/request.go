package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func DecodeJSONWithLogs[T any](w http.ResponseWriter, r *http.Request, log *slog.Logger, dst *T) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		log.Warn("error decoding request",
			slog.String("error", err.Error()),
			slog.Any("request_body", r.Body),
			slog.Any("destination_struct", dst),
		)
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}
