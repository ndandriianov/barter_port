package tags

import (
	"barter-port/contracts/openapi/deals/types"
	offersapp "barter-port/internal/deals/application/offers"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"
)

type Handlers struct {
	offersService *offersapp.Service
}

func NewHandlers(offersService *offersapp.Service) *Handlers {
	return &Handlers{offersService: offersService}
}

func (h *Handlers) HandleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.offersService.ListTags(r.Context())
	if err != nil {
		logger.LogFrom(r.Context(), slog.Default()).Error("failed to list tags", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	resp := make(types.ListTagsResponse, 0, len(tags))
	for _, tag := range tags {
		resp = append(resp, tag)
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) HandleDeleteAdminTag(w http.ResponseWriter, r *http.Request) {
	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	params := types.DeleteAdminTagParams{
		Name: r.URL.Query().Get("name"),
	}

	err := h.offersService.DeleteTag(r.Context(), userID, string(params.Name))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidTagName):
			httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrAdminOnly), errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrTagNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			logger.LogFrom(r.Context(), slog.Default()).Error("failed to delete tag", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
