package failures

import (
	"barter-port/contracts/openapi/deals/types"
	failuressvc "barter-port/internal/deals/application/failures"
	"barter-port/internal/deals/domain"
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
	dealsService *failuressvc.Service
}

func NewHandlers(log *slog.Logger, dealsService *failuressvc.Service) *Handlers {
	return &Handlers{
		log:          log,
		dealsService: dealsService,
	}
}

func (h *Handlers) GetDealsForFailureReview(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealsForFailureReview"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	deals, err := h.dealsService.GetDealsForFailureReview(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAdminOnly):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error getting failure review deals", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, common.MapDealIDsWithParticipantIDsToDTO(deals))
}

func (h *Handlers) VoteForFailure(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "VoteForFailure"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.VoteForFailureRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	err := h.dealsService.VoteForFailure(r.Context(), dealID, userID, req.UserId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired),
			errors.Is(err, domain.ErrFailureAlreadyResolved):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error voting for failure", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) RevokeVoteForFailure(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "RevokeVoteForFailure"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err := h.dealsService.RevokeVoteForFailure(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired),
			errors.Is(err, domain.ErrFailureAlreadyResolved):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error revoking failure vote", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetFailureVotes(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetFailureVotes"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	votes, err := h.dealsService.GetFailureVotes(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error getting failure votes", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapFailureVotesToDTO(votes))
}

func (h *Handlers) GetFailureMaterials(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetFailureMaterials"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	materials, err := h.dealsService.GetFailureMaterials(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrAdminOnly):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error getting failure materials", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapFailureMaterialsToDTO(materials))
}

func (h *Handlers) ModeratorResolutionForFailure(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "ModeratorResolutionForFailure"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.ModeratorResolutionForFailureRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	record, err := h.dealsService.ModeratorResolutionForFailure(
		r.Context(),
		dealID,
		userID,
		req.Confirmed,
		req.UserId,
		req.PunishmentPoints,
		req.Comment,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidFailureDecision):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrFailureAlreadyResolved):
			httpx.WriteEmptyError(w, http.StatusConflict)
		case errors.Is(err, domain.ErrFailureNotFound),
			errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrAdminOnly):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error resolving failure", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapFailureRecordToDTO(record))
}

func (h *Handlers) GetModeratorResolutionForFailure(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetModeratorResolutionForFailure"))

	dealID, ok := parseFailureDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	record, err := h.dealsService.GetModeratorResolutionForFailure(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error getting moderator resolution for failure", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapFailureRecordToDTO(record))
}

func parseFailureDealID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	dealIDStr := chi.URLParam(r, "dealId")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}

	return dealID, true
}
