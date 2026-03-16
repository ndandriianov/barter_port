package logger

import (
	"barter-port/internal/libs/authkit"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/context"
)

type ctxKeyLogger struct{}

func Middleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())
			log := base.With(slog.String("request_id", reqID))
			userId, ok := authkit.UserIDFromContext(r.Context())
			if ok {
				log = log.With(slog.String("user_id", userId.String()))
			}
			ctx := context.WithValue(r.Context(), ctxKeyLogger{}, log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LogFrom(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if v := ctx.Value(ctxKeyLogger{}); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	fallback.Warn("no logger in context, using fallback logger")
	return fallback
}
