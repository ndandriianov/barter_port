package auth_jwt

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"barter-port/internal/infrastructure/repository/user"
	"barter-port/internal/model"
	"barter-port/internal/service/auth/jwt"
	"barter-port/internal/transport/helpers"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

const bearerPrefix = "Bearer "

type contextKey string

const ContextKeyUser contextKey = "user"

var (
	errMissingToken        = errors.New("missing token")
	errInvalidToken        = errors.New("invalid token")
	errTokenExpired        = errors.New("token expired")
	errInternalServerError = errors.New("internal server error")
)

type UserGetter interface {
	GetByID(id uuid.UUID) (model.User, error)
}

// GetClaims retrieves the JWT claims from the context. It returns the claims and a boolean indicating whether the claims were found.
func GetClaims(ctx context.Context) (*jwt.Claims, bool) {
	c, ok := ctx.Value(ContextKeyUser).(*jwt.Claims)
	return c, ok
}

func Middleware(logger *slog.Logger, jwtManager *jwt.Manager, users UserGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.GetReqID(r.Context())
			logger = logger.With(slog.String("request_id", requestID))
			logger.Info("handling request with JWT authentication")

			raw, err := extractBearerToken(r)
			if err != nil {
				if errors.Is(err, errMissingToken) {
					logger.Warn("missing token in request")
					helpers.HandleError(w, logger, http.StatusUnauthorized, errMissingToken)
					return
				}
				logger.Error("unexpected error extracting token", slog.String("error", err.Error()))
				helpers.HandleError(w, logger, http.StatusInternalServerError, errInternalServerError)
				return
			}

			claims, err := jwtManager.ParseAccessToken(raw)
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					logger.Info("token expired")
					helpers.HandleError(w, logger, http.StatusUnauthorized, errTokenExpired)
					return
				}
				logger.Warn("invalid token", slog.String("error", err.Error()))
				helpers.HandleError(w, logger, http.StatusUnauthorized, errInvalidToken)
				return
			}

			u, err := users.GetByID(claims.UserID)
			if err != nil {
				if errors.Is(err, user.ErrUserNotFound) {
					logger.Warn(
						`user not found for token UserID`,
						slog.String("user_id", claims.UserID.String()),
					)
					helpers.HandleError(w, logger, http.StatusUnauthorized, errInvalidToken)
					return
				}
				logger.Error(
					"unexpected error fetching user",
					slog.String("error", err.Error()),
					slog.String("user_id", claims.UserID.String()),
				)
				helpers.HandleError(w, logger, http.StatusInternalServerError, errInternalServerError)
				return
			}

			logger.Info(
				"user authenticated successfully",
				slog.String("user_id", u.ID.String()),
			)
			ctx := context.WithValue(r.Context(), ContextKeyUser, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts and cleans from bearerPrefix the token string from the Authorization header.
//
// It returns error errMissingToken if the header is missing.
func extractBearerToken(r *http.Request) (string, error) {
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		return "", errMissingToken
	}

	if len(tokenStr) > len(bearerPrefix) && tokenStr[:len(bearerPrefix)] == bearerPrefix {
		tokenStr = tokenStr[len(bearerPrefix):]
	}

	return tokenStr, nil
}
