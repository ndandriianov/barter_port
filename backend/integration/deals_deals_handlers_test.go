package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetDealsUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/deals", nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDealsMyEmpty(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	ids := mustGetDealIDs(t, userID, true)
	require.Empty(t, ids)
}

func TestGetDealsMyTrueReturnsParticipantDeals(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	ids := mustGetDealIDs(t, userID, true)
	require.Contains(t, ids, dealID)
}

func TestGetDealsMyTrueExcludesOtherUsersDeals(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	mustCreateDeal(t, userA)

	ids := mustGetDealIDs(t, userB, true)
	require.Empty(t, ids)
}

func TestGetDealsOpenTrueExcludesClosedDeals(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	openDealID := mustCreateDeal(t, userID)
	closedDealID := mustCreateDeal(t, userID)
	_ = mustChangeDealStatus(t, closedDealID, userID, types.Cancelled)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals?open=true", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deals types.GetDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deals))

	var ids []uuid.UUID
	for _, deal := range deals {
		ids = append(ids, deal.Id)
	}

	require.Contains(t, ids, openDealID)
	require.NotContains(t, ids, closedDealID)
}

func TestGetDealsReturnsItemsWithCopiedPhotos(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Camera",
		Description: "Film camera",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG, tinyPNG})
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offer.Id, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	deals := mustGetDealsResponse(t, userID, true)
	require.Len(t, deals, 1)
	require.Len(t, deals[0].Items, 1)

	item := deals[0].Items[0]
	require.NotNil(t, item.OfferId)
	require.Equal(t, offer.Id, *item.OfferId)
	require.NotNil(t, item.PhotoIds)
	require.NotNil(t, item.PhotoUrls)
	require.Len(t, *item.PhotoIds, 2)
	require.Len(t, *item.PhotoUrls, 2)
	require.Contains(t, (*item.PhotoUrls)[0], "/offer-photos/item-"+item.Id.String()+"/photo-0")
	require.Contains(t, (*item.PhotoUrls)[1], "/offer-photos/item-"+item.Id.String()+"/photo-1")
	require.NotEqual(t, (*offer.PhotoIds)[0], (*item.PhotoIds)[0])
	require.NotEqual(t, (*offer.PhotoIds)[1], (*item.PhotoIds)[1])
}

func TestGetDealByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	dealIDs := mustGetDealIDs(t, userID, true)
	require.Len(t, dealIDs, 1)

	deal := mustGetDealByID(t, userID, dealIDs[0])
	require.Equal(t, dealIDs[0], deal.Id)
	require.False(t, deal.CreatedAt.IsZero())
	require.Len(t, deal.Items, 1)
	require.Equal(t, userID, deal.Items[0].AuthorId)
}

func TestGetDealByIDCopiesOfferPhotosIntoItemAndPersistsMetadata(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Turntable",
		Description: "Works well",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG, tinyPNG})
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offer.Id, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	dealID := mustGetDealIDs(t, userID, true)[0]
	deal := mustGetDealByID(t, userID, dealID)
	require.Len(t, deal.Items, 1)

	item := deal.Items[0]
	require.NotNil(t, item.OfferId)
	require.Equal(t, offer.Id, *item.OfferId)
	require.NotNil(t, item.PhotoIds)
	require.NotNil(t, item.PhotoUrls)
	require.Len(t, *item.PhotoIds, 2)
	require.Len(t, *item.PhotoUrls, 2)
	require.Contains(t, (*item.PhotoUrls)[0], "/offer-photos/item-"+item.Id.String()+"/photo-0")
	require.Contains(t, (*item.PhotoUrls)[1], "/offer-photos/item-"+item.Id.String()+"/photo-1")
	require.NotEqual(t, (*offer.PhotoIds)[0], (*item.PhotoIds)[0])
	require.NotEqual(t, (*offer.PhotoIds)[1], (*item.PhotoIds)[1])

	pool := OpenDatabase(t, globalFixture, "deals_db")
	var photoIDs []uuid.UUID
	var photoURLs []string
	err := pool.QueryRow(
		context.Background(),
		`SELECT array_agg(id ORDER BY position), array_agg(url ORDER BY position) FROM items_photos WHERE item_id = $1`,
		item.Id,
	).Scan(&photoIDs, &photoURLs)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID(*item.PhotoIds), photoIDs)
	require.Equal(t, []string(*item.PhotoUrls), photoURLs)
}

func TestGetDealByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+uuid.NewString(), userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDealByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/not-a-uuid", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDealByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/deals/"+uuid.NewString(), nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDealSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String(), userA, mustJSONBody(t, types.UpdateDealRequest{
		Name: "Renamed deal",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal types.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	require.NotNil(t, deal.Name)
	require.Equal(t, "Renamed deal", *deal.Name)
}

func TestUpdateDealForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	strangerID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String(), strangerID, mustJSONBody(t, types.UpdateDealRequest{
		Name: "Renamed deal",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateDealNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+uuid.NewString(), userID, mustJSONBody(t, types.UpdateDealRequest{
		Name: "Renamed deal",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateDealUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	req := mustRequest(t, http.MethodPatch, dealsURL()+"/deals/"+uuid.NewString(), mustJSONBody(t, types.UpdateDealRequest{
		Name: "Renamed deal",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAddDealItemUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := doAddDealItem(t, uuid.NewString(), nil, []byte(`{"offerId":"`+uuid.NewString()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAddDealItemInvalidDealID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	resp := doAddDealItem(t, "not-a-uuid", &userID, []byte(`{"offerId":"`+uuid.NewString()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddDealItemSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)

	resp := doAddDealItem(t, dealID.String(), &userID, []byte(`{"offerId":"`+offerID.String()+`","quantity":2}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal types.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	var found bool
	for _, item := range deal.Items {
		if item.AuthorId == userID && item.Quantity == 2 {
			found = true
		}
	}
	require.True(t, found)
}

func TestAddDealItemCopiesOfferPhotosIntoNewItem(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Projector",
		Description: "With HDMI",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG})

	resp := doAddDealItem(t, dealID.String(), &userID, []byte(`{"offerId":"`+offer.Id.String()+`","quantity":2}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal types.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	var addedItem *types.Item
	for i := range deal.Items {
		item := &deal.Items[i]
		if item.OfferId != nil && *item.OfferId == offer.Id {
			addedItem = item
			break
		}
	}
	require.NotNil(t, addedItem)
	require.NotNil(t, addedItem.PhotoIds)
	require.NotNil(t, addedItem.PhotoUrls)
	require.Len(t, *addedItem.PhotoIds, 1)
	require.Len(t, *addedItem.PhotoUrls, 1)
	require.Contains(t, (*addedItem.PhotoUrls)[0], "/offer-photos/item-"+addedItem.Id.String()+"/photo-0")
	require.NotEqual(t, (*offer.PhotoIds)[0], (*addedItem.PhotoIds)[0])

	pool := OpenDatabase(t, globalFixture, "deals_db")
	var photoCount int
	err := pool.QueryRow(
		context.Background(),
		`SELECT count(*) FROM items_photos WHERE item_id = $1`,
		addedItem.Id,
	).Scan(&photoCount)
	require.NoError(t, err)
	require.Equal(t, 1, photoCount)
}

func TestAddDealItemNotParticipantForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	strangerID := uuid.New()
	dealID := mustCreateDeal(t, ownerID)
	offerID := mustCreateOffer(t, strangerID)

	resp := doAddDealItem(t, dealID.String(), &strangerID, []byte(`{"offerId":"`+offerID.String()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAddDealItemOfferNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	resp := doAddDealItem(t, dealID.String(), &userID, []byte(`{"offerId":"`+uuid.NewString()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddDealItemForeignOfferForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)
	offerB := mustCreateOffer(t, userB)

	resp := doAddDealItem(t, dealID.String(), &userA, []byte(`{"offerId":"`+offerB.String()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAddDealItemInvalidQuantity(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)

	resp := doAddDealItem(t, dealID.String(), &userID, []byte(`{"offerId":"`+offerID.String()+`","quantity":0}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddDealItemClosedDealForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)
	_ = mustChangeDealStatus(t, dealID, userID, types.Cancelled)

	resp := doAddDealItem(t, dealID.String(), &userID, []byte(`{"offerId":"`+offerID.String()+`","quantity":1}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateDealItemUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	body, err := json.Marshal(types.UpdateDealItemRequest{Name: stringPtr("x")})
	require.NoError(t, err)

	req := mustRequest(t, http.MethodPatch, dealsURL()+"/deals/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDealItemEmptyPatchReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	deal := mustGetDealByID(t, userID, dealID)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+deal.Items[0].Id.String(), userID, mustJSONBody(t, types.UpdateDealItemRequest{}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateDealItemAuthorCanEditContent(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	deal := mustGetDealByID(t, userID, dealID)

	newName := "updated item"
	newDescription := "updated description"
	newQty := 7
	item := mustUpdateDealItem(t, userID, dealID, deal.Items[0].Id, types.UpdateDealItemRequest{
		Name:        &newName,
		Description: &newDescription,
		Quantity:    &newQty,
	})

	require.Equal(t, deal.Items[0].Id, item.Id)
	require.Equal(t, newName, item.Name)
	require.Equal(t, newDescription, item.Description)
	require.EqualValues(t, newQty, item.Quantity)
}

func TestUpdateDealItemNonAuthorCannotEditContent(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String(), userB, mustJSONBody(t, types.UpdateDealItemRequest{
		Name: stringPtr("forbidden update"),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateDealItemParticipantCanClaimAndReleaseReceiver(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)

	claimed := mustUpdateDealItem(t, userB, dealID, itemIDByA, types.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
	})
	require.NotNil(t, claimed.ReceiverId)
	require.Equal(t, userB, *claimed.ReceiverId)

	released := mustUpdateDealItem(t, userB, dealID, itemIDByA, types.UpdateDealItemRequest{
		ReleaseReceiver: boolPtr(true),
	})
	require.Nil(t, released.ReceiverId)
}

func TestChangeDealStatusUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := doChangeDealStatus(t, uuid.New(), nil, []byte(`{"expectedStatus":"Discussion"}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestChangeDealStatusInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/not-a-uuid/status", userID, bytes.NewReader([]byte(`{"expectedStatus":"Discussion"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	resp := doChangeDealStatus(t, uuid.New(), &userID, []byte(`{"expectedStatus":"Discussion"}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestChangeDealStatusInvalidJSONReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	resp := doChangeDealStatus(t, dealID, &userID, []byte(`not-json`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusUnknownStatusReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	resp := doChangeDealStatus(t, dealID, &userID, []byte(`{"expectedStatus":"unknown"}`))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusInvalidTransitionReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	body, err := json.Marshal(types.ChangeDealStatusRequest{ExpectedStatus: types.Confirmed})
	require.NoError(t, err)

	resp := doChangeDealStatus(t, dealID, &userID, body)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusConsensusMovesToDiscussion(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)
	otherItemID := mustGetDealItemIDByAuthor(t, userA, dealID, userB)

	_ = mustUpdateDealItem(t, userB, dealID, itemIDByA, types.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
	})
	_ = mustUpdateDealItem(t, userA, dealID, otherItemID, types.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
	})

	firstVote := mustChangeDealStatus(t, dealID, userA, types.Discussion)
	require.Equal(t, types.LookingForParticipants, firstVote.Status)

	secondVote := mustChangeDealStatus(t, dealID, userB, types.Discussion)
	require.Equal(t, types.Discussion, secondVote.Status)
}

func TestChangeDealStatusCancelledAppliesImmediately(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	updated := mustChangeDealStatus(t, dealID, userA, types.Cancelled)
	require.Equal(t, types.Cancelled, updated.Status)

	dealAfter := mustGetDealByID(t, userB, dealID)
	require.Equal(t, types.Cancelled, dealAfter.Status)
}

func TestGetDealStatusVotesUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := doGetDealStatusVotes(t, uuid.NewString(), nil)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDealStatusVotesInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	resp := doGetDealStatusVotes(t, "not-a-uuid", &userID)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDealStatusVotesNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	resp := doGetDealStatusVotes(t, uuid.NewString(), &userID)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDealStatusVotesEmptyWhenNoVotes(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	resp := doGetDealStatusVotes(t, dealID.String(), &userA)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var votes types.GetDealStatusVotesResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&votes))
	require.Empty(t, votes)
}

func TestGetDealStatusVotesReturnsRecordedVotes(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)

	_ = mustUpdateDealItem(t, userB, dealID, itemIDByA, types.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
	})
	otherItemID := mustGetDealItemIDByAuthor(t, userA, dealID, userB)
	_ = mustUpdateDealItem(t, userA, dealID, otherItemID, types.UpdateDealItemRequest{
		ClaimReceiver: boolPtr(true),
	})

	_ = mustChangeDealStatus(t, dealID, userA, types.Discussion)

	resp := doGetDealStatusVotes(t, dealID.String(), &userB)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var votes types.GetDealStatusVotesResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&votes))
	require.Len(t, votes, 1)
	require.Equal(t, userA, votes[0].UserId)
	require.Equal(t, types.Discussion, votes[0].Vote)
}
