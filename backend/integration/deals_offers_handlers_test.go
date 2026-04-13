package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateOfferSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
	})

	require.Equal(t, userID, offer.AuthorId)
	require.Equal(t, "Vintage bike", offer.Name)
	require.Equal(t, "City bike in good condition", offer.Description)
	require.Equal(t, types.Good, offer.Type)
	require.Equal(t, types.Give, offer.Action)
	require.Nil(t, offer.PhotoUrls)
}

func TestCreateOfferMultipartWithPhotosSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG, tinyPNG})

	require.Equal(t, userID, offer.AuthorId)
	require.NotNil(t, offer.PhotoUrls)
	require.Len(t, *offer.PhotoUrls, 2)
	require.Contains(t, (*offer.PhotoUrls)[0], "/offer-photos/offer-"+offer.Id.String()+"/photo-0")
	require.Contains(t, (*offer.PhotoUrls)[1], "/offer-photos/offer-"+offer.Id.String()+"/photo-1")

	fetched := mustGetOfferByID(t, userID, offer.Id)
	require.Equal(t, offer.PhotoUrls, fetched.PhotoUrls)
}

func TestCreateOfferUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	req := mustRequest(t, http.MethodPost, dealsURL()+"/offers", mustJSONBody(t, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateOfferInvalidTypeReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers", userID, bytes.NewReader([]byte(`{
		"name":"Vintage bike",
		"description":"City bike in good condition",
		"type":"weird",
		"action":"give"
	}`)))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetOfferByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	offer := mustGetOfferByID(t, userID, offerID)

	require.Equal(t, offerID, offer.Id)
	require.Equal(t, userID, offer.AuthorId)
}

func TestGetOfferByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+uuid.NewString(), userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetOfferByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/not-a-uuid", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetOfferByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/offers/"+uuid.NewString(), nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetOffersUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/offers?sort=ByTime", nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetOffersInvalidSortReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers?sort=wrong", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetOffersMyDefaultFalseIncludesOtherUsersOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	offerB := mustCreateOffer(t, userB)

	result := mustGetOffers(t, userA, nil)
	require.NotEmpty(t, result.Offers)

	var foundA, foundB bool
	for _, offer := range result.Offers {
		if offer.Id == offerA {
			foundA = true
		}
		if offer.Id == offerB {
			foundB = true
		}
	}

	require.True(t, foundA)
	require.True(t, foundB)
}

func TestGetOffersMyTrueReturnsOnlyCurrentUserOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	offerB := mustCreateOffer(t, userB)

	result := mustGetOffers(t, userA, boolPtr(true))
	require.NotEmpty(t, result.Offers)

	var foundA, foundB bool
	for _, offer := range result.Offers {
		if offer.Id == offerA {
			foundA = true
		}
		if offer.Id == offerB {
			foundB = true
		}
		require.Equal(t, userA, offer.AuthorId)
	}

	require.True(t, foundA)
	require.False(t, foundB)
}

func TestGetOffersInvalidMyReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers?sort=ByTime&my=not-bool", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetOffersResponseIsJSON(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	_ = mustCreateOffer(t, userID)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers?sort=ByTime", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var decoded types.ListOffersResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
}

func TestUpdateOfferSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Old name",
		Description: "Old description",
		Type:        types.Good,
		Action:      types.Give,
	})

	newName := "New name"
	newDescription := "New description"
	newType := types.Service
	newAction := types.Take
	updated := mustUpdateOffer(t, userID, offer.Id, types.UpdateOfferRequest{
		Name:        &newName,
		Description: &newDescription,
		Type:        &newType,
		Action:      &newAction,
	})

	require.Equal(t, offer.Id, updated.Id)
	require.Equal(t, userID, updated.AuthorId)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, newDescription, updated.Description)
	require.Equal(t, types.Service, updated.Type)
	require.Equal(t, types.Take, updated.Action)
	require.NotNil(t, updated.UpdatedAt)

	fetched := mustGetOfferByID(t, userID, offer.Id)
	require.Equal(t, newName, fetched.Name)
	require.Equal(t, newDescription, fetched.Description)
	require.Equal(t, types.Service, fetched.Type)
	require.Equal(t, types.Take, fetched.Action)
	require.NotNil(t, fetched.UpdatedAt)
}

func TestUpdateOfferForbiddenForNonAuthor(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	otherUserID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	name := "Forbidden rename"

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), otherUserID, mustJSONBody(t, types.UpdateOfferRequest{
		Name: &name,
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateOfferRejectsEmptyPatch(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), userID, bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteOfferSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	mustDeleteOffer(t, userID, offerID)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+offerID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteOfferForbiddenForNonAuthor(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	otherUserID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/offers/"+offerID.String(), otherUserID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteOfferKeepsExistingDealItems(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	before := mustGetDealIDs(t, userID, true)
	beforeSet := make(map[uuid.UUID]struct{}, len(before))
	for _, id := range before {
		beforeSet[id] = struct{}{}
	}

	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	after := mustGetDealIDs(t, userID, true)
	var dealID uuid.UUID
	for _, id := range after {
		if _, ok := beforeSet[id]; !ok {
			dealID = id
			break
		}
	}
	require.NotEqual(t, uuid.Nil, dealID)

	dealBeforeDelete := mustGetDealByID(t, userID, dealID)
	var itemBeforeDelete *types.Item
	for i := range dealBeforeDelete.Items {
		if dealBeforeDelete.Items[i].AuthorId == userID {
			itemBeforeDelete = &dealBeforeDelete.Items[i]
			break
		}
	}
	require.NotNil(t, itemBeforeDelete)
	require.NotNil(t, itemBeforeDelete.OfferId)
	require.Equal(t, offerID, *itemBeforeDelete.OfferId)

	mustDeleteOffer(t, userID, offerID)

	dealAfterDelete := mustGetDealByID(t, userID, dealID)
	var itemAfterDelete *types.Item
	for i := range dealAfterDelete.Items {
		if dealAfterDelete.Items[i].Id == itemBeforeDelete.Id {
			itemAfterDelete = &dealAfterDelete.Items[i]
			break
		}
	}
	require.NotNil(t, itemAfterDelete)
	require.Nil(t, itemAfterDelete.OfferId)
}
