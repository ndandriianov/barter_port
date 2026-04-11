package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJoinDealSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	requests := mustGetJoinRequests(t, userA, dealID)
	require.Len(t, requests, 1)
	require.Equal(t, requesterID, requests[0].UserId)
	require.Equal(t, dealID, requests[0].DealId)
	require.Empty(t, requests[0].Voters)
}

func TestJoinDealUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodPost, dealsURL()+"/deals/"+uuid.NewString()+"/joins", nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDealJoinRequestsForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	strangerID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	joinReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	joinResp := mustDo(t, joinReq)
	defer func() { _ = joinResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, joinResp.StatusCode)

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID.String()+"/joins", strangerID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestProcessJoinRequestAcceptAllAddsParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	joinReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	joinResp := mustDo(t, joinReq)
	defer func() { _ = joinResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, joinResp.StatusCode)

	acceptA := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins/"+requesterID.String()+"?accept=true", userA, nil)
	respA := mustDo(t, acceptA)
	defer func() { _ = respA.Body.Close() }()
	require.Equal(t, http.StatusNoContent, respA.StatusCode)

	requestsAfterA := mustGetJoinRequests(t, userA, dealID)
	require.Len(t, requestsAfterA, 1)
	require.Equal(t, requesterID, requestsAfterA[0].UserId)
	require.Equal(t, []uuid.UUID{userA}, requestsAfterA[0].Voters)

	acceptB := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins/"+requesterID.String()+"?accept=true", userB, nil)
	respB := mustDo(t, acceptB)
	defer func() { _ = respB.Body.Close() }()
	require.Equal(t, http.StatusNoContent, respB.StatusCode)

	ids := mustGetDealIDs(t, requesterID, true)
	require.Contains(t, ids, dealID)

	requestsAfterB := mustGetJoinRequests(t, userA, dealID)
	require.Empty(t, requestsAfterB)
}

func TestProcessJoinRequestRejectRemovesRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	joinReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	joinResp := mustDo(t, joinReq)
	defer func() { _ = joinResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, joinResp.StatusCode)

	rejectReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins/"+requesterID.String()+"?accept=false", userA, nil)
	rejectResp := mustDo(t, rejectReq)
	defer func() { _ = rejectResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, rejectResp.StatusCode)

	requests := mustGetJoinRequests(t, userB, dealID)
	require.Empty(t, requests)
}

func TestLeaveDealRemovesJoinRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	joinReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	joinResp := mustDo(t, joinReq)
	defer func() { _ = joinResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, joinResp.StatusCode)

	leaveReq := mustUserRequest(t, http.MethodDelete, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	leaveResp := mustDo(t, leaveReq)
	defer func() { _ = leaveResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, leaveResp.StatusCode)

	requests := mustGetJoinRequests(t, userA, dealID)
	require.Empty(t, requests)
}

func TestJoinRequestsResponseDecodesAsExpected(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	requesterID := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	joinReq := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID.String()+"/joins", requesterID, nil)
	joinResp := mustDo(t, joinReq)
	defer func() { _ = joinResp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, joinResp.StatusCode)

	requests := mustGetJoinRequests(t, userB, dealID)
	require.IsType(t, types.GetDealJoinRequestsResponse{}, requests)
}
