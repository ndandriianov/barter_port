package http

import (
	dealsdocfirst "barter-port/docs/doc-first/deals"
	"barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	joinsh "barter-port/internal/deals/infrastructure/transport/http/joins"
	"barter-port/internal/deals/infrastructure/transport/http/offers"
	"barter-port/pkg/authkit"
	"barter-port/pkg/authkit/validators"
	"barter-port/pkg/logger"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	logg *slog.Logger,
	validator *validators.LocalJWT,
	offersHandlers *offers.Handlers,
	draftsHandlers *draftsh.Handlers,
	dealsHandlers *deals.Handlers,
	joinsHandlers *joinsh.Handlers,
) http.Handler {
	if logg == nil {
		log.Fatal("logger is required")
	}
	if offersHandlers == nil {
		log.Fatal("offers handlers are required")
	}
	if draftsHandlers == nil {
		log.Fatal("drafts handlers are required")
	}
	if dealsHandlers == nil {
		log.Fatal("deals handlers are required")
	}
	if joinsHandlers == nil {
		log.Fatal("joins handlers are required")
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
			r.Post("/", offersHandlers.HandleCreateOffer)
			r.Get("/", offersHandlers.HandleGetOffers)
		})
		r.Route("/deals", func(r chi.Router) {
			r.Get("/", dealsHandlers.GetDeals)
			r.Get("/{dealId}", dealsHandlers.GetDealByID)
			r.Post("/{dealId}/items", dealsHandlers.AddDealItem)
			r.Get("/{dealId}/status", dealsHandlers.GetDealStatusVotes)
			r.Patch("/{dealId}/status", dealsHandlers.ChangeDealStatus)
			r.Patch("/{dealId}/items/{itemId}", dealsHandlers.UpdateDealItem)
			r.Post("/{dealId}/joins", joinsHandlers.JoinDeal)
			r.Get("/{dealId}/joins", joinsHandlers.GetDealJoinRequests)
			r.Delete("/{dealId}/joins", joinsHandlers.LeaveDeal)
			r.Post("/{dealId}/joins/{userId}", joinsHandlers.ProcessJoinRequest)
			r.Post("/drafts", draftsHandlers.CreateDraft)
			r.Get("/drafts", draftsHandlers.GetDrafts)
			r.Get("/drafts/{draftId}", draftsHandlers.GetDraftByID)
			r.Patch("/drafts/{draftId}", draftsHandlers.ConfirmDraft)
			r.Patch("/drafts/{draftId}/cancel", draftsHandlers.CancelDraft)
		})
	})

	return r
}
