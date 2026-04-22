package favourites

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
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	offersService *offersapp.Service
}

func NewHandlers(offersService *offersapp.Service) *Handlers {
	return &Handlers{offersService: offersService}
}

func (h *Handlers) HandleListFavoriteOffers(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	params, err := decodeListFavoriteOffersRequest(r)
	if err != nil {
		log.Error("failed to decode favorite offers request", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, err.Error())
		return
	}

	cursor, err := newFavoriteCursorFromParams(params.CursorFavoritedAt, params.CursorId)
	if err != nil {
		log.Error("failed to parse favorite offers cursor", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid cursor")
		return
	}

	limit := 10
	if params.CursorLimit != nil {
		limit = *params.CursorLimit
		if limit <= 0 {
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid limit")
			return
		}
	}

	offers, nextCursor, err := h.offersService.GetFavoriteOffers(r.Context(), userID, cursor, limit)
	if err != nil {
		log.Error("failed to get favorite offers", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	respOffers := make([]types.FavoritedOffer, len(offers))
	for i, offer := range offers {
		respOffers[i] = offer.ToDto()
	}

	var respCursor *types.FavoriteOffersCursor
	if nextCursor != nil {
		respCursor = new(nextCursor.ToDto())
	}

	httpx.WriteJSON(w, http.StatusOK, types.ListFavoriteOffersResponse{
		Offers:     respOffers,
		NextCursor: respCursor,
	})
}

func (h *Handlers) HandleAddOfferToFavorites(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	if err := h.offersService.AddOfferToFavorites(r.Context(), userID, offerID); err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("failed to add offer to favorites", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) HandleRemoveOfferFromFavorites(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	if err := h.offersService.RemoveOfferFromFavorites(r.Context(), userID, offerID); err != nil {
		log.Error("failed to remove offer from favorites", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeListFavoriteOffersRequest(r *http.Request) (types.ListFavoriteOffersParams, error) {
	params := types.ListFavoriteOffersParams{}
	query := r.URL.Query()

	if rawFavoritedAt := query.Get("cursor_favorited_at"); rawFavoritedAt != "" {
		value, err := time.Parse(time.RFC3339, rawFavoritedAt)
		if err != nil {
			return types.ListFavoriteOffersParams{}, errors.New("invalid cursor_favorited_at")
		}
		params.CursorFavoritedAt = new(value)
	}
	if rawID := query.Get("cursor_id"); rawID != "" {
		value, err := uuid.Parse(rawID)
		if err != nil {
			return types.ListFavoriteOffersParams{}, errors.New("invalid cursor_id")
		}
		params.CursorId = new(value)
	}
	if rawLimit := query.Get("cursor_limit"); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil {
			return types.ListFavoriteOffersParams{}, errors.New("invalid cursor_limit")
		}
		params.CursorLimit = new(value)
	}

	return params, nil
}

func newFavoriteCursorFromParams(favoritedAt *types.CursorFavoritedAt, id *types.CursorId) (*domain.FavoriteOffersCursor, error) {
	if favoritedAt == nil && id == nil {
		return nil, nil
	}
	if favoritedAt == nil || id == nil {
		return nil, errors.New("cursor requires cursor_favorited_at and cursor_id")
	}

	return &domain.FavoriteOffersCursor{
		FavoritedAt: *favoritedAt,
		Id:          *id,
	}, nil
}

func parseOfferID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "offerId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
