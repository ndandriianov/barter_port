package http

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"log/slog"
	"net/http"
)

type DealsHandlers struct {
	log          *slog.Logger
	dealsService *deals.Service
}

func NewDealsHandlers(log *slog.Logger, dealsService *deals.Service) *DealsHandlers {
	return &DealsHandlers{
		dealsService: dealsService,
	}
}

func (h *DealsHandlers) CreateDraft(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CreateDraft"))
	log.Info("handling create draft request")

	var req types.CreateDraftDealRequest
	err := httpx.DecodeJSON(r, &req)
	if err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create draft request")

	authorID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	items := make([]domain.ItemIDsAndQuantities, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.ItemIDsAndQuantities{
			ID:       item.ItemID,
			Quantity: item.Quantity,
		}
	}

	id, err := h.dealsService.CreateDraft(r.Context(), authorID, req.Name, req.Description, items)
	if err != nil {
		log.Error("error creating draft", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, types.CreateDraftDealResponse{Id: id})
}
