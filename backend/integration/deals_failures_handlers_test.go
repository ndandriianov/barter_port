package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetDealsForFailureReviewNonAdminForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/review", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetDealsForFailureReviewAdminSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateDiscussionDeal(t, userA, userB)
	mustVoteForFailure(t, userA, dealID, userB)

	req := mustBearerRequest(t, http.MethodGet, dealsURL()+"/deals/failures/review", mustAdminAccessToken(t), nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deals types.FailureModerationDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deals))

	var ids []uuid.UUID
	for _, deal := range deals {
		ids = append(ids, deal.Id)
	}
	require.Contains(t, ids, dealID)
}

func TestVoteForFailureAndGetVotes(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateDiscussionDeal(t, userA, userB)
	mustVoteForFailure(t, userA, dealID, userB)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/"+dealID.String()+"/votes", userB, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var votes types.FailureVotesResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&votes))
	require.Len(t, votes, 1)
	require.Equal(t, userA, votes[0].UserId)
	require.Equal(t, userB, votes[0].Vote)
}

func TestRevokeVoteForFailureSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()
	userD := uuid.New()
	dealID, _ := mustCreateDiscussionDeal(t, userA, userB, userC, userD)

	mustVoteForFailure(t, userA, dealID, userB)

	revokeReq := mustUserRequest(t, http.MethodDelete, dealsURL()+"/deals/failures/"+dealID.String()+"/votes", userA, nil)
	revokeResp := mustDo(t, revokeReq)
	defer func() { _ = revokeResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

	votesReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/"+dealID.String()+"/votes", userB, nil)
	votesResp := mustDo(t, votesReq)
	defer func() { _ = votesResp.Body.Close() }()
	require.Equal(t, http.StatusOK, votesResp.StatusCode)

	var votes types.FailureVotesResponse
	require.NoError(t, json.NewDecoder(votesResp.Body).Decode(&votes))
	require.Empty(t, votes)
}

func TestGetFailureMaterialsAdminSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateDiscussionDeal(t, userA, userB)
	mustVoteForFailure(t, userA, dealID, userB)

	req := mustBearerRequest(t, http.MethodGet, dealsURL()+"/deals/failures/"+dealID.String()+"/materials", mustAdminAccessToken(t), nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var materials types.FailureMaterialResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&materials))
	require.Equal(t, dealID, materials.Deal.Id)
}

func TestModeratorResolutionForFailureSuccessAndReadableByParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateDiscussionDeal(t, userA, userB)
	mustVoteForFailure(t, userA, dealID, userB)

	comment := "confirmed by admin"
	points := 3
	adminToken := mustAdminAccessToken(t)

	resolveReq := mustBearerRequest(t, http.MethodPost, dealsURL()+"/deals/failures/"+dealID.String()+"/moderator-resolution", adminToken, mustJSONBody(t, types.ModeratorResolutionForFailureRequest{
		Confirmed:        true,
		UserId:           &userB,
		PunishmentPoints: &points,
		Comment:          &comment,
	}))
	resolveReq.Header.Set("Content-Type", "application/json")

	resolveResp := mustDo(t, resolveReq)
	defer func() { _ = resolveResp.Body.Close() }()
	require.Equal(t, http.StatusOK, resolveResp.StatusCode)

	var resolution types.DealFailureModeratorResolution
	require.NoError(t, json.NewDecoder(resolveResp.Body).Decode(&resolution))
	require.NotNil(t, resolution.Confirmed)
	require.True(t, *resolution.Confirmed)
	require.NotNil(t, resolution.UserId)
	require.Equal(t, userB, *resolution.UserId)
	require.NotNil(t, resolution.PunishmentPoints)
	require.Equal(t, points, *resolution.PunishmentPoints)

	getReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/"+dealID.String()+"/moderator-resolution", userA, nil)
	getResp := mustDo(t, getReq)
	defer func() { _ = getResp.Body.Close() }()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetched types.DealFailureModeratorResolution
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&fetched))
	require.NotNil(t, fetched.Confirmed)
	require.True(t, *fetched.Confirmed)
	require.NotNil(t, fetched.UserId)
	require.Equal(t, userB, *fetched.UserId)

	deal := mustGetDealByID(t, userA, dealID)
	require.Equal(t, types.Failed, deal.Status)
}
