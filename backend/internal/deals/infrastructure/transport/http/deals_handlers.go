package http

import (
	"barter-port/contracts/openapi/deals/types"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DealsHandlers struct {
	log          *slog.Logger
	dealsService *dealssvc.Service
}

func NewDealsHandlers(log *slog.Logger, dealsService *dealssvc.Service) *DealsHandlers {
	return &DealsHandlers{
		log:          log,
		dealsService: dealsService,
	}
}

// ================================================================================
// CREATE DRAFT
// ================================================================================

func (h *DealsHandlers) CreateDraft(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CreateDraft"))
	log.Info("handling create draft request")

	var req types.CreateDraftDealRequest
	err := httpx.DecodeJSON(r, &req)
	if err != nil {
		log.Error("error decoding request", slog.Any("error", err))
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	log = log.With(slog.Any("request", req))
	log.Debug("decoded create draft request")

	authorID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offers := make([]domain.OfferIDAndInfo, len(req.Offers))
	for i, item := range req.Offers {
		offers[i] = domain.OfferIDAndInfo{
			ID: item.OfferID,
			Info: domain.OfferInfo{
				Quantity: item.Quantity,
			},
		}
	}

	id, err := h.dealsService.CreateDraft(r.Context(), authorID, req.Name, req.Description, offers)
	if err != nil {
		if errors.Is(err, domain.ErrNoOffers) {
			log.Warn("no offers in request")
			httpx.WriteError(w, http.StatusBadRequest, err)
			return
		}
		log.Error("error creating draft", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, types.CreateDraftDealResponse{Id: id})
}

// ================================================================================
// GET DRAFTS
// ================================================================================

func (h *DealsHandlers) GetDrafts(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDrafts"))
	log.Info("handling get drafts request")

	createdByMe := false
	createdByMeStr := r.URL.Query().Get("createdByMe")
	if createdByMeStr != "" {
		parsed, err := strconv.ParseBool(createdByMeStr)
		if err != nil {
			log.Warn("invalid createdByMe query param", slog.String("createdByMe", createdByMeStr), slog.Any("error", err))
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid createdByMe")
			return
		}
		createdByMe = parsed
	}

	participating := true
	participatingStr := r.URL.Query().Get("participating")
	if participatingStr != "" {
		parsed, err := strconv.ParseBool(participatingStr)
		if err != nil {
			log.Warn("invalid participating query param", slog.String("participating", participatingStr), slog.Any("error", err))
			httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid participating")
			return
		}
		participating = parsed
	}

	_ = participating // TODO: если админ и false, то выдать все

	authorID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	log.Debug("parsed query params",
		slog.String("createdByMeStr", createdByMeStr),
		slog.Bool("createdByMe", createdByMe),
		slog.String("participatingStr", participatingStr),
		slog.Bool("participating", participating),
	)

	// вызов сервиса
	draftsIDs, err := h.dealsService.GetDraftsByAuthor(r.Context(), authorID, createdByMe)
	if err != nil {
		log.Error("error getting drafts", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapDraftIDsWithAuthorIDsToDTO(draftsIDs))
}

// ================================================================================
// GET DRAFT BY ID
// ================================================================================

func (h *DealsHandlers) GetDraftByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDraftByID"))
	log.Info("handling get draft by id request")

	idStr := chi.URLParam(r, "draftId")
	if idStr == "" {
		log.Error("draftId is required")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing draft id")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	draft, err := h.dealsService.GetDraftByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDraftNotFound) {
			log.Info("draft not found", slog.String("draftId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		}
		log.Error("error getting draft", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, draft.ToDTO())
}

// ================================================================================
// CONFIRM DRAFT
// ================================================================================

func (h *DealsHandlers) ConfirmDraft(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "ConfirmDraft"))
	log.Info("handling confirm draft request")

	idStr := chi.URLParam(r, "draftId")
	if idStr == "" {
		log.Error("draftId is required")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	draftID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing draft id", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	users, err := h.dealsService.ConfirmDraft(r.Context(), draftID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDraftNotFound):
			log.Info("draft not found", slog.String("draftId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		case errors.Is(err, domain.ErrUserNotInDraft):
			log.Info("user is not in draft", slog.String("draftId", idStr), slog.String("userID", userID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		default:
			log.Error("error confirming draft", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
	}

	respUsers := make([]types.UserConfirm, 0, len(users))

	for _, user := range users {
		respUsers = append(respUsers, types.UserConfirm{
			Confirmed: user.Confirmed,
			UserId:    user.UserID,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, types.ConfirmDraftDealResponse{Users: respUsers})
}

// ================================================================================
// GET DEALS
// ================================================================================

func (h *DealsHandlers) GetDeals(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDeals"))
	log.Info("handling get deals request")

	my := r.URL.Query().Get("my") == "true"
	// TODO: open deals filtering is not yet supported (no status field in schema)
	_ = r.URL.Query().Get("open")

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	deals, err := h.dealsService.GetDeals(r.Context(), userID, my)
	if err != nil {
		log.Error("error getting deals", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapDealIDsWithParticipantIDsToDTO(deals))
}

// ================================================================================
// GET DEAL BY ID
// ================================================================================

func (h *DealsHandlers) GetDealByID(w http.ResponseWriter, r *http.Request) {
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

	httpx.WriteJSON(w, http.StatusOK, mapDealToDTO(deal))
}

// ================================================================================
// UPDATE DEAL ITEM
// ================================================================================

func (h *DealsHandlers) UpdateDealItem(w http.ResponseWriter, r *http.Request) {
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
		case errors.Is(err, domain.ErrForbidden),
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
// CANCEL DRAFT
// ================================================================================

func (h *DealsHandlers) CancelDraft(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CancelDraft"))
	log.Info("handling cancel draft request")

	idStr := chi.URLParam(r, "draftId")
	if idStr == "" {
		log.Error("draftId is required")
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	draftID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error("error parsing draft id", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err = h.dealsService.CancelDraft(r.Context(), draftID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDraftNotFound):
			log.Info("draft not found", slog.String("draftId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		case errors.Is(err, domain.ErrUserNotInDraft):
			log.Info("user is not in draft", slog.String("draftId", idStr), slog.String("userID", userID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		default:
			log.Error("error cancelling draft", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
