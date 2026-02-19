package transport

import (
	"errors"
	"net/http"

	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
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
	if err := decodeJSON(r, &req); err != nil {
		handleError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	res, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			handleError(w, http.StatusBadRequest, auth.ErrInvalidEmail)
		case errors.Is(err, auth.ErrPasswordTooShort):
			handleError(w, http.StatusBadRequest, auth.ErrPasswordTooShort)
		case errors.Is(err, auth.ErrEmailAlreadyInUse):
			handleError(w, http.StatusBadRequest, auth.ErrEmailAlreadyInUse)
		default:
			handleError(w, http.StatusInternalServerError, ErrInternalServerError)
		}
	}

	writeJSON(w, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyReq
	if err := decodeJSON(r, &req); err != nil {
		handleError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			handleError(w, http.StatusBadRequest, auth.ErrInvalidToken)
		case errors.Is(err, auth.ErrTokenExpired):
			handleError(w, http.StatusBadRequest, auth.ErrTokenExpired)
		case errors.Is(err, auth.ErrUserNotFound):
			handleError(w, http.StatusNotFound, auth.ErrUserNotFound)
		default:
			handleError(w, http.StatusInternalServerError, ErrInternalServerError)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
