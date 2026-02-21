package logger

import (
	"log/slog"
	"os"
)

func NewJSONLogger(level slog.Level, service string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler).With(
		slog.String("service", service),
	)
}
