package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOffersIncludeDistanceWhenViewerAndOfferHaveLocations(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	mustUpdateCurrentUserLocation(t, viewer.UserID, 55.751244, 37.618423)

	lat := 55.761244
	lon := 37.628423
	offer := mustCreateOfferDetails(t, author.UserID, types.CreateOfferRequest{
		Name:        "Offer with geo",
		Description: "Near viewer",
		Type:        types.Good,
		Action:      types.Give,
		Latitude:    &lat,
		Longitude:   &lon,
	})

	list := mustGetOffers(t, viewer.UserID, nil)

	var listed *types.Offer
	for i := range list.Offers {
		if list.Offers[i].Id == offer.Id {
			listed = &list.Offers[i]
			break
		}
	}

	require.NotNil(t, listed)
	require.NotNil(t, listed.DistanceMeters)
	require.Greater(t, *listed.DistanceMeters, int64(0))

	fetched := mustGetOfferByID(t, viewer.UserID, offer.Id)
	require.NotNil(t, fetched.DistanceMeters)
	require.Equal(t, *listed.DistanceMeters, *fetched.DistanceMeters)
}

func TestOwnOfferIncludesDistanceWhenUserHasLocation(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	user := mustRegisterDealsUser(t)
	mustUpdateCurrentUserLocation(t, user.UserID, 55.751244, 37.618423)

	lat := 55.761244
	lon := 37.628423
	offer := mustCreateOfferDetails(t, user.UserID, types.CreateOfferRequest{
		Name:        "Own offer with geo",
		Description: "Distance should be returned for own offer too",
		Type:        types.Good,
		Action:      types.Give,
		Latitude:    &lat,
		Longitude:   &lon,
	})

	fetched := mustGetOfferByID(t, user.UserID, offer.Id)
	require.NotNil(t, fetched.DistanceMeters)
	require.Greater(t, *fetched.DistanceMeters, int64(0))
}

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
	require.Nil(t, offer.PhotoIds)
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
	require.NotNil(t, offer.PhotoIds)
	require.NotNil(t, offer.PhotoUrls)
	require.Len(t, *offer.PhotoIds, 2)
	require.Len(t, *offer.PhotoUrls, 2)
	require.NotEqual(t, uuid.Nil, (*offer.PhotoIds)[0])
	require.NotEqual(t, uuid.Nil, (*offer.PhotoIds)[1])
	require.Contains(t, (*offer.PhotoUrls)[0], "/offer-photos/offer-"+offer.Id.String()+"/photo-0")
	require.Contains(t, (*offer.PhotoUrls)[1], "/offer-photos/offer-"+offer.Id.String()+"/photo-1")

	fetched := mustGetOfferByID(t, userID, offer.Id)
	require.Equal(t, offer.PhotoIds, fetched.PhotoIds)
	require.Equal(t, offer.PhotoUrls, fetched.PhotoUrls)
}

func TestCreateOfferWithTagsSuccess(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	tags := []types.TagName{" VeloTag ", "RepairTag", "velotag"}
	offer := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Tagged bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &tags,
	})

	require.Equal(t, []types.TagName{"repairtag", "velotag"}, offer.Tags)

	fetched := mustGetOfferByID(t, userID, offer.Id)
	require.Equal(t, []types.TagName{"repairtag", "velotag"}, fetched.Tags)
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

func TestViewOfferByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	viewerID := uuid.New()
	offerID := mustCreateOffer(t, authorID)

	before := mustGetOfferByID(t, viewerID, offerID)
	require.EqualValues(t, 0, before.Views)

	mustViewOfferByID(t, viewerID, offerID)

	after := mustGetOfferByID(t, viewerID, offerID)
	require.EqualValues(t, 1, after.Views)
}

func TestViewOfferByIDAffectsPopularitySorting(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	viewerID := uuid.New()

	popularOffer := mustCreateOffer(t, authorID)
	lessPopularOffer := mustCreateOffer(t, authorID)

	mustViewOfferByID(t, viewerID, popularOffer)
	mustViewOfferByID(t, viewerID, popularOffer)
	mustViewOfferByID(t, viewerID, lessPopularOffer)

	result := mustGetOffersBySort(t, authorID, "ByPopularity", new(true))
	require.Len(t, result.Offers, 2)
	require.Equal(t, popularOffer, result.Offers[0].Id)
	require.EqualValues(t, 2, result.Offers[0].Views)
	require.Equal(t, lessPopularOffer, result.Offers[1].Id)
	require.EqualValues(t, 1, result.Offers[1].Views)
}

func TestViewOfferByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers/"+uuid.NewString()+"/view", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestViewOfferByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers/not-a-uuid/view", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestViewOfferByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodPost, dealsURL()+"/offers/"+uuid.NewString()+"/view", nil))
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

	result := mustGetOffers(t, userA, new(true))
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

func TestGetOffersFiltersByTagsQuery(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	fullTags := []types.TagName{"filteralpha", "filterbeta"}
	bikeOnly := []types.TagName{"filteralpha"}

	target := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Repair bike",
		Description: "Tagged target offer",
		Type:        types.Service,
		Action:      types.Give,
		Tags:        &fullTags,
	})
	_ = mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Just bike",
		Description: "Partial match",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &bikeOnly,
	})
	_ = mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "No tags",
		Description: "Should not match",
		Type:        types.Good,
		Action:      types.Give,
	})

	result := mustGetOffersBySortAndTags(t, userID, "ByTime", new(true), &fullTags, false)
	require.Len(t, result.Offers, 1)
	require.Equal(t, target.Id, result.Offers[0].Id)
	require.Equal(t, userID, result.Offers[0].AuthorId)
	require.Equal(t, []types.TagName{"filteralpha", "filterbeta"}, result.Offers[0].Tags)
}

func TestGetOffersFiltersWithoutTags(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	tags := []types.TagName{"emptyfiltertag"}
	_ = mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Tagged",
		Description: "Has tags",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &tags,
	})
	untagged := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Untagged",
		Description: "No tags",
		Type:        types.Good,
		Action:      types.Give,
	})

	result := mustGetOffersBySortAndTags(t, userID, "ByTime", new(true), nil, true)
	require.Len(t, result.Offers, 1)
	require.Equal(t, untagged.Id, result.Offers[0].Id)
	require.Equal(t, userID, result.Offers[0].AuthorId)
	require.Empty(t, result.Offers[0].Tags)
}

func TestGetOffersByTimeSupportsCursorPaginationWithinSameSecond(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	myOffersOnly := true
	newerOfferID := mustCreateOffer(t, userID)
	olderOfferID := mustCreateOffer(t, userID)

	baseTime := time.Date(2026, time.April, 22, 12, 0, 0, 0, time.UTC)
	mustSetOfferCreatedAt(t, newerOfferID, baseTime.Add(900*time.Millisecond))
	mustSetOfferCreatedAt(t, olderOfferID, baseTime.Add(500*time.Millisecond))

	firstPage := mustGetOffersPage(t, userID, "ByTime", nil, 1, &myOffersOnly)
	require.Len(t, firstPage.Offers, 1)
	require.Equal(t, newerOfferID, firstPage.Offers[0].Id)
	require.NotNil(t, firstPage.NextCursor)

	secondPage := mustGetOffersPage(t, userID, "ByTime", firstPage.NextCursor, 1, &myOffersOnly)
	require.Len(t, secondPage.Offers, 1)
	require.Equal(t, olderOfferID, secondPage.Offers[0].Id)
}

func TestGetSubscribedOffersUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/offers/subscriptions?sort=ByTime", nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubscribedOffersInvalidSortReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	user := mustRegisterDealsUser(t)
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/subscriptions?sort=wrong", user.UserID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetSubscribedOffersReturnsOnlySubscribedAuthorsOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	subscriber := mustRegisterDealsUser(t)
	subscribedAuthor := mustRegisterDealsUser(t)
	otherAuthor := mustRegisterDealsUser(t)

	subscribedOffer := mustCreateOffer(t, subscribedAuthor.UserID)
	otherOffer := mustCreateOffer(t, otherAuthor.UserID)

	mustSubscribeToUser(t, subscriber.UserID, subscribedAuthor.UserID)

	result := mustGetSubscribedOffers(t, subscriber.UserID)
	require.NotEmpty(t, result.Offers)

	var foundSubscribed, foundOther bool
	for _, offer := range result.Offers {
		if offer.Id == subscribedOffer {
			foundSubscribed = true
		}
		if offer.Id == otherOffer {
			foundOther = true
		}
		require.Equal(t, subscribedAuthor.UserID, offer.AuthorId)
	}

	require.True(t, foundSubscribed)
	require.False(t, foundOther)
}

func TestGetSubscribedOffersByPopularityReturnsMostViewedFirst(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	subscriber := mustRegisterDealsUser(t)
	author := mustRegisterDealsUser(t)

	popularOffer := mustCreateOffer(t, author.UserID)
	lessPopularOffer := mustCreateOffer(t, author.UserID)

	mustSubscribeToUser(t, subscriber.UserID, author.UserID)

	mustViewOfferByID(t, subscriber.UserID, popularOffer)
	mustViewOfferByID(t, subscriber.UserID, popularOffer)
	mustViewOfferByID(t, subscriber.UserID, lessPopularOffer)

	result := mustGetSubscribedOffersBySort(t, subscriber.UserID, "ByPopularity")
	require.Len(t, result.Offers, 2)
	require.Equal(t, popularOffer, result.Offers[0].Id)
	require.EqualValues(t, 2, result.Offers[0].Views)
	require.Equal(t, lessPopularOffer, result.Offers[1].Id)
	require.EqualValues(t, 1, result.Offers[1].Views)
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
	updated := mustUpdateOffer(t, userID, offer.Id, types.UpdateOfferRequest{
		Name:        &newName,
		Description: &newDescription,
		Type:        new(types.Service),
		Action:      new(types.Take),
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

func TestUpdateOfferPhotosKeepsRemainingOrderAndAppendsNewPhotos(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG, tinyPNG})

	require.NotNil(t, offer.PhotoIds)
	require.NotNil(t, offer.PhotoUrls)
	require.Len(t, *offer.PhotoIds, 2)
	require.Len(t, *offer.PhotoUrls, 2)

	deletePhotoIDs := []uuid.UUID{(*offer.PhotoIds)[0]}
	updated := mustUpdateOfferMultipartDetails(t, userID, offer.Id, types.UpdateOfferRequest{
		DeletePhotoIds: &deletePhotoIDs,
	}, [][]byte{tinyPNG})

	require.NotNil(t, updated.PhotoIds)
	require.NotNil(t, updated.PhotoUrls)
	require.Len(t, *updated.PhotoIds, 2)
	require.Len(t, *updated.PhotoUrls, 2)
	require.Equal(t, (*offer.PhotoIds)[1], (*updated.PhotoIds)[0])
	require.Equal(t, (*offer.PhotoUrls)[1], (*updated.PhotoUrls)[0])
	require.NotEqual(t, (*offer.PhotoIds)[0], (*updated.PhotoIds)[1])
	require.NotEqual(t, (*offer.PhotoIds)[1], (*updated.PhotoIds)[1])
	require.Contains(t, (*updated.PhotoUrls)[1], "/offer-photos/offer-"+offer.Id.String()+"/photo-2")

	fetched := mustGetOfferByID(t, userID, offer.Id)
	require.Equal(t, updated.PhotoIds, fetched.PhotoIds)
	require.Equal(t, updated.PhotoUrls, fetched.PhotoUrls)
}

func TestUpdateOfferReplacesTags(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	initialTags := []types.TagName{"updatetagalpha", "updatetagbeta"}
	offer := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &initialTags,
	})

	newTags := []types.TagName{"updatetaggamma"}
	updated := mustUpdateOffer(t, userID, offer.Id, types.UpdateOfferRequest{
		Tags: &newTags,
	})
	require.Equal(t, []types.TagName{"updatetaggamma"}, updated.Tags)

	emptyTags := []types.TagName{}
	cleared := mustUpdateOffer(t, userID, offer.Id, types.UpdateOfferRequest{
		Tags: &emptyTags,
	})
	require.Empty(t, cleared.Tags)
}

func TestUpdateOfferRejectsUnknownPhotoID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offer := mustCreateOfferMultipartDetails(t, userID, types.CreateOfferRequest{
		Name:        "Vintage bike",
		Description: "City bike in good condition",
		Type:        types.Good,
		Action:      types.Give,
	}, [][]byte{tinyPNG})

	deletePhotoIDs := []uuid.UUID{uuid.New()}
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	require.NoError(t, writer.WriteField("deletePhotoIds", deletePhotoIDs[0].String()))
	require.NoError(t, writer.Close())

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offer.Id.String(), userID, &payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateOfferForbiddenForNonAuthor(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	authorID := uuid.New()
	otherUserID := uuid.New()
	offerID := mustCreateOffer(t, authorID)
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), otherUserID, mustJSONBody(t, types.UpdateOfferRequest{
		Name: new("Forbidden rename"),
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

func mustRegisterDealsUser(t *testing.T) registerResponse {
	t.Helper()

	user := registerAuthUser(t, globalFixture)
	waitForUsersProjection(t, globalFixture, user.UserID)

	return user
}
