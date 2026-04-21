package integration

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
	"barter-port/contracts/openapi/deals/types"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetDealItemReviewEligibilityCanCreate(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA, _ := mustCreateCompletedReviewableTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String()+"/reviews/eligibility", userB, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var eligibility types.ReviewEligibility
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&eligibility))
	require.True(t, eligibility.CanCreate)
	require.Equal(t, types.OfferItem, eligibility.ContextType)
	require.NotNil(t, eligibility.ProviderId)
	require.Equal(t, userA, *eligibility.ProviderId)
	require.NotNil(t, eligibility.OfferRef)
	require.NotNil(t, eligibility.ItemRef)
	require.Nil(t, eligibility.Reason)
}

func TestCreateDealItemReviewSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)
	dumpUsersLogs(t)

	fixture := globalFixture
	userA := mustRegisterProjectedUser(t, fixture)
	userB := mustRegisterProjectedUser(t, fixture)
	dealID, itemIDByA, _ := mustCreateCompletedReviewableTwoPartyDeal(t, userA, userB)
	comment := "excellent"

	review := mustCreateDealItemReview(t, userB, dealID, itemIDByA, 5, &comment)
	require.Equal(t, userB, review.AuthorId)
	require.Equal(t, userA, review.ProviderId)
	require.Equal(t, dealID, review.DealId)
	require.Equal(t, 5, review.Rating)
	require.Equal(t, comment, *review.Comment)
	require.NotNil(t, review.OfferRef)
	require.NotNil(t, review.ItemRef)

	itemRefID := uuid.UUID(review.ItemRef.ItemId)
	offerRefID := uuid.UUID(review.OfferRef.OfferId)
	event := requireReviewCreationRewardEvent(t, fixture, userB, dealID, &itemRefID, &offerRefID, userA)
	waitForCurrentUserReputationAPIEvent(t, fixture, userB, dealsusers.ReviewCreationRewardMessageType, event.SourceID)
	waitForCurrentUserReputationPoints(t, fixture, userB, dealCompletionRewardPoints+reviewCreationRewardPoints)
}

func TestCreateDealItemReviewConflictAfterDuplicate(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA, _ := mustCreateCompletedReviewableTwoPartyDeal(t, userA, userB)
	_ = mustCreateDealItemReview(t, userB, dealID, itemIDByA, 5, new("excellent"))

	req := mustUserRequest(
		t,
		http.MethodPost,
		dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String()+"/reviews",
		userB,
		mustJSONBody(t, types.CreateReviewRequest{Rating: 4, Comment: new("still good")}),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateDealItemReviewForbiddenForNonReceiver(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA, _ := mustCreateCompletedReviewableTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(
		t,
		http.MethodPost,
		dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String()+"/reviews",
		userA,
		mustJSONBody(t, types.CreateReviewRequest{Rating: 5}),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetDealItemReviewEligibilityAlreadyReviewed(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+ctx.DealID.String()+"/items/"+ctx.ItemID.String()+"/reviews/eligibility", ctx.ReceiverID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var eligibility types.ReviewEligibility
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&eligibility))
	require.False(t, eligibility.CanCreate)
	require.NotNil(t, eligibility.Reason)
	require.Equal(t, types.AlreadyReviewed, *eligibility.Reason)
}

func TestGetDealPendingReviewsReflectsAvailability(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	reqReceiver := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+ctx.DealID.String()+"/reviews-pending", ctx.ReceiverID, nil)
	respReceiver := mustDo(t, reqReceiver)
	defer func() { _ = respReceiver.Body.Close() }()
	require.Equal(t, http.StatusOK, respReceiver.StatusCode)

	var receiverPending types.GetPendingDealReviewsResponse
	require.NoError(t, json.NewDecoder(respReceiver.Body).Decode(&receiverPending))
	require.Len(t, receiverPending, 1)
	require.False(t, receiverPending[0].CanCreate)
	require.NotNil(t, receiverPending[0].Reason)
	require.Equal(t, types.AlreadyReviewed, *receiverPending[0].Reason)

	reqProvider := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+ctx.DealID.String()+"/reviews-pending", ctx.ProviderID, nil)
	respProvider := mustDo(t, reqProvider)
	defer func() { _ = respProvider.Body.Close() }()
	require.Equal(t, http.StatusOK, respProvider.StatusCode)

	var providerPending types.GetPendingDealReviewsResponse
	require.NoError(t, json.NewDecoder(respProvider.Body).Decode(&providerPending))
	require.Len(t, providerPending, 1)
	require.True(t, providerPending[0].CanCreate)
}

func TestGetDealItemReviewsReturnsCreatedReview(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+ctx.DealID.String()+"/items/"+ctx.ItemID.String()+"/reviews", ctx.ProviderID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var reviews []types.Review
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reviews))
	require.Len(t, reviews, 1)
	require.Equal(t, ctx.Review.Id, reviews[0].Id)
}

func TestGetDealReviewsReturnsCreatedReview(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+ctx.DealID.String()+"/reviews", ctx.ProviderID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var reviews types.GetDealReviewsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reviews))
	require.Len(t, reviews, 1)
	require.Equal(t, ctx.Review.Id, reviews[0].Id)
}

func TestGetOfferReviewsAndSummary(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	reviewsReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+ctx.OfferID.String()+"/reviews", ctx.ProviderID, nil)
	reviewsResp := mustDo(t, reviewsReq)
	defer func() { _ = reviewsResp.Body.Close() }()
	require.Equal(t, http.StatusOK, reviewsResp.StatusCode)

	var reviews types.GetOfferReviewsResponse
	require.NoError(t, json.NewDecoder(reviewsResp.Body).Decode(&reviews))
	require.Len(t, reviews, 1)
	require.Equal(t, ctx.Review.Id, reviews[0].Id)

	summaryReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+ctx.OfferID.String()+"/reviews-summary", ctx.ProviderID, nil)
	summaryResp := mustDo(t, summaryReq)
	defer func() { _ = summaryResp.Body.Close() }()
	require.Equal(t, http.StatusOK, summaryResp.StatusCode)

	var summary types.ReviewSummary
	require.NoError(t, json.NewDecoder(summaryResp.Body).Decode(&summary))
	require.Equal(t, 1, summary.Count)
	require.Equal(t, 5.0, summary.AvgRating)
	require.Equal(t, 1, summary.RatingBreakdown.Rating5)
}

func TestGetProviderReviewsAndSummary(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	reviewsReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/providers/"+ctx.ProviderID.String()+"/reviews", ctx.ReceiverID, nil)
	reviewsResp := mustDo(t, reviewsReq)
	defer func() { _ = reviewsResp.Body.Close() }()
	require.Equal(t, http.StatusOK, reviewsResp.StatusCode)

	var reviews types.GetProviderReviewsResponse
	require.NoError(t, json.NewDecoder(reviewsResp.Body).Decode(&reviews))
	require.Len(t, reviews, 1)
	require.Equal(t, ctx.Review.Id, reviews[0].Id)

	summaryReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/providers/"+ctx.ProviderID.String()+"/reviews-summary", ctx.ReceiverID, nil)
	summaryResp := mustDo(t, summaryReq)
	defer func() { _ = summaryResp.Body.Close() }()
	require.Equal(t, http.StatusOK, summaryResp.StatusCode)

	var summary types.ReviewSummary
	require.NoError(t, json.NewDecoder(summaryResp.Body).Decode(&summary))
	require.Equal(t, 1, summary.Count)
	require.Equal(t, 1, summary.RatingBreakdown.Rating5)
}

func TestGetAuthorReviewsReturnsCreatedReview(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/authors/"+ctx.ReceiverID.String()+"/reviews", ctx.ProviderID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var reviews types.GetAuthorReviewsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reviews))
	require.Len(t, reviews, 1)
	require.Equal(t, ctx.Review.Id, reviews[0].Id)
}

func TestGetReviewByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/reviews/"+ctx.Review.Id.String(), ctx.ProviderID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var review types.Review
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&review))
	require.Equal(t, ctx.Review.Id, review.Id)
}

func TestUpdateReviewSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ctx := mustCreateReviewedOfferItemContext(t)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/reviews/"+ctx.Review.Id.String(), ctx.ReceiverID, mustJSONBody(t, types.UpdateReviewRequest{
		Rating:  new(4),
		Comment: new("updated"),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var review types.Review
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&review))
	require.Equal(t, 4, review.Rating)
	require.NotNil(t, review.Comment)
	require.Equal(t, "updated", *review.Comment)
	require.NotNil(t, review.UpdatedAt)
}

func TestDeleteReviewSuccessAllowsRecreate(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)
	dumpUsersLogs(t)

	fixture := globalFixture
	ctx := mustCreateReviewedOfferItemContextWithRegisteredUsers(t)
	sourceID := dealsusers.BuildReviewCreationRewardSourceID(ctx.DealID, &ctx.ItemID, &ctx.OfferID, ctx.ReceiverID, ctx.ProviderID)
	waitForCurrentUserReputationPoints(t, fixture, ctx.ReceiverID, dealCompletionRewardPoints+reviewCreationRewardPoints)

	deleteReq := mustUserRequest(t, http.MethodDelete, dealsURL()+"/reviews/"+ctx.Review.Id.String(), ctx.ReceiverID, nil)
	deleteResp := mustDo(t, deleteReq)
	defer func() { _ = deleteResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	recreated := mustCreateDealItemReview(t, ctx.ReceiverID, ctx.DealID, ctx.ItemID, 5, new("recreated"))
	require.NotEqual(t, ctx.Review.Id, recreated.Id)

	time.Sleep(2 * time.Second)
	require.Equal(t, 1, countUserReputationEvents(t, fixture, ctx.ReceiverID, dealsusers.ReviewCreationRewardMessageType, sourceID))
	me := mustGetCurrentUser(t, fixture, ctx.ReceiverID)
	require.Equal(t, dealCompletionRewardPoints+reviewCreationRewardPoints, int(me.ReputationPoints))
}
