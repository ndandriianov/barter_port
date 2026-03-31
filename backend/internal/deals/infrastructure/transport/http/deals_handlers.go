package http

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DealsHandlers struct {
	log          *slog.Logger
	dealsService *deals.Service
}

func NewDealsHandlers(log *slog.Logger, dealsService *deals.Service) *DealsHandlers {
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
// GET MY DRAFTS
// ================================================================================

func (h *DealsHandlers) GetMyDrafts(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetMyDrafts"))
	log.Info("handling get my draft request")

	authorID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	draftsIDs, err := h.dealsService.GetDraftIDsByAuthor(r.Context(), authorID)
	if err != nil {
		log.Error("error getting drafts", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, draftsIDs)
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
