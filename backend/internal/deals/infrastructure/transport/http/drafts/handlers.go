package drafts

import (
	"barter-port/contracts/openapi/deals/types"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
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
// CREATE DRAFT
// ================================================================================

func (h *Handlers) CreateDraft(w http.ResponseWriter, r *http.Request) {
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

	if err = h.dealsService.EnsureRequesterCanCreateDraftDeal(r.Context(), authorID, offers); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			log.Warn("draft creation is forbidden by hidden users policy", slog.String("author_id", authorID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		}
		log.Error("error checking hidden users policy for draft creation", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	id, err := h.dealsService.CreateDraft(r.Context(), authorID, req.Name, req.Description, offers, nil)
	if err != nil {
		if errors.Is(err, domain.ErrNoOffers) {
			log.Warn("no offers in request")
			httpx.WriteError(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			log.Warn("forbidden to create draft", slog.String("author_id", authorID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
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

func (h *Handlers) GetDrafts(w http.ResponseWriter, r *http.Request) {
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

func (h *Handlers) GetDraftByID(w http.ResponseWriter, r *http.Request) {
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
// DELETE DRAFT
// ================================================================================

func (h *Handlers) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "DeleteDraft"))
	log.Info("handling delete draft request")

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

	err = h.dealsService.DeleteDraftByID(r.Context(), draftID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDraftNotFound):
			log.Info("draft not found", slog.String("draftId", idStr))
			httpx.WriteEmptyError(w, http.StatusNotFound)
			return
		case errors.Is(err, domain.ErrForbidden):
			log.Info("user is not in draft", slog.String("draftId", idStr), slog.String("userID", userID.String()))
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		default:
			log.Error("error deleting draft", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// ================================================================================
// CONFIRM DRAFT
// ================================================================================

func (h *Handlers) ConfirmDraft(w http.ResponseWriter, r *http.Request) {
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

func (h *Handlers) CancelDraft(w http.ResponseWriter, r *http.Request) {
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
