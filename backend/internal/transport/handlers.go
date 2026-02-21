package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport/helpers"
	"github.com/ndandriianov/barter_port/backend/internal/transport/middleware/auth_jwt"
)

var (
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInternalServerError = errors.New("internal server error")
	ErrUnauthorized        = errors.New("unauthorized")
)

type Handlers struct {
	logger      *slog.Logger
	authService *auth.Service
}

func NewHandlers(logger *slog.Logger, authService *auth.Service) *Handlers {
	return &Handlers{
		logger:      logger,
		authService: authService,
	}
}

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

	res, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			logger.Info(
				"invalid email format in register request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrInvalidEmail)

		case errors.Is(err, auth.ErrPasswordTooShort):
			logger.Info(
				"password too short in register request",
				slog.Int("password_length", len(req.Password)),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrPasswordTooShort)

		case errors.Is(err, auth.ErrEmailAlreadyInUse):
			logger.Info(
				"email already in use in register request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrEmailAlreadyInUse)

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
		slog.String("user_id", res.UserID),
		slog.String("email", res.Email),
	)

	helpers.WriteJSON(w, logger, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

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

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmailToken):
			logger.Info("invalid email_token in verify email request")
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrInvalidEmailToken)

		case errors.Is(err, auth.ErrInvalidEmailToken):
			logger.Info("email_token expired in verify email request")
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrInvalidEmailToken)

		case errors.Is(err, auth.ErrUserNotFound):
			logger.Info("user not found in verify email request")
			helpers.HandleError(w, logger, http.StatusNotFound, auth.ErrUserNotFound)

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

	helpers.WriteJSON(w, logger, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling login request")

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

	res, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			logger.Info(
				"invalid credentials in login request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusBadRequest, auth.ErrInvalidCredentials)

		case errors.Is(err, auth.ErrEmailNotVerified):
			logger.Info(
				"email not verified in login request",
				slog.String("email", req.Email),
			)
			helpers.HandleError(w, logger, http.StatusForbidden, auth.ErrEmailNotVerified)

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

	helpers.WriteJSON(w, logger, http.StatusOK, loginResp{AccessToken: res.AccessToken})
}

type meResp struct {
	UserID string `json:"userId"`
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With(slog.String("request_id", requestID))
	logger.Info("handling me request")

	userID, ok := auth_jwt.UserIDFromContext(r.Context())
	if !ok {
		logger.Warn("user ID not found in context")
		helpers.HandleError(w, logger, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	logger.Info("successfully fetched user info")

	helpers.WriteJSON(w, logger, http.StatusOK, meResp{UserID: userID})
}
