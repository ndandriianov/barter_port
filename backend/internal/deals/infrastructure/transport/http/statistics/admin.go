package statistics

import (
	"barter-port/contracts/openapi/deals/types"
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

func (h *Handlers) HandleGetAdminPlatformStatistics(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "HandleGetAdminPlatformStatistics"))

	userID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	result, err := h.service.GetAdminPlatformStatistics(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		default:
			log.Error("failed to get admin platform statistics", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	topTags := make([]types.AdminTopTagStat, 0, len(result.Offers.TopTags))
	for _, item := range result.Offers.TopTags {
		topTags = append(topTags, types.AdminTopTagStat{
			Tag:         types.TagName(item.Tag),
			OffersCount: item.OffersCount,
		})
	}

	topFavorites := make([]types.AdminTopFavoriteOfferStat, 0, len(result.Offers.TopByFavorites))
	for _, item := range result.Offers.TopByFavorites {
		topFavorites = append(topFavorites, types.AdminTopFavoriteOfferStat{
			OfferId:        item.OfferID,
			AuthorId:       item.AuthorID,
			Name:           item.Name,
			FavoritesCount: item.FavoritesCount,
		})
	}

	topReportedUsers := make([]types.AdminTopReportedUserStat, 0, len(result.Reports.TopUsersByReceivedReports))
	for _, item := range result.Reports.TopUsersByReceivedReports {
		topReportedUsers = append(topReportedUsers, types.AdminTopReportedUserStat{
			UserId:       item.UserID,
			ReportsCount: item.ReportsCount,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, types.AdminPlatformStatistics{
		Offers: types.AdminPlatformOfferStatistics{
			Total: result.Offers.Total,
			Hidden: types.AdminHiddenOfferStatistics{
				Moderated:      result.Offers.Hidden.Moderated,
				HiddenByAuthor: result.Offers.Hidden.HiddenByAuthor,
			},
			ByType: types.AdminOfferTypeDistribution{
				Good:    result.Offers.ByType.Good,
				Service: result.Offers.ByType.Service,
			},
			ByAction: types.AdminOfferActionDistribution{
				Give: result.Offers.ByAction.Give,
				Take: result.Offers.ByAction.Take,
			},
			TopTags:        topTags,
			AveragePerUser: result.Offers.AveragePerUser,
			Drafts:         result.Offers.Drafts,
			TotalViews:     result.Offers.TotalViews,
			TopByFavorites: topFavorites,
			AverageRating:  result.Offers.AverageRating,
		},
		Deals: types.AdminPlatformDealStatistics{
			Total: result.Deals.Total,
			ByStatus: types.AdminDealStatusDistribution{
				LookingForParticipants: result.Deals.ByStatus.LookingForParticipants,
				Discussion:             result.Deals.ByStatus.Discussion,
				Confirmed:              result.Deals.ByStatus.Confirmed,
				Completed:              result.Deals.ByStatus.Completed,
				Failed:                 result.Deals.ByStatus.Failed,
				Cancelled:              result.Deals.ByStatus.Cancelled,
			},
			SuccessfulConversionRate: result.Deals.SuccessfulConversionRate,
			AverageParticipants:      result.Deals.AverageParticipants,
			MultiPartyShare:          result.Deals.MultiPartyShare,
		},
		Reports: types.AdminPlatformReportStatistics{
			Total:                     result.Reports.Total,
			Pending:                   result.Reports.Pending,
			BlockedOffers:             result.Reports.BlockedOffers,
			AdminFailureResolutions:   result.Reports.AdminFailureResolutions,
			TopUsersByReceivedReports: topReportedUsers,
		},
		Reviews: types.AdminPlatformReviewStatistics{
			Total:         result.Reviews.Total,
			AverageRating: result.Reviews.AverageRating,
			RatingDistribution: types.AdminRatingDistribution{
				OneStar:    result.Reviews.RatingDistribution.OneStar,
				TwoStars:   result.Reviews.RatingDistribution.TwoStars,
				ThreeStars: result.Reviews.RatingDistribution.ThreeStars,
				FourStars:  result.Reviews.RatingDistribution.FourStars,
				FiveStars:  result.Reviews.RatingDistribution.FiveStars,
			},
		},
	})
}

func (h *Handlers) HandleGetAdminUserStatistics(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFrom(r.Context(), h.log).With(slog.String("handler", "HandleGetAdminUserStatistics"))

	requesterID, ok := authkit.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteEmptyError(w, http.StatusUnauthorized)
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteErrorStr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	result, err := h.service.GetAdminUserStatistics(r.Context(), requesterID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			httpx.WriteEmptyError(w, http.StatusForbidden)
		case errors.Is(err, domain.ErrUserNotFound):
			httpx.WriteEmptyError(w, http.StatusNotFound)
		default:
			log.Error("failed to get admin user statistics", slog.Any("error", err))
			httpx.WriteEmptyError(w, http.StatusInternalServerError)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, types.AdminUserStatistics{
		Deals: types.AdminUserDealStatistics{
			Completed: result.Deals.Completed,
			Active:    result.Deals.Active,
			Failed: types.AdminUserFailedDealStatistics{
				Total:       result.Deals.Failed.Total,
				Responsible: result.Deals.Failed.Responsible,
				Affected:    result.Deals.Failed.Affected,
			},
			Cancelled: result.Deals.Cancelled,
		},
		Offers: types.AdminUserOfferStatistics{
			Published:  result.Offers.Published,
			TotalViews: result.Offers.TotalViews,
		},
		Reviews: types.AdminUserReviewStatistics{
			Received:              result.Reviews.Received,
			AverageReceivedRating: result.Reviews.AverageReceivedRating,
			Written:               result.Reviews.Written,
		},
		Reports: types.AdminUserReportStatistics{
			Received: types.AdminUserReceivedReportStatistics{
				Accepted: result.Reports.Received.Accepted,
				Rejected: result.Reports.Received.Rejected,
			},
			Filed: result.Reports.Filed,
		},
	})
}
