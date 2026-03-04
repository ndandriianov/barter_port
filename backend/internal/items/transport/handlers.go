package transport

import (
	"barter-port/internal/items/model"
	"barter-port/internal/items/service"
	"barter-port/internal/libs/platform/http_api"
	"barter-port/internal/libs/platform/logger"
	"errors"
	"log/slog"
	"net/http"

	"golang.org/x/net/context"
)

type itemService interface {
	CreateItem(ctx context.Context, name string, itemType model.ItemType, action model.ItemAction, description string) error
	GetItems(ctx context.Context, sortType model.SortType, cursor model.UniversalCursor, limit int) ([]model.Item, model.UniversalCursor, error)
}

type Handlers struct {
	itemService itemService
}

func NewHandlers(itemService itemService) *Handlers {
	return &Handlers{itemService: itemService}
}

//
// CREATE_ITEM request and response structures and handler
//

// CreateItemRequest represents the request payload for creating an item.
// swagger:model CreateItemRequest
type CreateItemRequest struct {
	Name        string           `json:"name"`        // Name of the item
	Type        model.ItemType   `json:"type"`        // Type of the item
	Action      model.ItemAction `json:"action"`      // Action associated with the item
	Description string           `json:"description"` // Description of the item
}

// HandleCreateItem handles the creation of a new item.
// @Summary Create a new item
// @Description Create a new item with the provided details
// @Tags items
// @Accept json
// @Produce json
// @Param request body CreateItemRequest true "Create Item Request"
// @Success 201 {string} string "Created"
// @Failure 400 {object} http_api.ErrorResponse "Invalid input"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /items [post]
func (h *Handlers) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
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

//
// GET_ITEMS request and response structures and handler
//

// GetItemRequest represents the request payload for fetching items.
// swagger:model GetItemRequest
type GetItemRequest struct {
	SortType model.SortType        `json:"sort_type"` // Sorting type for items
	Cursor   model.UniversalCursor `json:"cursor"`    // Cursor for pagination
	Limit    int                   `json:"limit"`     // Maximum number of items to fetch
}

// GetItemResponse represents the response payload for fetching items.
// swagger:model GetItemResponse
type GetItemResponse struct {
	Items  []model.Item          `json:"items"`  // List of items
	Cursor model.UniversalCursor `json:"cursor"` // Cursor for the next page
}

// HandleGetItems handles fetching a list of items.
// @Summary Get a list of items
// @Description Fetch a list of items with optional sorting and pagination
// @Tags items
// @Accept json
// @Produce json
// @Param request body GetItemRequest true "Get Items Request"
// @Success 200 {object} GetItemResponse "List of items"
// @Failure 400 {object} http_api.ErrorResponse "Invalid input"
// @Failure 500 {object} http_api.ErrorResponse "Internal server error"
// @Router /items [get]
func (h *Handlers) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling get items request")

	var req GetItemRequest
	if ok := http_api.DecodeJSONWithLogs(w, r, log, &req); !ok {
		return
	}

	log.Debug("decoded get items request",
		slog.Any("req", req),
	)

	items, cursor, err := h.itemService.GetItems(r.Context(), req.SortType, req.Cursor, req.Limit)
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

	http_api.WriteJSON(w, log, http.StatusOK, GetItemResponse{Items: items, Cursor: cursor})
}
