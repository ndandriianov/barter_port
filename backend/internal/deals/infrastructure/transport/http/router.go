package http

import (
	dealsdocfirst "barter-port/docs/doc-first/deals"
	"barter-port/pkg/authkit"
	"barter-port/pkg/authkit/validators"
	"barter-port/pkg/logger"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logg *slog.Logger, validator *validators.LocalJWT, h *OffersHandlers, dh *DealsHandlers) http.Handler {
	if logg == nil {
		log.Fatal("logger is required")
	}
	if h == nil {
		log.Fatal("handlers are required")
	}
	if dh == nil {
		log.Fatal("deals handlers are required")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	openAPISpecHandler := http.StripPrefix("/swagger/", http.FileServer(http.FS(dealsdocfirst.SpecFS)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logg.Error("failed to write health response", slog.String("error", err.Error()))
		}
	})
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/swagger.yaml", http.StatusPermanentRedirect)
	})
	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/swagger.yaml", http.StatusPermanentRedirect)
	})
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		openAPISpecHandler.ServeHTTP(w, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(authkit.Middleware(logg, validator, nil))
		r.Use(logger.Middleware(logg))
		r.Route("/offers", func(r chi.Router) {
			r.Post("/", h.HandleCreateOffer)
			r.Get("/", h.HandleGetOffers)
		})
		r.Route("/deals", func(r chi.Router) {
			r.Get("/", dh.GetDeals)
			r.Get("/{dealId}", dh.GetDealByID)
			r.Post("/drafts", dh.CreateDraft)
			r.Get("/drafts", dh.GetDrafts)
			r.Get("/drafts/{draftId}", dh.GetDraftByID)
			r.Patch("/drafts/{draftId}", dh.ConfirmDraft)
			r.Patch("/drafts/{draftId}/cancel", dh.CancelDraft)
		})
	})

	return r
}
