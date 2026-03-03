package transport

import (
	"barter-port/internal/items/model"
	"barter-port/internal/items/service"
	"barter-port/internal/libs/platform/http_api"
	"errors"
	"log/slog"
	"net/http"

	"golang.org/x/net/context"
)

type itemService interface {
	CreateItem(ctx context.Context, name string, itemType model.ItemType, action model.ItemAction, description string) error
	GetItems(ctx context.Context, query model.ItemQuery) ([]model.Item, error)
}

type Handlers struct {
	itemService itemService
}

func NewHandlers(itemService itemService) *Handlers {
	return &Handlers{itemService: itemService}
}

type CreateItemRequest struct {
	Name        string           `json:"name"`
	Type        model.ItemType   `json:"type"`
	Action      model.ItemAction `json:"action"`
	Description string           `json:"description"`
}

func (h *Handlers) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	log := http_api.LogFrom(r.Context(), slog.Default())
	log.Info("handling register request")

	var req CreateItemRequest
	if ok := http_api.DecodeJSONWithLogs(w, r, log, &req); !ok {
		return
	}

	log.Debug("decoded create item request",
		slog.Any("req", req),
	)

	err := h.itemService.CreateItem(r.Context(), req.Name, req.Type, req.Action, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrInvalidItemName) {
			log.Warn("invalid item name",
				slog.String("error", err.Error()),
				slog.String("name", req.Name),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, err)
			return
		}
		log.Error("failed to create item", slog.String("error", err.Error()))
		http_api.HandleError(w, log, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
