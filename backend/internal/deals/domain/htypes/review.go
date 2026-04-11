package htypes

import (
	"barter-port/contracts/openapi/deals/types"

	"github.com/google/uuid"
)

type ReviewSummary struct {
	Count     int
	AvgRating float64
	Rating1   int
	Rating2   int
	Rating3   int
	Rating4   int
	Rating5   int
}

type ReviewEligibility struct {
	CanCreate   bool
	ContextType types.ReviewContextType
	ProviderID  *uuid.UUID
	OfferID     *uuid.UUID // nil for item-only
	ItemID      *uuid.UUID // nil for offer-only
	DealID      uuid.UUID
	Reason      *types.ReviewEligibilityReason
}

// PendingReview is an alias for ReviewEligibility — same shape, different semantic context.
type PendingReview = ReviewEligibility
