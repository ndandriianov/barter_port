package reviews

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"barter-port/contracts/openapi/deals/types"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/domain"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
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
// OFFER REVIEWS
// ================================================================================

func (h *Handlers) GetOfferReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetOfferReviews"))

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	reviews, err := h.dealsService.GetOfferReviews(r.Context(), offerID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting offer reviews", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewsToDTO(reviews))
}

func (h *Handlers) GetOfferReviewsSummary(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetOfferReviewsSummary"))

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	summary, err := h.dealsService.GetOfferReviewsSummary(r.Context(), offerID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting offer reviews summary", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewSummaryToDTO(summary))
}

// ================================================================================
// PROVIDER REVIEWS
// ================================================================================

func (h *Handlers) GetProviderReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetProviderReviews"))

	providerID, ok := parseProviderID(w, r)
	if !ok {
		return
	}

	reviews, err := h.dealsService.GetProviderReviews(r.Context(), providerID)
	if err != nil {
		log.Error("error getting provider reviews", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewsToDTO(reviews))
}

func (h *Handlers) GetProviderReviewsSummary(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetProviderReviewsSummary"))

	providerID, ok := parseProviderID(w, r)
	if !ok {
		return
	}

	summary, err := h.dealsService.GetProviderReviewsSummary(r.Context(), providerID)
	if err != nil {
		log.Error("error getting provider reviews summary", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewSummaryToDTO(summary))
}

// ================================================================================
// AUTHOR REVIEWS
// ================================================================================

func (h *Handlers) GetAuthorReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetAuthorReviews"))

	authorID, ok := parseAuthorID(w, r)
	if !ok {
		return
	}

	reviews, err := h.dealsService.GetAuthorReviews(r.Context(), authorID)
	if err != nil {
		log.Error("error getting author reviews", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewsToDTO(reviews))
}

// ================================================================================
// SINGLE REVIEW OPERATIONS
// ================================================================================

func (h *Handlers) GetReviewByID(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetReviewByID"))

	reviewID, ok := parseReviewID(w, r)
	if !ok {
		return
	}

	review, err := h.dealsService.GetReviewByID(r.Context(), reviewID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting review by id", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, review.ToDTO())
}

func (h *Handlers) UpdateReview(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "UpdateReview"))

	reviewID, ok := parseReviewID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.UpdateReviewRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if req.Rating == nil && req.Comment == nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	review, err := h.dealsService.UpdateReview(r.Context(), userID, reviewID, req.Rating, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrInvalidDealStatus):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		default:
			log.Error("error updating review", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, review.ToDTO())
}

func (h *Handlers) DeleteReview(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "DeleteReview"))

	reviewID, ok := parseReviewID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	err := h.dealsService.DeleteReview(r.Context(), userID, reviewID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReviewNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrInvalidDealStatus):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		default:
			log.Error("error deleting review", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ================================================================================
// DEAL REVIEWS
// ================================================================================

func (h *Handlers) GetDealReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealReviews"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	reviews, err := h.dealsService.GetDealReviews(r.Context(), dealID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting deal reviews", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewsToDTO(reviews))
}

func (h *Handlers) GetDealPendingReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealPendingReviews"))

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

	pending, err := h.dealsService.GetDealPendingReviews(r.Context(), userID, dealID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrInvalidDealStatus):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		default:
			log.Error("error getting deal pending reviews", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapPendingReviewsToDTO(pending))
}

// ================================================================================
// DEAL ITEM REVIEWS
// ================================================================================

func (h *Handlers) GetDealItemReviewEligibility(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealItemReviewEligibility"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	itemID, ok := parseItemID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	eligibility, err := h.dealsService.GetDealItemReviewEligibility(r.Context(), userID, dealID, itemID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrItemNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting deal item review eligibility", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapEligibilityToDTO(eligibility))
}

func (h *Handlers) GetDealItemReviews(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "GetDealItemReviews"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	itemID, ok := parseItemID(w, r)
	if !ok {
		return
	}

	reviews, err := h.dealsService.GetDealItemReviews(r.Context(), dealID, itemID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrItemNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("error getting deal item reviews", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, mapReviewsToDTO(reviews))
}

func (h *Handlers) CreateDealItemReview(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "CreateDealItemReview"))

	dealID, ok := parseDealID(w, r)
	if !ok {
		return
	}

	itemID, ok := parseItemID(w, r)
	if !ok {
		return
	}

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		log.Error("failed to get userID from context")
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var req types.CreateReviewRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCannotDecodeRequestBody)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	review, err := h.dealsService.CreateDealItemReview(r.Context(), userID, dealID, itemID, req.Rating, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDealNotFound), errors.Is(err, domain.ErrItemNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrInvalidDealStatus),
			errors.Is(err, domain.ErrReceiverMissing),
			errors.Is(err, domain.ErrProviderMissing),
			errors.Is(err, domain.ErrSameProviderAndReceiver):
			httpx.WriteEmptyError(w, http.StatusBadRequest)
		case errors.Is(err, domain.ErrReviewAlreadyExists):
			httpx.WriteEmptyError(w, http.StatusConflict)
		default:
			log.Error("error creating deal item review", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, review.ToDTO())
}

// ================================================================================
// URL PARAM HELPERS
// ================================================================================

func parseReviewID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "reviewId")
}

func parseDealID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "dealId")
}

func parseItemID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "itemId")
}

func parseOfferID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "offerId")
}

func parseProviderID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "providerId")
}

func parseAuthorID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	return parseUUIDParam(w, r, "authorId")
}

func parseUUIDParam(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
