package http

import (
	"barter-port/contracts/openapi/users/types"
	"barter-port/internal/users/application/user"
	"barter-port/internal/users/domain"
	"barter-port/pkg/authkit"
	httpapi "barter-port/pkg/http_api"
	httplog "barter-port/pkg/logger"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

type Handlers struct {
	userService *user.Service
}

func NewHandlers(userService *user.Service) *Handlers {
	return &Handlers{userService: userService}
}

// ================================================================================
// GetUser
// ================================================================================

func (h *Handlers) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		log.Warn("invalid user id", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Info("user not found", slog.String("user_id", userID.String()))
			writeError(w, http.StatusNotFound, "user not found")
			return
		}

		log.Error("failed to get user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, types.User{
		Id:   u.Id,
		Name: u.Name,
		Bio:  u.Bio,
	})
}

// ================================================================================
// GetMe
// ================================================================================

func (h *Handlers) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	me, err := h.getMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Error("current user is absent in users storage", slog.String("user_id", userID.String()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		log.Error("failed to get current user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, me)
}

// ================================================================================
// UpdateMe
// ================================================================================

func (h *Handlers) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Warn("failed to get user id from context")
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req types.UpdateUserRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		log.Warn("failed to decode update user request", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Name == nil && req.Bio == nil {
		writeError(w, http.StatusBadRequest, "empty update payload")
		return
	}

	if req.Name != nil {
		if err := h.userService.UpdateName(r.Context(), userID, *req.Name); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	if req.Bio != nil {
		if err := h.userService.UpdateBio(r.Context(), userID, req.Bio); err != nil {
			handleUpdateError(w, log, err, userID)
			return
		}
	}

	me, err := h.getMe(r.Context(), userID)
	if err != nil {
		log.Error("failed to load updated user", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, me)
}

func handleUpdateError(w http.ResponseWriter, log *slog.Logger, err error, userID uuid.UUID) {
	updateErrLog := log.With(slog.Any("userID", userID), slog.Any("error", err))

	if errors.Is(err, domain.ErrUserNotFound) {
		updateErrLog.Info("user not found")
		writeError(w, http.StatusNotFound, "user not found")
	} else {
		updateErrLog.Error("failed to update user")
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ================================================================================
// helpers
// ================================================================================

func (h *Handlers) getMe(ctx context.Context, userID uuid.UUID) (types.Me, error) {
	me, err := h.userService.GetMe(ctx, userID)
	if err != nil {
		return types.Me{}, err
	}

	return types.Me{
		Id:        me.Id,
		Name:      me.Name,
		Bio:       me.Bio,
		Email:     openapitypes.Email(me.Email),
		CreatedAt: me.CreatedAt,
	}, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	httpapi.WriteJSON(w, status, types.ErrorResponse{
		Message: &message,
	})
}
