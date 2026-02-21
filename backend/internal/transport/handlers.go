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
	logger := h.logger.With("request_id", requestID)
	logger.Info("handling register request")

	var req registerReq
	if err := helpers.DecodeJSON(r, &req); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	res, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrInvalidEmail)
		case errors.Is(err, auth.ErrPasswordTooShort):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrPasswordTooShort)
		case errors.Is(err, auth.ErrEmailAlreadyInUse):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrEmailAlreadyInUse)
		default:
			helpers.HandleError(w, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With("request_id", requestID)
	logger.Info("handling verify email request")

	var req verifyEmailReq
	if err := helpers.DecodeJSON(r, &req); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrInvalidToken)
		case errors.Is(err, auth.ErrTokenExpired):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrTokenExpired)
		case errors.Is(err, auth.ErrUserNotFound):
			helpers.HandleError(w, http.StatusNotFound, auth.ErrUserNotFound)
		default:
			helpers.HandleError(w, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With("request_id", requestID)
	logger.Info("handling login request")

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.HandleError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	res, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			helpers.HandleError(w, http.StatusBadRequest, auth.ErrInvalidCredentials)
		case errors.Is(err, auth.ErrEmailNotVerified):
			helpers.HandleError(w, http.StatusForbidden, auth.ErrEmailNotVerified)
		default:
			helpers.HandleError(w, http.StatusInternalServerError, ErrInternalServerError)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, loginResp{AccessToken: res.AccessToken})
}

type meResp struct {
	UserID string `json:"userId"`
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	logger := h.logger.With("request_id", requestID)
	logger.Info("handling me request")

	userID, ok := auth_jwt.UserIDFromContext(r.Context())
	if !ok {
		helpers.HandleError(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, meResp{UserID: userID})
}
