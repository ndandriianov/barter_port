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
	CreateItem(ctx context.Context, name string, itemType model.ItemType, action model.ItemAction, description string) (*model.Item, error)
	GetItems(ctx context.Context, sortType model.SortType, cursor *model.UniversalCursor, limit int) ([]model.Item, model.UniversalCursor, error)
}

type Handlers struct {
	itemService itemService
}

func NewHandlers(itemService itemService) *Handlers {
	return &Handlers{itemService: itemService}
}

// ================================================================================
// CREATE ITEM
// ================================================================================

func (h *Handlers) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling register request")

	var req types.CreateItemRequest
	if err := http_api.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		http_api.WriteJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Вы отправили некорректный запрос",
		})
		return
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create item request")

	itemType, err := model.ItemTypeString(string(req.Type))
	if err != nil {
		log.Error("invalid item type", slog.String("type", string(req.Type)), slog.Any("error", err))
		http_api.WriteJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Code:    "INVALID_ITEM_TYPE",
			Message: "Невозможно создать объявление с таким типом",
		})
	}

	action, err := model.ItemActionString(string(req.Action))
	if err != nil {
		log.Error("invalid item action", slog.String("action", string(req.Action)), slog.Any("error", err))
		http_api.WriteJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Code:    "INVALID_ITEM_ACTION",
			Message: "Невозможно создать объявление с таким действием",
		})
	}

	item, err := h.itemService.CreateItem(r.Context(), req.Name, itemType, action, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrInvalidItemName) {
			log.Warn("invalid item name", slog.String("error", err.Error()))
			http_api.WriteJSON(w, http.StatusBadRequest, types.ErrorResponse{
				Code:    "INVALID_ITEM_NAME",
				Message: "Некорректное название объявления",
			})
			return
		}
		log.Error("failed to create item", slog.String("error", err.Error()))
		http_api.WriteJSON(w, http.StatusInternalServerError, types.ErrorResponse{
			Code:    "INTERNAL",
			Message: "Произошла ошибка, повторите ошибку позднее",
		})
		return
	}

	http_api.WriteJSON(w, http.StatusCreated, item.ToDto())
}

// ================================================================================
// GET ITEMS
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
		http_api.WriteJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Code:    "INVALID_SORT_TYPE",
			Message: "Вы указали несуществующий тип сортировки",
		})
		return
	}

	cursor, err := model.NewUniversalCursor(createdAtStr, viewsStr, idStr)
	if err != nil {
		log.Error("failed to create items cursor", slog.Any("error", err))
		cursor = nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Warn("invalid limit", slog.Any("error", err))
		limit = 10
	}

	log.Debug("parsing finished", slog.Any("cursor", cursor), slog.Int("limit", limit))

	// Fetch items from the service
	items, nextCursor, err := h.itemService.GetItems(r.Context(), sortType, cursor, limit)
	if err != nil {
		log.Error("failed to get items", slog.String("error", err.Error()))
		http_api.WriteJSON(w, http.StatusInternalServerError, types.ErrorResponse{
			Code:    "INTERNAL",
			Message: "Произошла ошибка, повторите ошибку позднее",
		})
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

	http_api.WriteJSONWithLogs(w, log, http.StatusOK, types.ListItemsResponse{
		Items:      respItems,
		NextCursor: &respCursor,
	})
}
