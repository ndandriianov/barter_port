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
