package transport

import (
	"barter-port/internal/items/model"
	"barter-port/internal/items/service"
	"barter-port/internal/libs/platform/http_api"
	"barter-port/internal/libs/platform/logger"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	Name        string           `json:"name" example:"name"`
	Type        model.ItemType   `json:"type" swaggertype:"string" enums:"good,service"`
	Action      model.ItemAction `json:"action" swaggertype:"string" enums:"give,take"`
	Description string           `json:"description" example:"description"`
}

// HandleCreateItem handles the creation of a new item.
// @Security BearerAuth
// @Summary Create a new item
// @Description Create a new item with the provided details
// @Tags items
// @Accept json
// @Param request body CreateItemRequest true "Create Item Request"
// @Success 201
// @Failure 400
// @Failure 500
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
			w.WriteHeader(http.StatusBadRequest)
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
	SortType model.SortType        `json:"sort_type"`          // Sorting type for items
	Cursor   model.UniversalCursor `json:"cursor"`             // Cursor for pagination
	Limit    int                   `json:"limit" example:"10"` // Maximum number of items to fetch
}

// GetItemResponse represents the response payload for fetching items.
// swagger:model GetItemResponse
type GetItemResponse struct {
	Items  []model.Item          `json:"items"`  // List of items
	Cursor model.UniversalCursor `json:"cursor"` // Cursor for the next page
}

// HandleGetItems handles fetching a list of items.
// @Security BearerAuth
// @Summary Get a list of items
// @Description Fetch a list of items with optional sorting and pagination
// @Tags items
// @Produce json
// @Param sort_type query string true "Sort type (ByTime, ByPopularity)"
// @Param created_at query string false "Creation time for cursor (ISO 8601 format)"
// @Param views query int false "Views for cursor"
// @Param id query string false "UUID for cursor"
// @Param limit query int false "Maximum number of items to fetch"
// @Success 200 {object} GetItemResponse "List of items and pagination cursor"
// @Failure 400 {object} http_api.ErrorResponse "Invalid input"
// @Failure 500
// @Router /items [get]
func (h *Handlers) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling get items request")

	// Parse query parameters

	sortTypeStr := r.URL.Query().Get("sort_type")
	createdAtStr := r.URL.Query().Get("created_at")
	viewsStr := r.URL.Query().Get("views")
	idStr := r.URL.Query().Get("id")
	limitStr := r.URL.Query().Get("limit")

	var sortType model.SortType
	if sortTypeStr != "" {
		var err error
		sortType, err = model.SortTypeString(sortTypeStr)
		if err != nil {
			log.Warn("invalid sort type",
				slog.String("error", err.Error()),
				slog.String("sort_type", sortTypeStr),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid sort type"))
			return
		}
	}

	var cursor model.UniversalCursor
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Warn("invalid created_at",
				slog.String("error", err.Error()),
				slog.String("created_at", createdAtStr),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid created_at"))
			return
		}
		cursor.CreatedAt = &createdAt
	}

	if viewsStr != "" {
		views, err := strconv.Atoi(viewsStr)
		if err != nil {
			log.Warn("invalid views",
				slog.String("error", err.Error()),
				slog.String("views", viewsStr),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid views"))
			return
		}
		cursor.Views = &views
	}

	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Warn("invalid id",
				slog.String("error", err.Error()),
				slog.String("id", idStr),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid id"))
			return
		}
		cursor.Id = id
	}

	limit := 10 // Default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			log.Warn("invalid limit",
				slog.String("error", err.Error()),
				slog.String("limit", limitStr),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid limit"))
			return
		}
	}

	log.Debug("parsed get items request",
		slog.String("sort_type", sortType.String()),
		slog.Any("cursor", cursor),
		slog.Int("limit", limit),
	)

	// Fetch items from the service

	items, nextCursor, err := h.itemService.GetItems(r.Context(), sortType, cursor, limit)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSortType) {
			log.Warn("invalid sort type",
				slog.String("error", err.Error()),
				slog.Any("sort_type", sortType),
			)
			http_api.HandleError(w, log, http.StatusBadRequest, service.ErrInvalidSortType)
			return
		}
		log.Error("failed to get items", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http_api.WriteJSON(w, log, http.StatusOK, GetItemResponse{Items: items, Cursor: nextCursor})
}
