package transport

import (
	"barter-port/internal/contracts/openapi/items/types"
	"barter-port/internal/items/model"
	"barter-port/internal/items/service"
	"barter-port/internal/libs/platform/http_api"
	"barter-port/internal/libs/platform/logger"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

// ================================================================================
// GET_ITEMS
// ================================================================================

func (h *Handlers) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	// Parse query parameters
	sortTypeStr := r.URL.Query().Get("sort")
	createdAtStr := r.URL.Query().Get("cursor_created_at")
	viewsStr := r.URL.Query().Get("cursor_views")
	idStr := r.URL.Query().Get("cursor_id")
	limitStr := r.URL.Query().Get("limit")

	log = log.With(
		slog.String("sortTypeStr", sortTypeStr),
		slog.String("createdAtStr", createdAtStr),
		slog.String("viewsStr", viewsStr),
		slog.String("idStr", idStr),
		slog.String("limitStr", limitStr),
	)
	log.Info("handling get items request")

	sortType, err := model.SortTypeString(sortTypeStr)
	if err != nil {
		log.Error("invalid sort type", slog.Any("error", err))
		http_api.HandleError(w, log, http.StatusBadRequest, errors.New("invalid sort type"))
		return
	}

	cursor, err := model.NewUniversalCursor(createdAtStr, viewsStr, idStr)
	if err != nil {
		log.Error("failed to create items cursor", slog.Any("error", err))
		http_api.HandleError(w, log, http.StatusBadRequest, errors.New("failed to create items cursor"))
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Warn("invalid limit", slog.Any("error", err))
		limit = 10
	}

	log.Debug("parsing finished", slog.Any("cursor", cursor), slog.Int("limit", limit))

	// Fetch items from the service
	items, nextCursor, err := h.itemService.GetItems(r.Context(), sortType, *cursor, limit)
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

	respItems := make([]types.Item, len(items))
	for i, item := range items {
		respItems[i] = types.Item{
			Action:      types.ItemAction(item.Action.String()),
			CreatedAt:   item.CreatedAt,
			Description: item.Description,
			Id:          item.ID,
			Name:        item.Name,
			Type:        types.ItemType(item.Type.String()),
			Views:       int64(item.Views),
		}
	}

	var viewsPtr *int64 = nil
	if nextCursor.Views != nil {
		viewsPtr = new(int64(*nextCursor.Views))
	}

	respCursor := types.ItemsCursor{
		CreatedAt: nextCursor.CreatedAt,
		Id:        nextCursor.Id,
		Views:     viewsPtr,
	}

	http_api.WriteJSON(w, log, http.StatusOK, types.ListItemsResponse{
		Items:      respItems,
		NextCursor: &respCursor,
	})
}
