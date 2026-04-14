package offer_reports

import (
	"barter-port/contracts/openapi/deals/types"
	offerreportsapp "barter-port/internal/deals/application/offer-reports"
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

type Handlers struct {
	service *offerreportsapp.Service
}

func NewHandlers(service *offerreportsapp.Service) *Handlers {
	return &Handlers{service: service}
}

// ================================================================================
// POST /offers/{offerId}/reports
// ================================================================================

func (h *Handlers) HandleCreateOfferReport(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "CreateOfferReport"))

	reporterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	var req types.CreateOfferReportJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	created, report, err := h.service.CreateReport(r.Context(), reporterID, offerID, req.MessageId)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrSelfReport):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrReporterAlreadyAttached):
			httpx.WriteEmptyError(w, http.StatusConflict)
		default:
			log.Error("failed to create offer report", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	if created {
		httpx.WriteJSON(w, http.StatusCreated, report.ToDto())
	} else {
		httpx.WriteJSON(w, http.StatusOK, report.ToDto())
	}
}

// ================================================================================
// GET /offers/{offerId}/reports
// ================================================================================

func (h *Handlers) HandleGetOfferReports(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "GetOfferReports"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	offerID, ok := parseOfferID(w, r)
	if !ok {
		return
	}

	offer, reports, messagesByReport, err := h.service.GetOfferReports(r.Context(), userID, offerID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOfferNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("failed to get offer reports", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	threads := make([]types.OfferReportThread, 0, len(reports))
	for _, rpt := range reports {
		msgs := messagesByReport[rpt.ID]
		dtoMsgs := make([]types.OfferReportMessage, 0, len(msgs))
		for _, m := range msgs {
			dtoMsgs = append(dtoMsgs, m.ToDto())
		}
		threads = append(threads, types.OfferReportThread{
			Report:   rpt.ToDto(),
			Messages: dtoMsgs,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, types.OfferReportsForOffer{
		Offer:   offer.ToDto(),
		Reports: threads,
	})
}

// ================================================================================
// GET /admin/offer-reports
// ================================================================================

func (h *Handlers) HandleListAdminReports(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "ListAdminReports"))

	adminID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	var status *domain.OfferReportStatus
	if s := r.URL.Query().Get("status"); s != "" {
		parsed := domain.OfferReportStatus(s)
		switch parsed {
		case domain.OfferReportStatusPending, domain.OfferReportStatusAccepted, domain.OfferReportStatusRejected:
			status = &parsed
		default:
			httpx.WriteEmptyError(w, http.StatusBadRequest)
			return
		}
	}

	reports, err := h.service.ListAdminReports(r.Context(), adminID, status)
	if err != nil {
		if errors.Is(err, domain.ErrAdminOnly) {
			httpx.WriteEmptyError(w, http.StatusForbidden)
			return
		}
		log.Error("failed to list admin reports", slog.Any("error", err))
		httpx.WriteEmptyError(w, http.StatusInternalServerError)
		return
	}

	resp := make(types.ListOfferReportsResponse, 0, len(reports))
	for _, rpt := range reports {
		resp = append(resp, rpt.ToDto())
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ================================================================================
// GET /admin/offer-reports/{reportId}
// ================================================================================

func (h *Handlers) HandleGetAdminReportDetails(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "GetAdminReportDetails"))

	adminID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	reportID, ok := parseReportID(w, r)
	if !ok {
		return
	}

	report, offer, messages, err := h.service.GetAdminReportDetails(r.Context(), adminID, reportID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAdminOnly):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrReportNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("failed to get admin report details", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	dtoMsgs := make([]types.OfferReportMessage, 0, len(messages))
	for _, m := range messages {
		dtoMsgs = append(dtoMsgs, m.ToDto())
	}

	var offerDTO types.Offer
	if offer != nil {
		offerDTO = offer.ToDto()
	}

	httpx.WriteJSON(w, http.StatusOK, types.OfferReportDetails{
		Report:   report.ToDto(),
		Offer:    offerDTO,
		Messages: dtoMsgs,
	})
}

// ================================================================================
// POST /admin/offer-reports/{reportId}/resolution
// ================================================================================

func (h *Handlers) HandleResolveReport(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), slog.Default()).With(slog.String("handler", "ResolveReport"))

	adminID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	reportID, ok := parseReportID(w, r)
	if !ok {
		return
	}

	var req types.ResolveOfferReportForAdminJSONRequestBody
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return
	}

	report, err := h.service.ResolveReport(r.Context(), adminID, reportID, req.Accepted, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAdminOnly):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrReportNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		case errors.Is(err, domain.ErrAlreadyReviewed):
			httpx.WriteEmptyError(w, http.StatusConflict)
		default:
			log.Error("failed to resolve report", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, report.ToDto())
}

// ================================================================================
// helpers
// ================================================================================

func parseOfferID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "offerId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func parseReportID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "reportId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.WriteEmptyError(w, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
