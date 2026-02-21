package auth_jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ndandriianov/barter_port/backend/internal/model"
	"github.com/ndandriianov/barter_port/backend/internal/repository/user"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport/helpers"
)

type contextKey struct{}

var userCtxKey contextKey

var (
	errMissingToken            = errors.New("missing token")
	errInvalidAuthHeader       = errors.New("invalid auth header")
	errUnexpectedSigningMethod = errors.New("unexpected signing method")
	errInvalidToken            = errors.New("invalid token")
	errTokenExpired            = errors.New("token expired")
	errInternalServerError     = errors.New("internal server error")
)

type UserGetter interface {
	GetByID(id string) (model.User, error)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(userCtxKey).(string)
	return u, ok
}

// TODO: добавить логирование ошибок для мониторинга и обнаружения потенциальных атак, но не логировать чувствительную информацию из токенов

func Middleware(secret []byte, users UserGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			raw, err := extractBearerToken(r)
			if err != nil {
				if errors.Is(err, errMissingToken) {
					helpers.HandleError(w, http.StatusUnauthorized, errMissingToken)
					return
				}
				if errors.Is(err, errInvalidAuthHeader) {
					helpers.HandleError(w, http.StatusUnauthorized, errInvalidAuthHeader)
					return
				}
				helpers.HandleError(w, http.StatusInternalServerError, errInternalServerError)
				return
			}

			claims, err := parseToken(raw, secret)
			if err != nil {
				if errors.Is(err, errUnexpectedSigningMethod) {
					// TODO: log this as a potential attack attempt
					helpers.HandleError(w, http.StatusUnauthorized, errInvalidToken)
					return
				}
				if errors.Is(err, errTokenExpired) {
					// TODO: log this for monitoring purposes, but don't treat it as an attack attempt
					helpers.HandleError(w, http.StatusUnauthorized, errTokenExpired)
					return
				}
				// TODO: log the error for monitoring purposes
				helpers.HandleError(w, http.StatusUnauthorized, errInvalidToken)
				return
			}

			u, err := users.GetByID(claims.Subject)
			if err != nil {
				if errors.Is(err, user.ErrUserNotFound) {
					helpers.HandleError(w, http.StatusUnauthorized, errInvalidToken)
					return
				}
				helpers.HandleError(w, http.StatusInternalServerError, errInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, u.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errMissingToken
	}

	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errInvalidAuthHeader
	}

	return parts[1], nil
}

func parseToken(raw string, secret []byte) (*auth.Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &auth.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethod
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errTokenExpired
		}
		return nil, errInvalidToken
	}

	claims, ok := token.Claims.(*auth.Claims)
	if !ok || claims.Subject == "" {
		return nil, errInvalidToken
	}

	return claims, nil
}

// TODO: СДЕЛАТЬ КОРРЕКТНУЮ ЗАПИСЬ ИНФОРМАЦИИ О ПОЛЬЗОВАТЕЛЕ В КОНТЕКСТ, ЧТОБЫ В БУДУЩЕМ МОЖНО БЫЛО ИСПОЛЬЗОВАТЬ ЭТУ ИНФОРМАЦИЮ ДЛЯ РАЗЛИЧНЫХ ЦЕЛЕЙ (НАПРИМЕР, РАЗРЕШЕНИЯ). НАСТРОИТЬ ИНИЦИАЛИЗАЦИЮ MIDDLEWATR
