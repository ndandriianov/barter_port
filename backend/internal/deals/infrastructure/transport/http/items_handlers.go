package http

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type itemService interface {
	CreateItem(ctx context.Context, userID uuid.UUID, name string, itemType domain.ItemType, action domain.ItemAction, description string) (*domain.Item, error)
	GetItems(ctx context.Context, sortType domain.SortType, cursor *domain.UniversalCursor, limit int) ([]domain.Item, *domain.UniversalCursor, error)
}

type ItemsHandlers struct {
	itemService itemService
}

func NewHandlers(itemService itemService) *ItemsHandlers {
	return &ItemsHandlers{itemService: itemService}
}

// ================================================================================
// CREATE ITEM
// ================================================================================

func (h *ItemsHandlers) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling register request")

	var req types.CreateItemRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create item request")

	itemType, err := domain.ItemTypeString(string(req.Type))
	if err != nil {
		log.Error("invalid item type", slog.String("type", string(req.Type)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid item type")
		return
	}

	action, err := domain.ItemActionString(string(req.Action))
	if err != nil {
		log.Error("invalid item action", slog.String("action", string(req.Action)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid item action")
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	item, err := h.itemService.CreateItem(r.Context(), userID, req.Name, itemType, action, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidItemName) {
			log.Warn("invalid item name", slog.Any("error", err))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidItemName)
			return
		}
		log.Error("failed to create item", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, item.ToDto())
}

// ================================================================================
// GET ITEMS
// ================================================================================

func (h *ItemsHandlers) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	// Parse query parameters
	sortTypeStr := r.URL.Query().Get("sort")
	createdAtStr := r.URL.Query().Get("cursor_created_at")
	viewsStr := r.URL.Query().Get("cursor_views")
	idStr := r.URL.Query().Get("cursor_id")
	limitStr := r.URL.Query().Get("cursor_limit")

	log = log.With(
		slog.String("sortTypeStr", sortTypeStr),
		slog.String("createdAtStr", createdAtStr),
		slog.String("viewsStr", viewsStr),
		slog.String("idStr", idStr),
		slog.String("limitStr", limitStr),
	)
	log.Info("handling get items request")

	sortType, err := domain.SortTypeString(sortTypeStr)
	if err != nil {
		log.Error("invalid sort type", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid sort type")
		return
	}

	cursor, err := domain.NewUniversalCursor(createdAtStr, viewsStr, idStr)
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
		log.Error("failed to get items", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	respItems := make([]types.Item, len(items))
	for i, item := range items {
		respItems[i] = item.ToDto()
	}

	var respCursor *types.ItemsCursor
	if nextCursor != nil {
		respCursor = new(nextCursor.ToDto())
	}

	httpx.WriteJSON(w, http.StatusOK, types.ListItemsResponse{
		Items:      respItems,
		NextCursor: respCursor,
	})
}
