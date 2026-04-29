package http

import (
	"barter-port/internal/auth/application"
	"barter-port/internal/auth/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/db"
	"barter-port/pkg/httpx"
	"barter-port/pkg/jwt"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const RefreshCookieName = "refresh_token"

var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrMissingRefreshCookie = errors.New("missing refresh token cookie")
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, exec db.DB, token domain.RefreshToken) error
	GetByJTI(ctx context.Context, exec db.DB, jti string) (domain.RefreshToken, error)
	Revoke(ctx context.Context, exec db.DB, jti string) error
	DeleteAllForUser(ctx context.Context, exec db.DB, userID uuid.UUID) error
}

type Handlers struct {
	logger      *slog.Logger
	authService *application.Service
	jwtManager  *jwt.Manager
	db          *pgxpool.Pool
	refreshRepo RefreshTokenRepository
}

func NewHandlers(
	logger *slog.Logger,
	authService *application.Service,
	jwtManager *jwt.Manager,
	db *pgxpool.Pool,
	refreshRepo RefreshTokenRepository,
) *Handlers {
	return &Handlers{
		logger:      logger,
		authService: authService,
		jwtManager:  jwtManager,
		db:          db,
		refreshRepo: refreshRepo,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentialsReq body credentialsReq true "Register request"
// @Success 200 {object} registerResp
// @Failure 400 {object} httpx.ErrorResponse "Invalid request"
// @Failure 500 "Internal server error"
// @Router /auth/register [post]
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling register request")

	var req credentialsReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	res, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			logger.Info("invalid email format", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidEmail)

		case errors.Is(err, domain.ErrPasswordTooShort):
			logger.Info("password too short", slog.Int("password_length", len(req.Password)))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrPasswordTooShort)

		case errors.Is(err, domain.ErrEmailAlreadyInUse):
			logger.Info("email already in use", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrEmailAlreadyInUse)

		default:
			logger.Error("unexpected error", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("successfully registered user",
		slog.String("user_id", res.UserID.String()),
		slog.String("email", res.Email),
	)

	httpx.WriteJSON(w, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

// RetrySendVerificationEmail godoc
// @Summary Retry sending verification email
// @Description Generates a new email verification token and sends a verification email if the user's email is not verified.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentialsReq body credentialsReq true "Retry send verification email request"
// @Success 200 "Verification email sent successfully"
// @Failure 400 {object} httpx.ErrorResponse "Invalid request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 "Internal server error"
// @Router /auth/retry-send-verification-email [post]
func (h *Handlers) RetrySendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling RetrySendVerificationEmail request")

	var req credentialsReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if err := h.authService.RetrySendVerificationEmail(r.Context(), req.Email, req.Password); err != nil {
		log := logger.With(slog.String("email", req.Email), slog.Any("error", err))

		switch {
		case errors.Is(err, domain.ErrInvalidCredentials),
			errors.Is(err, domain.ErrEmailNotVerified),
			errors.Is(err, domain.ErrIncorrectPassword):
			log.Info("authentication failed")
			httpx.WriteEmptyError(w, http.StatusUnauthorized)

		default:
			log.Error("unexpected error in RetrySendVerificationEmail request")
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("successfully sent verification email", slog.String("email", req.Email))
	w.WriteHeader(http.StatusOK)
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verifies a user's email using a token
// @Tags auth
// @Accept json
// @Param verifyEmailReq body verifyEmailReq true "Verify email request"
// @Produce plain
// @Success 200 "status: ok"
// @Failure 400 {object} httpx.ErrorResponse "Invalid request or token"
// @Failure 404 {object} httpx.ErrorResponse "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/verify-email [post]
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling verify email request")

	var req verifyEmailReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmailToken):
			logger.Info("invalid email_token in verify email request")
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidEmailToken)

		case errors.Is(err, domain.ErrEmailTokenExpired):
			logger.Info("email_token expired in verify email request")
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrEmailTokenExpired)

		case errors.Is(err, domain.ErrUserNotFound):
			logger.Info("user not found in verify email request")
			httpx.WriteError(w, http.StatusNotFound, domain.ErrUserNotFound)

		default:
			logger.Error(
				"unexpected error in verify email request",
				slog.String("error", err.Error()),
			)
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("successfully verified email user")

	w.WriteHeader(http.StatusOK)
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Sends a password reset link to the user's email if the account exists
// @Tags auth
// @Accept json
// @Produce json
// @Param requestPasswordResetReq body requestPasswordResetReq true "Password reset request"
// @Success 200 "Password reset email sent successfully"
// @Failure 400 {object} httpx.ErrorResponse "Invalid request or email"
// @Failure 500 "Internal server error"
// @Router /auth/request-password-reset [post]
func (h *Handlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling request password reset request")

	var req requestPasswordResetReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if err := h.authService.RequestPasswordReset(r.Context(), req.Email); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			logger.Info("invalid email format in request password reset", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidEmail)

		default:
			logger.Error("unexpected error in request password reset",
				slog.String("email", req.Email),
				slog.Any("error", err),
			)
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ResetPassword godoc
// @Summary Reset password
// @Description Sets a new password using a password reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param resetPasswordReq body resetPasswordReq true "Reset password request"
// @Success 200 "Password reset successfully"
// @Failure 400 {object} httpx.ErrorResponse "Invalid request, token or password"
// @Failure 404 {object} httpx.ErrorResponse "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/reset-password [post]
func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling reset password request")

	var req resetPasswordReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordTooShort):
			logger.Info("new password too short in reset password request")
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrPasswordTooShort)

		case errors.Is(err, domain.ErrInvalidPasswordResetToken):
			logger.Info("invalid password reset token")
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidPasswordResetToken)

		case errors.Is(err, domain.ErrPasswordResetTokenExpired):
			logger.Info("password reset token expired")
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrPasswordResetTokenExpired)

		case errors.Is(err, domain.ErrUserNotFound):
			logger.Info("user not found in reset password")
			httpx.WriteError(w, http.StatusNotFound, domain.ErrUserNotFound)

		default:
			logger.Error("unexpected error in reset password request", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Login godoc
// @Summary Login user
// @Description Logs in a user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentialsReq body credentialsReq true "Login request"
// @Success 200 {object} loginResp
// @Failure 400 {object} httpx.ErrorResponse "Invalid request or credentialsReq"
// @Failure 401 {object} httpx.ErrorResponse "Incorrect password"
// @Failure 403 {object} httpx.ErrorResponse "Email not verified"
// @Failure 500 "Internal server error"
// @Router /auth/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling login request")

	// Парсинг email и password из тела запроса
	var req credentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("error decoding", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	// Проверка учетных данных и получение userID
	userID, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			logger.Info("invalid credentials", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidCredentials)

		case errors.Is(err, domain.ErrIncorrectPassword):
			logger.Info("incorrect password", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusUnauthorized, domain.ErrIncorrectPassword)

		case errors.Is(err, domain.ErrEmailNotVerified):
			logger.Info("email not verified", slog.String("email", req.Email))
			httpx.WriteError(w, http.StatusForbidden, domain.ErrEmailNotVerified)

		default:
			logger.Error("unexpected error", slog.Any("error", err), slog.String("email", req.Email))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("successfully logged in user", slog.String("email", req.Email))
	logger.Info("generating JWT for logged in user", slog.String("email", req.Email))

	// Генерация access и refresh токенов
	access, err := h.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		logger.Error("failed to generate access token for logged in user",
			slog.Any("error", err),
			slog.String("email", req.Email),
			slog.String("user_id", userID.String()),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	refresh, claims, err := h.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		logger.Error("failed to generate refresh token for logged in user",
			slog.Any("error", err),
			slog.String("email", req.Email),
			slog.String("user_id", userID.String()),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	// Сохранение refresh токена в репозитории
	err = h.refreshRepo.Save(r.Context(), h.db, domain.RefreshToken{
		JTI:       claims.ID,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
		Revoked:   false,
	})
	if err != nil {
		logger.Error("failed to save refresh token for logged in user",
			slog.Any("error", err),
			slog.String("email", req.Email),
			slog.String("refresh_token", refresh),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
	}

	// Установка refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	httpx.WriteJSON(w, http.StatusOK, loginResp{AccessToken: access})

	logger.Info("successfully generated tokens for logged in user",
		slog.String("email", req.Email),
		slog.String("user_id", userID.String()),
	)
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Refreshes access and refresh tokens using the refresh token
// @Tags auth
// @Accept json
// @Param Cookie header string true "refresh_token=<JWT refresh token>"
// @Produce json
// @Success 200 {object} refreshResponse
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized or invalid token"
// @Failure 500 "Internal server error"
// @Router /auth/refresh [post]
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling refresh request")

	// Парсинг refresh токена из тела запроса
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		logger.Warn("missing refresh token cookie", slog.Any("error", err))
		httpx.WriteError(w, http.StatusUnauthorized, ErrMissingRefreshCookie)
		return
	}

	oldRefreshClaims, err := h.jwtManager.ParseRefreshToken(cookie.Value)
	if err != nil {
		if errors.Is(err, authkit.ErrTokenExpired) {
			logger.Info("refresh token expired", slog.Any("error", err))
			httpx.WriteError(w, http.StatusUnauthorized, authkit.ErrTokenExpired)
			return
		}
		logger.Warn("invalid refresh token", slog.Any("error", err))
		httpx.WriteError(w, http.StatusUnauthorized, authkit.ErrInvalidToken)
		return
	}

	// Получение и проверка хранимого refresh токена по JTI из claims
	storedRefresh, err := h.refreshRepo.GetByJTI(r.Context(), h.db, oldRefreshClaims.ID)
	if err != nil {
		storedRefreshFailedLogger := logger.With(
			slog.String("jti", oldRefreshClaims.ID),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)

		if errors.Is(err, domain.ErrRefreshNotFound) {
			storedRefreshFailedLogger.Info("refresh token not found in repository")
			httpx.WriteError(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		storedRefreshFailedLogger.Error("error fetching refresh token from repository", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	if storedRefresh.Revoked || storedRefresh.ExpiresAt.Before(time.Now()) {
		httpx.WriteError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Создание новых токенов
	access, err := h.jwtManager.GenerateAccessToken(oldRefreshClaims.UserID)
	if err != nil {
		logger.Error(
			"failed to generate access token for logged in user",
			slog.Any("error", err),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	refresh, claims, err := h.jwtManager.GenerateRefreshToken(oldRefreshClaims.UserID)
	if err != nil {
		logger.Error(
			"failed to generate refresh token for logged in user",
			slog.Any("error", err),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	// Сохранение нового refresh токена и удаление старого
	err = h.refreshRepo.Save(r.Context(), h.db, domain.RefreshToken{
		JTI:       claims.ID,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
		Revoked:   false,
	})
	if err != nil {
		logger.Error(
			"failed to save new refresh token for user",
			slog.Any("error", err),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
			slog.String("new_jti", claims.ID),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	err = h.refreshRepo.Revoke(r.Context(), h.db, oldRefreshClaims.ID)
	if err != nil {
		logger.Error(
			"failed to revoke old refresh token for user",
			slog.Any("error", err),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
			slog.String("old_jti", oldRefreshClaims.ID),
		)
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	// Установка нового refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	httpx.WriteJSON(w, http.StatusOK, refreshResponse{AccessToken: access})

	logger.Info("successfully refreshed tokens for user", slog.String("user_id", oldRefreshClaims.UserID.String()))
}

// Logout godoc
// @Summary Logout user
// @Description Logs out a user by revoking the refresh token
// @Tags auth
// @Produce json
// @Success 200 "Logout successful"
// @Failure 500 "Internal server error"
// @Router /auth/logout [post]
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling logout request")

	cookie, err := r.Cookie(RefreshCookieName)
	// Отзыв refresh токена, если он есть. Если куки нет или он невалидный - просто очищаем куки и возвращаем успех
	if err == nil {
		if claims, err := h.jwtManager.ParseRefreshToken(cookie.Value); err == nil {
			if err := h.refreshRepo.Revoke(r.Context(), h.db, claims.ID); err != nil {
				logger.Error("failed to revoke refresh token during logout",
					slog.Any("error", err),
					slog.String("jti", claims.ID),
				)
				httpx.WriteEmptyError(w, http.StatusInternalServerError)
				return
			}
		}
	}

	clearRefreshCookie(w)
	w.WriteHeader(http.StatusOK)
}

type meResp struct {
	UserID uuid.UUID `json:"userId"`
}

// ChangePassword godoc
// @Security BearerAuth
// @Summary Change password
// @Description Changes the authenticated user's password after verifying old credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param changePasswordReq body changePasswordReq true "Change password request"
// @Success 200 "Password changed successfully"
// @Failure 400 {object} httpx.ErrorResponse "New password validation failed"
// @Failure 403 {object} httpx.ErrorResponse "Old credentials are invalid or missing"
// @Failure 404 {object} httpx.ErrorResponse "User not found"
// @Failure 500 "Internal server error"
// @Router /auth/change-password [post]
func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling change password request")

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		logger.Error("failed to fetch principal for password change")
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	var req changePasswordReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		logger.Warn("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req.OldEmail, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordTooShort):
			logger.Info("new password too short", slog.String("user_id", userID.String()))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrPasswordTooShort)

		case errors.Is(err, domain.ErrInvalidOldCredentials):
			logger.Info("invalid old credentials in password change", slog.String("user_id", userID.String()))
			httpx.WriteError(w, http.StatusForbidden, domain.ErrInvalidOldCredentials)

		case errors.Is(err, domain.ErrUserNotFound):
			logger.Info("user not found in password change", slog.String("user_id", userID.String()))
			httpx.WriteError(w, http.StatusNotFound, domain.ErrUserNotFound)

		default:
			logger.Error("unexpected error in change password request",
				slog.String("user_id", userID.String()),
				slog.Any("error", err),
			)
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Me godoc
// @Security BearerAuth
// @Summary Get user info
// @Description Retrieves information about the authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} meResp
// @Failure 500 "Internal server error"
// @Router /auth/me [get]
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling me request")

	userId, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		logger.Error("failed to fetch principal")
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	logger.Info("successfully fetched user info")
	httpx.WriteJSON(w, http.StatusOK, meResp{UserID: userId})
}

// GetAdminPlatformStatistics godoc
// @Summary Get auth platform statistics for admin
// @Description Returns auth-owned platform statistics for the administrative panel
// @Tags auth, admin, statistics
// @Produce json
// @Success 200 {object} adminAuthPlatformStatisticsResp
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 500 "Internal server error"
// @Security BearerAuth
// @Router /auth/admin/statistics/platform [get]
func (h *Handlers) GetAdminPlatformStatistics(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	stats, err := h.authService.GetAdminPlatformStatistics(r.Context(), requesterID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrUserNotFound):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			h.logger.Error("failed to get admin auth platform statistics", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, adminAuthPlatformStatisticsResp{
		Users: adminAuthPlatformUsersStatisticsResp{
			TotalRegistered: stats.Users.TotalRegistered,
			VerifiedEmails:  stats.Users.VerifiedEmails,
		},
	})
}

// GetAdminUserStatistics godoc
// @Summary Get auth user statistics for admin
// @Description Returns auth-owned user statistics for the administrative panel
// @Tags auth, admin, statistics
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} adminAuthUserStatisticsResp
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "User not found"
// @Failure 500 "Internal server error"
// @Security BearerAuth
// @Router /auth/admin/users/{id}/statistics [get]
func (h *Handlers) GetAdminUserStatistics(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	stats, err := h.authService.GetAdminUserStatistics(r.Context(), requesterID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrUserNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			h.logger.Error("failed to get admin auth user statistics", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, adminAuthUserStatisticsResp{
		UserID:        stats.UserID,
		RegisteredAt:  stats.RegisteredAt.UTC().Format(time.RFC3339),
		EmailVerified: stats.EmailVerified,
	})
}

type userCreationStatusResp struct {
	Status string `json:"status"`
}

// GetUserCreationStatus godoc
// @Summary Get user creation status
// @Description Retrieves user creation status by user ID
// @Tags auth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} userCreationStatusResp
// @Failure 400 {object} httpx.ErrorResponse "Invalid user ID"
// @Failure 404 {object} httpx.ErrorResponse "User creation event not found"
// @Failure 500 "Internal server error"
// @Router /auth/status/{userId} [get]
func (h *Handlers) GetUserCreationStatus(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling get user creation status request")

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		logger.Info("invalid user id in user creation status request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	status, err := h.authService.GetUserCreationStatus(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			logger.Info("user creation event not found", slog.String("user_id", userID.String()))
			httpx.WriteError(w, http.StatusNotFound, domain.ErrUserNotFound)
		default:
			logger.Error("failed to fetch user creation status",
				slog.String("user_id", userID.String()),
				slog.Any("error", err),
			)
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	logger.Info("successfully fetched user creation status", slog.String("user_id", userID.String()))
	httpx.WriteJSON(w, http.StatusOK, userCreationStatusResp{Status: status})
}

//
// === AUTH - SPECIFIC HELPERS ===
//

func setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   false, // TODO: В PROD ОБЯЗАТЕЛЬНО true (HTTPS)
		SameSite: http.SameSiteStrictMode,
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false, // TODO: В PROD ОБЯЗАТЕЛЬНО true (HTTPS)
		SameSite: http.SameSiteStrictMode,
	})
}
