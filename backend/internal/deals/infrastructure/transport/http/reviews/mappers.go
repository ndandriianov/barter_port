package reviews

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"barter-port/internal/deals/domain/htypes"
)

func mapReviewsToDTO(reviews []domain.Review) []types.Review {
	result := make([]types.Review, 0, len(reviews))
	for i := range reviews {
		result = append(result, reviews[i].ToDTO())
	}
	return result
}

func mapReviewSummaryToDTO(s htypes.ReviewSummary) types.ReviewSummary {
	return types.ReviewSummary{
		Count:     s.Count,
		AvgRating: s.AvgRating,
		RatingBreakdown: types.ReviewRatingBreakdown{
			Rating1: s.Rating1,
			Rating2: s.Rating2,
			Rating3: s.Rating3,
			Rating4: s.Rating4,
			Rating5: s.Rating5,
		},
	}
}

func mapEligibilityToDTO(e htypes.ReviewEligibility) types.ReviewEligibility {
	dto := types.ReviewEligibility{
		CanCreate:   e.CanCreate,
		ContextType: e.ContextType,
		ProviderId:  e.ProviderID,
		Reason:      e.Reason,
	}
	if e.OfferID != nil {
		dto.OfferRef = &types.OfferRef{OfferId: *e.OfferID}
	}
	if e.ItemID != nil {
		dto.ItemRef = &types.DealItemRef{DealId: e.DealID, ItemId: *e.ItemID}
	}
	return dto
}

func mapPendingReviewsToDTO(pending []htypes.PendingReview) types.GetPendingDealReviewsResponse {
	result := make(types.GetPendingDealReviewsResponse, 0, len(pending))
	for _, p := range pending {
		dto := types.PendingDealReview{
			DealId:      p.DealID,
			CanCreate:   p.CanCreate,
			ContextType: p.ContextType,
			ProviderId:  p.ProviderID,
			Reason:      p.Reason,
		}
		if p.OfferID != nil {
			dto.OfferRef = &types.OfferRef{OfferId: *p.OfferID}
		}
		if p.ItemID != nil {
			dto.ItemRef = &types.DealItemRef{DealId: p.DealID, ItemId: *p.ItemID}
		}
		result = append(result, dto)
	}
	return result
}
