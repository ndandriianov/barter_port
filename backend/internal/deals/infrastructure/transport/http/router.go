package http

import (
	dealsdocfirst "barter-port/docs/doc-first/deals"
	"barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	failuresh "barter-port/internal/deals/infrastructure/transport/http/failures"
	joinsh "barter-port/internal/deals/infrastructure/transport/http/joins"
	offergroupsh "barter-port/internal/deals/infrastructure/transport/http/offergroups"
	"barter-port/internal/deals/infrastructure/transport/http/offers"
	reviewsh "barter-port/internal/deals/infrastructure/transport/http/reviews"
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
	offerGroupsHandlers *offergroupsh.Handlers,
	draftsHandlers *draftsh.Handlers,
	dealsHandlers *deals.Handlers,
	failuresHandlers *failuresh.Handlers,
	joinsHandlers *joinsh.Handlers,
	reviewsHandlers *reviewsh.Handlers,
) http.Handler {
	if logg == nil {
		log.Fatal("logger is required")
	}
	if offersHandlers == nil {
		log.Fatal("offers handlers are required")
	}
	if offerGroupsHandlers == nil {
		log.Fatal("offer groups handlers are required")
	}
	if draftsHandlers == nil {
		log.Fatal("drafts handlers are required")
	}
	if dealsHandlers == nil {
		log.Fatal("deals handlers are required")
	}
	if failuresHandlers == nil {
		log.Fatal("failures handlers are required")
	}
	if joinsHandlers == nil {
		log.Fatal("joins handlers are required")
	}
	if reviewsHandlers == nil {
		log.Fatal("reviews handlers are required")
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
			r.Get("/{offerId}", offersHandlers.HandleGetOfferByID)
			r.Get("/{offerId}/reviews", reviewsHandlers.GetOfferReviews)
			r.Get("/{offerId}/reviews-summary", reviewsHandlers.GetOfferReviewsSummary)
		})
		r.Route("/offer-groups", func(r chi.Router) {
			r.Get("/", offerGroupsHandlers.ListOfferGroups)
			r.Post("/", offerGroupsHandlers.CreateOfferGroup)
			r.Get("/{offerGroupId}", offerGroupsHandlers.GetOfferGroupByID)
			r.Post("/{offerGroupId}/drafts", offerGroupsHandlers.CreateDraftFromOfferGroup)
		})
		r.Get("/providers/{providerId}/reviews", reviewsHandlers.GetProviderReviews)
		r.Get("/providers/{providerId}/reviews-summary", reviewsHandlers.GetProviderReviewsSummary)
		r.Get("/authors/{authorId}/reviews", reviewsHandlers.GetAuthorReviews)
		r.Get("/reviews/{reviewId}", reviewsHandlers.GetReviewByID)
		r.Patch("/reviews/{reviewId}", reviewsHandlers.UpdateReview)
		r.Delete("/reviews/{reviewId}", reviewsHandlers.DeleteReview)
		r.Route("/deals", func(r chi.Router) {
			r.Get("/", dealsHandlers.GetDeals)
			r.Get("/{dealId}", dealsHandlers.GetDealByID)
			r.Patch("/{dealId}", dealsHandlers.UpdateDeal)
			r.Get("/failures/review", failuresHandlers.GetDealsForFailureReview)
			r.Post("/failures/{dealId}/votes", failuresHandlers.VoteForFailure)
			r.Delete("/failures/{dealId}/votes", failuresHandlers.RevokeVoteForFailure)
			r.Get("/failures/{dealId}/votes", failuresHandlers.GetFailureVotes)
			r.Get("/failures/{dealId}/materials", failuresHandlers.GetFailureMaterials)
			r.Post("/failures/{dealId}/moderator-resolution", failuresHandlers.ModeratorResolutionForFailure)
			r.Get("/failures/{dealId}/moderator-resolution", failuresHandlers.GetModeratorResolutionForFailure)
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
			r.Delete("/drafts/{draftId}", draftsHandlers.DeleteDraft)
			r.Patch("/drafts/{draftId}", draftsHandlers.ConfirmDraft)
			r.Patch("/drafts/{draftId}/cancel", draftsHandlers.CancelDraft)
			r.Get("/{dealId}/reviews", reviewsHandlers.GetDealReviews)
			r.Get("/{dealId}/reviews-pending", reviewsHandlers.GetDealPendingReviews)
			r.Get("/{dealId}/items/{itemId}/reviews/eligibility", reviewsHandlers.GetDealItemReviewEligibility)
			r.Get("/{dealId}/items/{itemId}/reviews", reviewsHandlers.GetDealItemReviews)
			r.Post("/{dealId}/items/{itemId}/reviews", reviewsHandlers.CreateDealItemReview)
		})
	})

	return r
}
