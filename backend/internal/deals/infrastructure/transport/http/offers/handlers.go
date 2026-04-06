package offers

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/application/offers"
	"barter-port/internal/deals/domain"
	enums "barter-port/internal/deals/domain/enums"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Handlers struct {
	offerService *offers.Service
}

func NewHandlers(offerService *offers.Service) *Handlers {
	return &Handlers{offerService: offerService}
}

// ================================================================================
// CREATE OFFER
// ================================================================================

func (h *Handlers) HandleCreateOffer(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())
	log.Info("handling create offer request")

	var req types.CreateOfferRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create offer request")

	itemType, err := enums.ItemTypeString(string(req.Type))
	if err != nil {
		log.Error("invalid item type", slog.String("type", string(req.Type)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid item type")
		return
	}

	action, err := enums.OfferActionString(string(req.Action))
	if err != nil {
		log.Error("invalid offer action", slog.String("action", string(req.Action)), slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid offer action")
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	offer, err := h.offerService.CreateOffer(r.Context(), userID, req.Name, itemType, action, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOfferName) {
			log.Warn("invalid offer name", slog.Any("error", err))
			httpx.WriteError(w, http.StatusBadRequest, domain.ErrInvalidOfferName)
			return
		}
		log.Error("failed to create offer", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, offer.ToDto())
}

// ================================================================================
// GET OFFERS
// ================================================================================

func (h *Handlers) HandleGetOffers(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default())

	// Parse query parameters
	sortTypeStr := r.URL.Query().Get("sort")
	myStr := r.URL.Query().Get("my")
	createdAtStr := r.URL.Query().Get("cursor_created_at")
	viewsStr := r.URL.Query().Get("cursor_views")
	idStr := r.URL.Query().Get("cursor_id")
	limitStr := r.URL.Query().Get("cursor_limit")

	my := false
	if myStr != "" {
		parsedMy, err := strconv.ParseBool(myStr)
		if err != nil {
			log.Error("invalid my filter", slog.String("my", myStr), slog.Any("error", err))
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid my filter")
			return
		}
		my = parsedMy
	}

	log = log.With(
		slog.String("sortTypeStr", sortTypeStr),
		slog.String("myStr", myStr),
		slog.Bool("my", my),
		slog.String("createdAtStr", createdAtStr),
		slog.String("viewsStr", viewsStr),
		slog.String("idStr", idStr),
		slog.String("limitStr", limitStr),
	)
	log.Info("handling get offers request")

	sortType, err := enums.SortTypeString(sortTypeStr)
	if err != nil {
		log.Error("invalid sort type", slog.Any("error", err))
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid sort type")
		return
	}

	cursor, err := domain.NewUniversalCursor(createdAtStr, viewsStr, idStr)
	if err != nil {
		log.Error("failed to create offers cursor", slog.Any("error", err))
		cursor = nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		log.Warn("invalid limit", slog.Any("error", err))
		limit = 10
	}

	log.Debug("parsing finished", slog.Any("cursor", cursor), slog.Int("limit", limit), slog.Bool("my", my))

	var authorID *uuid.UUID
	if my {
		userID, ok := authkit.UserIDFromContext(r.Context())
		if !ok {
			log.Error("failed to get userID from context")
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
		authorID = &userID
	}

	// Fetch offers from the service
	offerList, nextCursor, err := h.offerService.GetOffers(r.Context(), sortType, cursor, limit, authorID)
	if err != nil {
		log.Error("failed to get items", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	respOffers := make([]types.Offer, len(offerList))
	for i, offer := range offerList {
		respOffers[i] = offer.ToDto()
	}

	var respCursor *types.OffersCursor
	if nextCursor != nil {
		respCursor = new(nextCursor.ToDto())
	}

	httpx.WriteJSON(w, http.StatusOK, types.ListOffersResponse{
		Offers:     respOffers,
		NextCursor: respCursor,
	})
}
