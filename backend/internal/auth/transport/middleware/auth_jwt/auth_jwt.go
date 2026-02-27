package auth_jwt

//import (
//	"barter-port/internal/auth/model"
//	"barter-port/internal/auth/repository/user"
//	"barter-port/internal/auth/service/jwt"
//	"barter-port/internal/auth/transport/helpers"
//	"barter-port/internal/authkit"
//	"context"
//	"errors"
//	"log/slog"
//	"net/http"
//
//	"github.com/go-chi/chi/v5/middleware"
//	"github.com/google/uuid"
//)
//
//const ContextKeyUser contextKey = "user"
//
//type UserGetter interface {
//	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
//}
//
//// GetClaims retrieves the JWT claims from the context. It returns the claims and a boolean indicating whether the claims were found.
//func GetClaims(ctx context.Context) (*jwt.Claims, bool) {
//	c, ok := ctx.Value(ContextKeyUser).(*jwt.Claims)
//	return c, ok
//}
//
//func Middleware(logger *slog.Logger, jwtManager *jwt.Manager, users UserGetter) func(http.Handler) http.Handler {
//	return func(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			requestID := middleware.GetReqID(r.Context())
//			logger = logger.With(slog.String("request_id", requestID))
//			logger.Info("handling request with JWT authentication")
//
//			raw, err := extractBearerToken(r)
//			if err != nil {
//				if errors.Is(err, authkit.ErrMissingToken) {
//					logger.Warn("missing token in request")
//					helpers.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrMissingToken)
//					return
//				}
//				logger.Error("unexpected error extracting token", slog.String("error", err.Error()))
//				helpers.HandleError(w, logger, http.StatusInternalServerError, authkit.ErrInternalServerError)
//				return
//			}
//
//			claims, err := jwtManager.ParseAccessToken(raw)
//			if err != nil {
//				if errors.Is(err, jwt.ErrTokenExpired) {
//					logger.Info("token expired")
//					helpers.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrTokenExpired)
//					return
//				}
//				logger.Warn("invalid token", slog.String("error", err.Error()))
//				helpers.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrInvalidToken)
//				return
//			}
//
//			u, err := users.GetByID(r.Context(), claims.UserID)
//			if err != nil {
//				if errors.Is(err, user.ErrUserNotFound) {
//					logger.Warn(
//						`user not found for token UserID`,
//						slog.String("user_id", claims.UserID.String()),
//					)
//					helpers.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrInvalidToken)
//					return
//				}
//				logger.Error(
//					"unexpected error fetching user",
//					slog.String("error", err.Error()),
//					slog.String("user_id", claims.UserID.String()),
//				)
//				helpers.HandleError(w, logger, http.StatusInternalServerError, authkit.ErrInternalServerError)
//				return
//			}
//
//			logger.Info(
//				"user authenticated successfully",
//				slog.String("user_id", u.ID.String()),
//			)
//			ctx := context.WithValue(r.Context(), ContextKeyUser, claims)
//			next.ServeHTTP(w, r.WithContext(ctx))
//		})
//	}
//}
//
//// extractBearerToken extracts and cleans from bearerPrefix the token string from the Authorization header.
////
//// It returns error errMissingToken if the header is missing.
