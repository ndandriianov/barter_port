package logger

import (
	"log/slog"
	"os"
)

func NewJSONLogger(level slog.Level, service string, layer string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	if service != "" && layer != "" {
		return slog.New(handler).With(
			slog.String("service", service),
			slog.String("layer", layer),
		)
	}

	if service != "" {
		return slog.New(handler).With(
			slog.String("service", service),
		)
	}

	if layer != "" {
		return slog.New(handler).With(
			slog.String("layer", layer),
		)
	}

	return slog.New(handler)
}
