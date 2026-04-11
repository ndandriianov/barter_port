package deals

import (
	"barter-port/contracts/openapi/deals/types"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/internal/deals/infrastructure/transport/http/common"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	log          *slog.Logger
	dealsService *dealssvc.Service
}

func NewHandlers(log *slog.Logger, dealsService *dealssvc.Service) *Handlers {
	return &Handlers{
		log:          log,
		dealsService: dealsService,
	}
}

// ================================================================================
// GET DEALS
// ================================================================================

func (h *Handlers) GetDeals(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDeals"))
	log.Info("handling get deals request")

	my := r.URL.Query().Get("my") == "true"
	open := r.URL.Query().Get("open") == "true"

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	deals, err := h.dealsService.GetDeals(r.Context(), userID, my, open)
	if err != nil {
		log.Error("error getting deals", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealIDsWithParticipantIDsToDTO(deals))
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

func (h *Handlers) GetDealByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealByID"))
	log.Info("handling get deal by id request")

	idStr := chi.URLParam(r, "dealId")
	if idStr == "" {
		log.Error("dealId is required")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	deal, err := h.dealsService.GetDealByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDealNotFound) {
			log.Info("deal not found", slog.String("dealId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}
		log.Error("error getting deal", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealToDTO(deal))
}

// ================================================================================
// UPDATE DEAL
// ================================================================================

func (h *Handlers) UpdateDeal(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "UpdateDeal"))
	log.Info("handling update deal request")

	idStr := chi.URLParam(r, "dealId")
	if idStr == "" {
		log.Error("dealId is required")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	dealID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.UpdateDealRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	deal, err := h.dealsService.UpdateDealName(r.Context(), dealID, userID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			log.Info("deal not found", slog.String("dealId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		case errors.Is(err, domain.ErrForbidden):
			log.Info("user not a participant", slog.String("dealId", idStr), slog.String("userID", userID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		default:
			log.Error("error updating deal", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealToDTO(deal))
}

// ================================================================================
// ADD DEAL ITEM
// ================================================================================

func (h *Handlers) AddDealItem(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "AddDealItem"))
	log.Info("handling add deal item request")

	dealIDStr := chi.URLParam(r, "dealId")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.AddDealItemRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	deal, err := h.dealsService.AddDealItem(r.Context(), userID, dealID, req.OfferId, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidQuantity):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error adding deal item", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealToDTO(deal))
}

// ================================================================================
// UPDATE DEAL ITEM
// ================================================================================

func (h *Handlers) UpdateDealItem(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "UpdateDealItem"))
	log.Info("handling update deal item request")

	dealIDStr := chi.URLParam(r, "dealId")
	itemIDStr := chi.URLParam(r, "itemId")

	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		log.Error("error parsing item id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	var req types.UpdateDealItemRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	claimProvider := req.ClaimProvider != nil && *req.ClaimProvider
	releaseProvider := req.ReleaseProvider != nil && *req.ReleaseProvider
	claimReceiver := req.ClaimReceiver != nil && *req.ClaimReceiver
	releaseReceiver := req.ReleaseReceiver != nil && *req.ReleaseReceiver

	hasContent := req.Name != nil || req.Description != nil || req.Quantity != nil
	hasRole := claimProvider || releaseProvider || claimReceiver || releaseReceiver
	if !hasContent && !hasRole {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	patch := htypes.ItemPatch{
		Name:            req.Name,
		Description:     req.Description,
		Quantity:        req.Quantity,
		ClaimProvider:   claimProvider,
		ReleaseProvider: releaseProvider,
		ClaimReceiver:   claimReceiver,
		ReleaseReceiver: releaseReceiver,
	}

	item, err := h.dealsService.UpdateDealItem(r.Context(), userID, dealID, itemID, patch)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrItemNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidDealStatus):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrFailureReviewRequired),
			errors.Is(err, domain.ErrRoleAlreadyTaken),
			errors.Is(err, domain.ErrNotRoleHolder),
			errors.Is(err, domain.ErrDuplicateRole):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error updating deal item", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, item.ToDTO())
}

// ================================================================================
// CHANGE DEAL STATUS
// ================================================================================

func (h *Handlers) ChangeDealStatus(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "ChangeDealStatus"))
	log.Info("handling change deal status request")

	idStr := chi.URLParam(r, "dealId")
	dealID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.ChangeDealStatusRequest
	if err = httpx.DecodeJSON(r, &req); err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	targetStatus, err := mapDealStatusFromDTO(req.ExpectedStatus)
	if err != nil {
		log.Error("unknown deal status", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	deal, err := h.dealsService.ProcessDealStatusUpdateRequest(r.Context(), dealID, userID, targetStatus)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidDealStatus):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		case errors.Is(err, domain.ErrDealParticipantsUnready),
			errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrFailureReviewRequired):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error changing deal status", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealToDTO(deal))
}

// ================================================================================
// GET DEAL STATUS VOTES
// ================================================================================

func (h *Handlers) GetDealStatusVotes(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealStatusVotes"))
	log.Info("handling get deal status votes request")

	idStr := chi.URLParam(r, "dealId")
	dealID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing deal id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	if _, ok := authkit.UserIDFromContext(r.Context()); !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	votes, err := h.dealsService.GetDealStatusVotes(r.Context(), dealID)
	if err != nil {
		if errors.Is(err, domain.ErrDealNotFound) {
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}

		log.Error("error getting deal status votes", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapStatusVotesToDTO(votes))
}
