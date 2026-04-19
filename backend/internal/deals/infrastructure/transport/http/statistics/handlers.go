package statistics

import (
	"log/slog"
	"net/http"

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

// ================================================================================
// Response types
// ================================================================================

type response struct {
	Deals   dealStats   `json:"deals"`
	Offers  offerStats  `json:"offers"`
	Reviews reviewStats `json:"reviews"`
	Reports reportStats `json:"reports"`
}

type dealStats struct {
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Active    int `json:"active"`
}

type offerStats struct {
	Total      int   `json:"total"`
	TotalViews int64 `json:"totalViews"`
}

type reviewStats struct {
	Written               int      `json:"written"`
	Received              int      `json:"received"`
	AverageRatingReceived *float64 `json:"averageRatingReceived"`
}

type reportOnMyOffersStats struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

type reportStats struct {
	OnMyOffers reportOnMyOffersStats `json:"onMyOffers"`
	FiledByMe  int                   `json:"filedByMe"`
}

func toResponse(r *statsrepo.Result) response {
	return response{
		Deals: dealStats{
			Completed: r.DealsCompleted,
			Failed:    r.DealsFailed,
			Active:    r.DealsActive,
		},
		Offers: offerStats{
			Total:      r.OffersTotal,
			TotalViews: r.OffersTotalViews,
		},
		Reviews: reviewStats{
			Written:               r.ReviewsWritten,
			Received:              r.ReviewsReceived,
			AverageRatingReceived: r.ReviewsAverageRatingReceived,
		},
		Reports: reportStats{
			OnMyOffers: reportOnMyOffersStats{
				Total:    r.ReportsOnMyOffersTotal,
				Pending:  r.ReportsOnMyOffersPending,
				Accepted: r.ReportsOnMyOffersAccepted,
				Rejected: r.ReportsOnMyOffersRejected,
			},
			FiledByMe: r.ReportsFiledByMe,
		},
	}
}
