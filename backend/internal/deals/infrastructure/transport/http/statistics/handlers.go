package statistics

import (
	"log/slog"
	"net/http"

	"barter-port/contracts/openapi/deals/types"
	statssvc "barter-port/internal/deals/application/statistics"
	statsrepo "barter-port/internal/deals/infrastructure/repository/statistics"
	"barter-port/pkg/authkit"
	"barter-port/pkg/httpx"
	"barter-port/pkg/logger"
)

type Handlers struct {
	log     *slog.Logger
	service *statssvc.Service
}

func NewHandlers(log *slog.Logger, service *statssvc.Service) *Handlers {
	return &Handlers{log: log, service: service}
}

func (h *Handlers) HandleGetMyStatistics(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "HandleGetMyStatistics"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	result, err := h.service.GetMyStatistics(r.Context(), userID)
	if err != nil {
		log.Error("failed to get statistics", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toResponse(result))
}

func toResponse(r *statsrepo.Result) types.MyStatistics {
	return types.MyStatistics{
		Deals: types.MyDealStats{
			Completed: r.DealsCompleted,
			Failed:    r.DealsFailed,
			Active:    r.DealsActive,
		},
		Offers: types.MyOfferStats{
			Total:      r.OffersTotal,
			TotalViews: r.OffersTotalViews,
		},
		Reviews: types.MyReviewStats{
			Written:               r.ReviewsWritten,
			Received:              r.ReviewsReceived,
			AverageRatingReceived: r.ReviewsAverageRatingReceived,
		},
		Reports: types.MyReportStats{
			OnMyOffers: types.MyReportOnMyOffersStats{
				Total:    r.ReportsOnMyOffersTotal,
				Pending:  r.ReportsOnMyOffersPending,
				Accepted: r.ReportsOnMyOffersAccepted,
				Rejected: r.ReportsOnMyOffersRejected,
			},
			FiledByMe: r.ReportsFiledByMe,
		},
	}
}
