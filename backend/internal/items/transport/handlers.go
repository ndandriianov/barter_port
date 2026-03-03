package transport

import (
	"barter-port/internal/items/model"
	"barter-port/internal/items/service"
	"barter-port/internal/libs/platform/http_api"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type itemService interface {
	CreateItem(ctx context.Context, name string, itemType model.ItemType, action model.ItemAction, description string) error
	GetItems(ctx context.Context, nextCursor uuid.UUID, limit int, sortType model.SortType) ([]model.Item, error)
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
			http_api.HandleError(w, log, http.StatusBadRequest, service.ErrInvalidItemName)
			return
		}
		log.Error("failed to create item", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type GetItemRequest struct {
	NextCursor uuid.UUID      `json:"next_cursor"`
	Limit      int            `json:"limit"`
	SortType   model.SortType `json:"sort_type"`
}

func (h *Handlers) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	log := http_api.LogFrom(r.Context(), slog.Default())
	log.Info("handling get items request")

	var req GetItemRequest
	if ok := http_api.DecodeJSONWithLogs(w, r, log, &req); !ok {
		return
	}

	log.Debug("decoded get items request",
		slog.Any("req", req),
	)

	items, err := h.itemService.GetItems(r.Context(), req.NextCursor, req.Limit, req.SortType)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSortType) {
			log.Warn("invalid sort type",
				slog.String("error", err.Error()),
				slog.Any("sort_type", req.SortType),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, service.ErrInvalidSortType)
			return
		}
		log.Error("failed to get items", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http_api.WriteJSON(w, log, http.StatusOK, items)
}
