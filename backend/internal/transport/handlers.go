package transport

import (
	"encoding/json"
	"net/http"

	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	res, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, registerResp{
		UserID: res.UserID,
		Email:  res.Email,
	})
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
