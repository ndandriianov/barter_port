package integration

import (
	dealsusers "barter-port/contracts/kafka/messages/deals-users"
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

	registered := registerAuthUser(t, globalFixture)
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/review", registered.UserID, nil)
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
	dumpUsersLogs(t)

	fixture := globalFixture
	userA := registerAuthUser(t, fixture)
	userB := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, userA.UserID)
	waitForUsersProjection(t, fixture, userB.UserID)

	reportComment := "confirmed by admin"
	dealID, _ := mustCreateDiscussionDeal(t, userA.UserID, userB.UserID)
	mustVoteForFailure(t, userA.UserID, dealID, userB.UserID)

	points := 3
	adminToken := mustAdminAccessToken(t)

	resolveReq := mustBearerRequest(t, http.MethodPost, dealsURL()+"/deals/failures/"+dealID.String()+"/moderator-resolution", adminToken, mustJSONBody(t, types.ModeratorResolutionForFailureRequest{
		Confirmed:        true,
		UserId:           &userB.UserID,
		PunishmentPoints: &points,
		Comment:          &reportComment,
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
	require.Equal(t, userB.UserID, *resolution.UserId)
	require.NotNil(t, resolution.PunishmentPoints)
	require.Equal(t, points, *resolution.PunishmentPoints)

	getReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/failures/"+dealID.String()+"/moderator-resolution", userA.UserID, nil)
	getResp := mustDo(t, getReq)
	defer func() { _ = getResp.Body.Close() }()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetched types.DealFailureModeratorResolution
	require.NoError(t, json.NewDecoder(getResp.Body).Decode(&fetched))
	require.NotNil(t, fetched.Confirmed)
	require.True(t, *fetched.Confirmed)
	require.NotNil(t, fetched.UserId)
	require.Equal(t, userB.UserID, *fetched.UserId)

	deal := mustGetDealByID(t, userA.UserID, dealID)
	require.Equal(t, types.Failed, deal.Status)

	event := waitForUserReputationEvent(t, fixture, userB.UserID, dealsusers.DealFailureResponsibleMessageType, dealID)
	require.Equal(t, userB.UserID, event.UserID)
	require.Equal(t, -points, event.Delta)
	require.NotNil(t, event.Comment)
	require.Equal(t, reportComment, *event.Comment)

	me := mustGetCurrentUser(t, fixture, userB.UserID)
	require.Equal(t, -points, me.ReputationPoints)

	events := mustGetCurrentUserReputationEvents(t, fixture, userB.UserID)
	apiEvent := mustFindReputationEvent(t, events, dealsusers.DealFailureResponsibleMessageType, dealID)
	require.Equal(t, event.ID, uuid.UUID(apiEvent.Id))
	require.Equal(t, -points, apiEvent.Delta)
	require.NotNil(t, apiEvent.Comment)
	require.Equal(t, reportComment, *apiEvent.Comment)
}
