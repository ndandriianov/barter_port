package joins

import (
	joinssvc "barter-port/internal/deals/application/joins"
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
	dealsService *joinssvc.Service
}

func NewHandlers(log *slog.Logger, dealsService *joinssvc.Service) *Handlers {
	return &Handlers{log: log, dealsService: dealsService}
}

func (h *Handlers) JoinDeal(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "JoinDeal"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err := h.dealsService.JoinDeal(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error joining deal", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) LeaveDeal(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "LeaveDeal"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err := h.dealsService.LeaveDeal(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error leaving deal", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetDealJoinRequests(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealJoinRequests"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	items, err := h.dealsService.GetDealJoinRequests(r.Context(), dealID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error getting deal join requests", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapJoinRequestsToDTO(items))
}

func (h *Handlers) ProcessJoinRequest(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "ProcessJoinRequest"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	requestedUserIDStr := chi.URLParam(r, "userId")
	requestedUserID, err := uuid.Parse(requestedUserIDStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	acceptRaw := r.URL.Query().Get("accept")
	if acceptRaw == "" {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}
	accept, err := strconv.ParseBool(acceptRaw)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	voterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err = h.dealsService.ProcessJoinRequest(r.Context(), dealID, requestedUserID, voterID, accept)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrJoinRequestNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrFailureReviewRequired):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("error processing join request", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseDealID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	dealIDStr := chi.URLParam(r, "dealId")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}

	return dealID, true
}
