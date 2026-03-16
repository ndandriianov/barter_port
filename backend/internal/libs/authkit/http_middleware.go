package authkit

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

type ErrorResponder func(w http.ResponseWriter, r *http.Request, status int, err error)

func defaultErrorResponder(w http.ResponseWriter, _ *http.Request, status int, _ error) {
	w.WriteHeader(status)
}

func Middleware(logger *slog.Logger, v Validator, onError ErrorResponder) func(http.Handler) http.Handler {
	if onError == nil {
		onError = defaultErrorResponder
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLogger := logger.With(slog.String("url", r.URL.Path), slog.String("method", r.Method))

			token, err := extractBearerToken(r)
			if err != nil {
				reqLogger.Warn("failed to extract bearer token", slog.String("error", err.Error()))
				onError(w, r, http.StatusUnauthorized, err)
				return
			}

			p, err := v.ValidateAccess(r.Context(), token)
			if err != nil {
				switch {
				case errors.Is(err, ErrTokenExpired):
					reqLogger.Info("token expired")
					onError(w, r, http.StatusUnauthorized, ErrTokenExpired)

				case errors.Is(err, ErrInvalidToken):
					reqLogger.Warn("invalid token", slog.String("error", err.Error()))
					onError(w, r, http.StatusUnauthorized, ErrInvalidToken)

				case errors.Is(err, ErrUnavailable):
					reqLogger.Error("authentication service unavailable", slog.String("error", err.Error()))
					onError(w, r, http.StatusServiceUnavailable, ErrUnavailable)

				default:
					reqLogger.Error("unexpected error validating token", slog.String("error", err.Error()))
					onError(w, r, http.StatusInternalServerError, err)
				}
				return
			}

			reqLogger.Debug("user authenticated successfully",
				slog.String("user_id", p.UserID.String()),
				slog.String("jti", p.JTI),
			)

			ctx := WithPrincipal(r.Context(), p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		return "", ErrMissingToken
	}

	tokenStr = strings.TrimSpace(tokenStr)
	if strings.HasPrefix(tokenStr, bearerPrefix) {
		tokenStr = strings.TrimPrefix(tokenStr, bearerPrefix)
	}

	if tokenStr == "" {
		return "", ErrMissingToken
	}

	return tokenStr, nil
}
