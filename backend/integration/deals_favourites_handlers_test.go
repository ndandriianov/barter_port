package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAddOfferToFavoritesSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	offer := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, offer)

	result := mustGetFavoriteOffers(t, viewer.UserID)
	require.Len(t, result.Offers, 1)
	require.Equal(t, offer, result.Offers[0].Id)
	require.Equal(t, author.UserID, result.Offers[0].AuthorId)
	require.NotNil(t, result.Offers[0].IsFavorite)
	require.True(t, *result.Offers[0].IsFavorite)
	require.False(t, result.Offers[0].FavoritedAt.IsZero())
}

func TestAddOfferToFavoritesIsIdempotent(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	offer := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, offer)
	mustAddOfferToFavorites(t, viewer.UserID, offer)

	result := mustGetFavoriteOffers(t, viewer.UserID)
	require.Len(t, result.Offers, 1)
	require.Equal(t, offer, result.Offers[0].Id)
}

func TestRemoveOfferFromFavoritesSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	favoriteOfferID := mustCreateOffer(t, author.UserID)
	otherOfferID := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, favoriteOfferID)
	mustRemoveOfferFromFavorites(t, viewer.UserID, favoriteOfferID)

	favorites := mustGetFavoriteOffers(t, viewer.UserID)
	require.Empty(t, favorites.Offers)

	publicList := mustGetOffers(t, viewer.UserID, nil)
	requireFavoriteFlag(t, publicList.Offers, favoriteOfferID, false)
	requireFavoriteFlag(t, publicList.Offers, otherOfferID, false)

	fetched := mustGetOfferByID(t, viewer.UserID, favoriteOfferID)
	require.NotNil(t, fetched.IsFavorite)
	require.False(t, *fetched.IsFavorite)
}

func TestGetFavoriteOffersSupportsCursorPagination(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	firstOfferID := mustCreateOffer(t, author.UserID)
	secondOfferID := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, firstOfferID)
	time.Sleep(20 * time.Millisecond)
	mustAddOfferToFavorites(t, viewer.UserID, secondOfferID)

	firstPage := mustGetFavoriteOffersPage(t, viewer.UserID, nil, 1)
	require.Len(t, firstPage.Offers, 1)
	require.Equal(t, secondOfferID, firstPage.Offers[0].Id)
	require.NotNil(t, firstPage.NextCursor)

	secondPage := mustGetFavoriteOffersPage(t, viewer.UserID, firstPage.NextCursor, 1)
	require.Len(t, secondPage.Offers, 1)
	require.Equal(t, firstOfferID, secondPage.Offers[0].Id)

	thirdPage := mustGetFavoriteOffersPage(t, viewer.UserID, secondPage.NextCursor, 1)
	require.Empty(t, thirdPage.Offers)
}

func TestFavoriteFlagsAppearInOffersListAndGetOfferByID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	favoriteOfferID := mustCreateOffer(t, author.UserID)
	otherOfferID := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, favoriteOfferID)

	publicList := mustGetOffers(t, viewer.UserID, nil)
	requireFavoriteFlag(t, publicList.Offers, favoriteOfferID, true)
	requireFavoriteFlag(t, publicList.Offers, otherOfferID, false)

	favoriteOffer := mustGetOfferByID(t, viewer.UserID, favoriteOfferID)
	require.NotNil(t, favoriteOffer.IsFavorite)
	require.True(t, *favoriteOffer.IsFavorite)

	otherOffer := mustGetOfferByID(t, viewer.UserID, otherOfferID)
	require.NotNil(t, otherOffer.IsFavorite)
	require.False(t, *otherOffer.IsFavorite)
}

func TestGetFavoriteOffersExcludesHiddenOfferForNonAuthor(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	author := mustRegisterDealsUser(t)
	viewer := mustRegisterDealsUser(t)
	offerID := mustCreateOffer(t, author.UserID)

	mustAddOfferToFavorites(t, viewer.UserID, offerID)
	mustSetOfferHidden(t, offerID, true)

	favorites := mustGetFavoriteOffers(t, viewer.UserID)
	require.Empty(t, favorites.Offers)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+offerID.String(), viewer.UserID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddOfferToFavoritesUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	req := mustRequest(t, http.MethodPut, dealsURL()+"/offers/"+uuid.NewString()+"/favorite", nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAddOfferToFavoritesInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	user := mustRegisterDealsUser(t)
	req := mustUserRequest(t, http.MethodPut, dealsURL()+"/offers/not-a-uuid/favorite", user.UserID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetFavoriteOffersUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	req := mustRequest(t, http.MethodGet, dealsURL()+"/offers/favorites", nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func requireFavoriteFlag(t *testing.T, offers []types.Offer, offerID uuid.UUID, want bool) {
	t.Helper()

	for _, offer := range offers {
		if offer.Id != offerID {
			continue
		}

		require.NotNil(t, offer.IsFavorite)
		require.Equal(t, want, *offer.IsFavorite)
		return
	}

	t.Fatalf("offer %s not found in response", offerID)
}
