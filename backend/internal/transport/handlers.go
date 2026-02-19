package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport/helpers"
	"github.com/ndandriianov/barter_port/backend/internal/transport/middleware/auth_jwt"
)

var (
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInternalServerError = errors.New("internal server error")
)

type Handlers struct {
	authService *auth.Service
}

func NewHandlers(authService *auth.Service) *Handlers {
	return &Handlers{
		authService: authService,
	}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
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
	}

	helpers.WriteJSON(w, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
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
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}

	res, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			helpers.WriteJSON(w, 401, map[string]string{"error": "invalid credentials"})
		case errors.Is(err, auth.ErrEmailNotVerified):
			helpers.WriteJSON(w, 403, map[string]string{"error": "email not verified"})
		default:
			log.Printf("login error: %v", err)
			helpers.WriteJSON(w, 500, map[string]string{"error": "internal error"})
		}
		return
	}

	helpers.WriteJSON(w, 200, loginResp{AccessToken: res.AccessToken})
}

type meResp struct {
	UserID string `json:"userId"`
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth_jwt.UserIDFromContext(r.Context())
	if !ok {
		helpers.WriteJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	helpers.WriteJSON(w, 200, meResp{UserID: userID})
}
