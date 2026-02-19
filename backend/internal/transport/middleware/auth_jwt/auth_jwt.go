package auth_jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport/helpers"
)

type ctxKey string

const ctxUserID ctxKey = "userID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserID)
	s, ok := v.(string)
	return s, ok
}

func Middleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				helpers.WriteJSON(w, 401, map[string]string{"error": "missing token"})
				return
			}

			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				helpers.WriteJSON(w, 401, map[string]string{"error": "invalid auth header"})
				return
			}

			raw := parts[1]

			token, err := jwt.ParseWithClaims(raw, &auth.Claims{}, func(t *jwt.Token) (any, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				helpers.WriteJSON(w, 401, map[string]string{"error": "invalid token"})
				return
			}

			claims, ok := token.Claims.(*auth.Claims)
			if !ok {
				helpers.WriteJSON(w, 401, map[string]string{"error": "invalid token"})
				return
			}

			userID := claims.Subject
			if userID == "" {
				helpers.WriteJSON(w, 401, map[string]string{"error": "invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
