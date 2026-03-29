package http

import (
	"barter-port/internal/auth/application"
	"barter-port/internal/auth/domain"
	"barter-port/internal/auth/infrastructure/repository/refresh_token"
	"barter-port/pkg/authkit"
	"barter-port/pkg/db"
	"barter-port/pkg/http_api"
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
	ErrInternalServerError  = errors.New("internal server error")
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
// @Failure 400 {object} http_api.ErrorResponse "Invalid request"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling register request")

	var req credentialsReq
	if err := http_api.DecodeJSON(r, &req); err != nil {
		logger.Warn(
			"error decoding register request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		http_api.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	res, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidEmail):
			logger.Info(
				"invalid email format in register request",
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrInvalidEmail)

		case errors.Is(err, application.ErrPasswordTooShort):
			logger.Info(
				"password too short in register request",
				slog.Int("password_length", len(req.Password)),
			)
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrPasswordTooShort)

		case errors.Is(err, application.ErrEmailAlreadyInUse):
			logger.Info(
				"email already in use in register request",
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrEmailAlreadyInUse)

		default:
			logger.Error(
				"unexpected error in register request",
				slog.String("error", err.Error()),
			)
			http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	logger.Info(
		"successfully registered user",
		slog.String("user_id", res.UserID.String()),
		slog.String("email", res.Email),
	)

	http_api.WriteJSONWithLogs(w, logger, http.StatusOK, registerResp{
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
// @Failure 400 {object} http_api.ErrorResponse "Invalid request"
// @Failure 401 {object} http_api.ErrorResponse "Unauthorized"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/retry-send-verification-email [post]
func (h *Handlers) RetrySendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling RetrySendVerificationEmail request")

	var req credentialsReq
	if err := http_api.DecodeJSON(r, &req); err != nil {
		logger.Warn(
			"error decoding RetrySendVerificationEmail request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		http_api.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if err := h.authService.RetrySendVerificationEmail(r.Context(), req.Email, req.Password); err != nil {
		log := logger.With(slog.String("email", req.Email), slog.Any("error", err))

		switch {
		case errors.Is(err, application.ErrInvalidCredentials),
			errors.Is(err, application.ErrEmailNotVerified),
			errors.Is(err, application.ErrIncorrectPassword):
			log.Info("authentication failed")
			http_api.HandleError(w, log, http.StatusUnauthorized, ErrUnauthorized)

		default:
			log.Error("unexpected error in RetrySendVerificationEmail request")
			http_api.HandleError(w, log, http.StatusInternalServerError, ErrInternalServerError)
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
// @Failure 400 {object} http_api.ErrorResponse "Invalid request or token"
// @Failure 404 {object} http_api.ErrorResponse "User not found"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/verify-email [post]
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling verify email request")

	var req verifyEmailReq
	if err := http_api.DecodeJSON(r, &req); err != nil {
		logger.Warn(
			"error decoding verify email request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		http_api.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidEmailToken):
			logger.Info("invalid email_token in verify email request")
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrInvalidEmailToken)

		case errors.Is(err, application.ErrInvalidEmailToken):
			logger.Info("email_token expired in verify email request")
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrInvalidEmailToken)

		case errors.Is(err, application.ErrUserNotFound):
			logger.Info("user not found in verify email request")
			http_api.HandleError(w, logger, http.StatusNotFound, application.ErrUserNotFound)

		default:
			logger.Error(
				"unexpected error in verify email request",
				slog.String("error", err.Error()),
			)
			http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	logger.Info("successfully verified email user")

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
// @Failure 400 {object} http_api.ErrorResponse "Invalid request or credentialsReq"
// @Failure 401 {object} http_api.ErrorResponse "Incorrect password"
// @Failure 403 {object} http_api.ErrorResponse "Email not verified"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling login request")

	// Парсинг email и password из тела запроса
	var req credentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(
			"error decoding login request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		http_api.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	// Проверка учетных данных и получение userID
	userID, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidCredentials):
			logger.Info(
				"invalid credentials in login request",
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusBadRequest, application.ErrInvalidCredentials)

		case errors.Is(err, application.ErrIncorrectPassword):
			logger.Info(
				"incorrect password in login request",
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusUnauthorized, application.ErrIncorrectPassword)

		case errors.Is(err, application.ErrEmailNotVerified):
			logger.Info(
				"email not verified in login request",
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusForbidden, application.ErrEmailNotVerified)

		default:
			logger.Error(
				"unexpected error in login request",
				slog.String("error", err.Error()),
				slog.String("email", req.Email),
			)
			http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	logger.Info("successfully logged in user", slog.String("email", req.Email))
	logger.Info("generating JWT for logged in user", slog.String("email", req.Email))

	// Генерация access и refresh токенов
	access, err := h.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		logger.Error(
			"failed to generate access token for logged in user",
			slog.String("error", err.Error()),
			slog.String("email", req.Email),
			slog.String("user_id", userID.String()),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	refresh, claims, err := h.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		logger.Error(
			"failed to generate refresh token for logged in user",
			slog.String("error", err.Error()),
			slog.String("email", req.Email),
			slog.String("user_id", userID.String()),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
		logger.Error(
			"failed to save refresh token for logged in user",
			slog.String("error", err.Error()),
			slog.String("email", req.Email),
			slog.String("refresh_token", refresh),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
	}

	// Установка refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	http_api.WriteJSONWithLogs(w, logger, http.StatusOK, loginResp{AccessToken: access})

	logger.Info(
		"successfully generated tokens for logged in user",
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
// @Failure 401 {object} http_api.ErrorResponse "Unauthorized or invalid token"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling refresh request")

	// Парсинг refresh токена из тела запроса
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		logger.Warn("missing refresh token cookie", slog.String("error", err.Error()))
		http_api.HandleError(w, logger, http.StatusUnauthorized, ErrMissingRefreshCookie)
		return
	}

	oldRefreshClaims, err := h.jwtManager.ParseRefreshToken(cookie.Value)
	if err != nil {
		if errors.Is(err, authkit.ErrTokenExpired) {
			logger.Info("refresh token expired", slog.String("error", err.Error()))
			http_api.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrTokenExpired)
			return
		}
		logger.Warn("invalid refresh token", slog.String("error", err.Error()))
		http_api.HandleError(w, logger, http.StatusUnauthorized, authkit.ErrInvalidToken)
		return
	}

	// Получение и проверка хранимого refresh токена по JTI из claims
	storedRefresh, err := h.refreshRepo.GetByJTI(r.Context(), h.db, oldRefreshClaims.ID)
	if err != nil {
		storedRefreshFailedLogger := logger.With(
			slog.String("jti", oldRefreshClaims.ID),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)

		if errors.Is(err, refresh_token.ErrRefreshNotFound) {
			storedRefreshFailedLogger.Info("refresh token not found in repository")
			http_api.HandleError(w, logger, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		storedRefreshFailedLogger.Error("error fetching refresh token from repository",
			slog.String("error", err.Error()),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	if storedRefresh.Revoked || storedRefresh.ExpiresAt.Before(time.Now()) {
		http_api.HandleError(w, logger, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	// Создание новых токенов
	access, err := h.jwtManager.GenerateAccessToken(oldRefreshClaims.UserID)
	if err != nil {
		logger.Error(
			"failed to generate access token for logged in user",
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	refresh, claims, err := h.jwtManager.GenerateRefreshToken(oldRefreshClaims.UserID)
	if err != nil {
		logger.Error(
			"failed to generate refresh token for logged in user",
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
			slog.String("new_jti", claims.ID),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	err = h.refreshRepo.Revoke(r.Context(), h.db, oldRefreshClaims.ID)
	if err != nil {
		logger.Error(
			"failed to revoke old refresh token for user",
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
			slog.String("old_jti", oldRefreshClaims.ID),
		)
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// Установка нового refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	http_api.WriteJSONWithLogs(w, logger, http.StatusOK, refreshResponse{AccessToken: access})

	logger.Info("successfully refreshed tokens for user", slog.String("user_id", oldRefreshClaims.UserID.String()))
}

// Logout godoc
// @Summary Logout user
// @Description Logs out a user by revoking the refresh token
// @Tags auth
// @Produce json
// @Success 200 "Logout successful"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
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
				logger.Error(
					"failed to revoke refresh token during logout",
					slog.String("error", err.Error()),
					slog.String("jti", claims.ID),
				)
				http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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

// Me godoc
// @Security BearerAuth
// @Summary Get user info
// @Description Retrieves information about the authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} meResp
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/me [get]
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling me request")

	userId, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		logger.Error("failed to fetch principal")
		http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	logger.Info("successfully fetched user info")
	http_api.WriteJSONWithLogs(w, logger, http.StatusOK, meResp{UserID: userId})
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
// @Failure 400 {object} http_api.ErrorResponse "Invalid user ID"
// @Failure 404 {object} http_api.ErrorResponse "User creation event not found"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /auth/status/{userId} [get]
func (h *Handlers) GetUserCreationStatus(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling get user creation status request")

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		logger.Info("invalid user id in user creation status request", slog.String("error", err.Error()))
		http_api.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	status, err := h.authService.GetUserCreationStatus(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrUserNotFound):
			logger.Info("user creation event not found", slog.String("user_id", userID.String()))
			http_api.HandleError(w, logger, http.StatusNotFound, application.ErrUserNotFound)
		default:
			logger.Error(
				"failed to fetch user creation status",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			http_api.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	logger.Info("successfully fetched user creation status", slog.String("user_id", userID.String()))
	http_api.WriteJSONWithLogs(w, logger, http.StatusOK, userCreationStatusResp{Status: status})
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
