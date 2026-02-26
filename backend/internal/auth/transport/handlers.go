package transport

import (
	"barter-port/internal/auth/model"
	"barter-port/internal/auth/service"
	"barter-port/internal/auth/service/jwt"
	"barter-port/internal/auth/transport/helpers"
	"barter-port/internal/auth/transport/middleware/auth_jwt"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

const RefreshCookieName = "refresh_token"

var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrInternalServerError  = errors.New("internal server error")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrMissingRefreshCookie = errors.New("missing refresh token cookie")
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, token model.RefreshToken) error
	GetByJTI(ctx context.Context, jti string) (model.RefreshToken, error)
	Revoke(ctx context.Context, jti string) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type Handlers struct {
	logger      *slog.Logger
	authService *service.Service
	jwtManager  *jwt.Manager
	refreshRepo RefreshTokenRepository
}

func NewHandlers(
	logger *slog.Logger,
	authService *service.Service,
	jwtManager *jwt.Manager,
	refreshRepo RefreshTokenRepository,
) *Handlers {
	return &Handlers{
		logger:      logger,
		authService: authService,
		jwtManager:  jwtManager,
		refreshRepo: refreshRepo,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param registerReq body registerReq true "Register request"
// @Success 200 {object} registerResp
// @Failure 400 {object} helpers.ErrorResponse "Invalid request"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling register request")

	var req registerReq
	if err := helpers.DecodeJSON(r, &req); err != nil {
		logger.Warn(
			"error decoding register request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		helpers.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	res, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			logger.Info(
				"invalid email format in register request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrInvalidEmail)

		case errors.Is(err, service.ErrPasswordTooShort):
			logger.Info(
				"password too short in register request",
				slog.Int("password_length", len(req.Password)),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrPasswordTooShort)

		case errors.Is(err, service.ErrEmailAlreadyInUse):
			logger.Info(
				"email already in use in register request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrEmailAlreadyInUse)

		default:
			logger.Error(
				"unexpected error in register request",
				slog.String("error", err.Error()),
			)
			helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	logger.Info(
		"successfully registered user",
		slog.String("user_id", res.UserID.String()),
		slog.String("email", res.Email),
	)

	helpers.WriteJSON(w, logger, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verifies a user's email using a token
// @Tags auth
// @Accept json
// @Produce plain
// @Success 200 "status: ok"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request or token"
// @Failure 404 {object} helpers.ErrorResponse "User not found"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/verify-email [post]
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling verify email request")

	var req verifyEmailReq
	if err := helpers.DecodeJSON(r, &req); err != nil {
		logger.Warn(
			"error decoding verify email request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		helpers.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmailToken):
			logger.Info("invalid email_token in verify email request")
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrInvalidEmailToken)

		case errors.Is(err, service.ErrInvalidEmailToken):
			logger.Info("email_token expired in verify email request")
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrInvalidEmailToken)

		case errors.Is(err, service.ErrUserNotFound):
			logger.Info("user not found in verify email request")
			helpers.HandleError(w, logger, http.StatusNotFound, service.ErrUserNotFound)

		default:
			logger.Error(
				"unexpected error in verify email request",
				slog.String("error", err.Error()),
			)
			helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
// @Param loginReq body loginReq true "Login request"
// @Success 200 {object} loginResp
// @Failure 400 {object} helpers.ErrorResponse "Invalid request or credentials"
// @Failure 401 {object} helpers.ErrorResponse "Incorrect password"
// @Failure 403 {object} helpers.ErrorResponse "Email not verified"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling login request")

	// Парсинг email и password из тела запроса
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(
			"error decoding login request",
			slog.Any("request_body", r.Body),
			slog.String("error", err.Error()),
		)
		helpers.HandleError(w, logger, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	// Проверка учетных данных и получение userID
	userID, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			logger.Info(
				"invalid credentials in login request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, service.ErrInvalidCredentials)

		case errors.Is(err, service.ErrIncorrectPassword):
			logger.Info(
				"incorrect password in login request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusUnauthorized, service.ErrIncorrectPassword)

		case errors.Is(err, service.ErrEmailNotVerified):
			logger.Info(
				"email not verified in login request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusForbidden, service.ErrEmailNotVerified)

		default:
			logger.Error(
				"unexpected error in login request",
				slog.String("error", err.Error()),
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// Сохранение refresh токена в репозитории
	err = h.refreshRepo.Save(r.Context(), model.RefreshToken{
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
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
	}

	// Установка refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	helpers.WriteJSON(w, logger, http.StatusOK, loginResp{AccessToken: access})

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
// @Produce json
// @Success 200 {object} refreshResponse
// @Failure 401 {object} helpers.ErrorResponse "Unauthorized or invalid token"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling refresh request")

	// Парсинг refresh токена из тела запроса
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		logger.Warn("missing refresh token cookie", slog.String("error", err.Error()))
		helpers.HandleError(w, logger, http.StatusUnauthorized, ErrMissingRefreshCookie)
		return
	}

	oldRefreshClaims, err := h.jwtManager.ParseRefreshToken(cookie.Value)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			logger.Info("refresh token expired", slog.String("error", err.Error()))
			helpers.HandleError(w, logger, http.StatusUnauthorized, jwt.ErrTokenExpired)
			return
		}
		logger.Warn("invalid refresh token", slog.String("error", err.Error()))
		helpers.HandleError(w, logger, http.StatusUnauthorized, jwt.ErrInvalidToken)
		return
	}

	// Получение и проверка хранимого refresh токена по JTI из claims
	storedRefresh, err := h.refreshRepo.GetByJTI(r.Context(), oldRefreshClaims.ID)
	if err != nil {
		// TODO: distinguish between "not found" and other errors in the repository
		logger.Error(
			"error fetching refresh token from repository",
			slog.String("error", err.Error()),
			slog.String("jti", oldRefreshClaims.ID),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	if storedRefresh.Revoked || storedRefresh.ExpiresAt.Before(time.Now()) {
		helpers.HandleError(w, logger, http.StatusUnauthorized, ErrUnauthorized)
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
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	refresh, claims, err := h.jwtManager.GenerateRefreshToken(oldRefreshClaims.UserID)
	if err != nil {
		logger.Error(
			"failed to generate refresh token for logged in user",
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
		)
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// Сохранение нового refresh токена и удаление старого
	err = h.refreshRepo.Save(r.Context(), model.RefreshToken{
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
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	err = h.refreshRepo.Revoke(r.Context(), oldRefreshClaims.ID)
	if err != nil {
		logger.Error(
			"failed to revoke old refresh token for user",
			slog.String("error", err.Error()),
			slog.String("user_id", oldRefreshClaims.UserID.String()),
			slog.String("old_jti", oldRefreshClaims.ID),
		)
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	// Установка нового refresh токена в cookie и отправка access токена в ответе
	setRefreshCookie(w, refresh, claims.ExpiresAt.Time)
	helpers.WriteJSON(w, logger, http.StatusOK, refreshResponse{AccessToken: access})

	logger.Info("successfully refreshed tokens for user", slog.String("user_id", oldRefreshClaims.UserID.String()))
}

// Logout godoc
// @Summary Logout user
// @Description Logs out a user by revoking the refresh token
// @Tags auth
// @Produce json
// @Success 200 "Logout successful"
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling logout request")

	cookie, err := r.Cookie(RefreshCookieName)
	// Отзыв refresh токена, если он есть. Если куки нет или он невалидный - просто очищаем куки и возвращаем успех
	if err == nil {
		if claims, err := h.jwtManager.ParseRefreshToken(cookie.Value); err == nil {
			if err := h.refreshRepo.Revoke(r.Context(), claims.ID); err != nil {
				logger.Error(
					"failed to revoke refresh token during logout",
					slog.String("error", err.Error()),
					slog.String("jti", claims.ID),
				)
				helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
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
// @Summary Get user info
// @Description Retrieves information about the authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} meResp
// @Failure 500 {object} helpers.ErrorResponse "Internal server error"
// @Router /auth/me [get]
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling me request")

	claims, ok := auth_jwt.GetClaims(r.Context())
	if !ok {
		logger.Error("failed to fetch claims")
		helpers.HandleError(w, logger, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	logger.Info("successfully fetched user info")
	helpers.WriteJSON(w, logger, http.StatusOK, meResp{UserID: claims.UserID})
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
