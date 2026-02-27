package authkit

import (
	"errors"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

type ErrorResponder func(w http.ResponseWriter, r *http.Request, status int, err error)

func defaultErrorResponder(w http.ResponseWriter, _ *http.Request, status int, _ error) {
	w.WriteHeader(status)
}

func Middleware(v Validator, onError ErrorResponder) func(http.Handler) http.Handler {
	if onError == nil {
		onError = defaultErrorResponder
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				if errors.Is(err, ErrMissingToken) {
					onError(w, r, http.StatusUnauthorized, ErrMissingToken)
					return
				}
				onError(w, r, http.StatusInternalServerError, err)
				return
			}

			p, err := v.ValidateAccess(r.Context(), token)
			if err != nil {
				switch {
				case errors.Is(err, ErrTokenExpired):
					onError(w, r, http.StatusUnauthorized, ErrTokenExpired)
				case errors.Is(err, ErrInvalidToken):
					onError(w, r, http.StatusUnauthorized, ErrInvalidToken)
				case errors.Is(err, ErrUnavailable):
					onError(w, r, http.StatusServiceUnavailable, ErrUnavailable)
				default:
					onError(w, r, http.StatusInternalServerError, err)
					return
				}
			}

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
