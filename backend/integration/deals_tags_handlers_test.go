package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"net/http"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListTagsReturnsSortedUniqueNormalizedTags(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	firstTags := []types.TagName{"specatag", "AlphaTag"}
	secondTags := []types.TagName{"alphatag", "OmegaTag"}

	_ = mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "First",
		Description: "First tagged offer",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &firstTags,
	})
	_ = mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Second",
		Description: "Second tagged offer",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &secondTags,
	})

	tags := mustListTags(t, userID)
	require.Contains(t, tags, types.TagName("alphatag"))
	require.Contains(t, tags, types.TagName("omegatag"))
	require.Contains(t, tags, types.TagName("specatag"))
	require.True(t, sort.SliceIsSorted(tags, func(i, j int) bool { return tags[i] < tags[j] }))
}

func TestDeleteAdminTagRemovesItFromOffers(t *testing.T) {
	dumpDealsLogs(t)

	adminToken := mustAdminAccessToken(t)
	userID := uuid.New()
	tags := []types.TagName{"deletetagalpha", "deletetagbeta"}

	first := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "First",
		Description: "First tagged offer",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &tags,
	})
	second := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Second",
		Description: "Second tagged offer",
		Type:        types.Good,
		Action:      types.Give,
		Tags:        &tags,
	})

	mustDeleteAdminTag(t, adminToken, "deletetagalpha")

	remainingTags := mustListTags(t, userID)
	require.NotContains(t, remainingTags, types.TagName("deletetagalpha"))
	require.Contains(t, remainingTags, types.TagName("deletetagbeta"))
	require.Equal(t, []types.TagName{"deletetagbeta"}, mustGetOfferByID(t, userID, first.Id).Tags)
	require.Equal(t, []types.TagName{"deletetagbeta"}, mustGetOfferByID(t, userID, second.Id).Tags)
}

func TestDeleteAdminTagForbiddenForNonAdmin(t *testing.T) {
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/admin/tags?name=bike", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
