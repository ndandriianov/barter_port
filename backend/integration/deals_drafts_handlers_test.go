package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateDraftNoOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, mustJSONBody(t, map[string]any{
		"offers": []any{},
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.Equal(t, domain.ErrNoOffers.Error(), *apiErr.Message)
}

func TestCreateDraftNoItemsWithNameAndDescription(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, mustJSONBody(t, map[string]any{
		"offers":      []any{},
		"name":        "My Draft Deal",
		"description": "This is a test draft",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.Equal(t, domain.ErrNoOffers.Error(), *apiErr.Message)
}

func TestCreateDraftWithOffersOnly(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, mustJSONBody(t, map[string]any{
		"offers": []map[string]any{
			{"offerID": offerID.String(), "quantity": 2},
		},
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created types.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)
}

func TestCreateDraftWithOffersAndNameAndDescription(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, mustJSONBody(t, map[string]any{
		"offers": []map[string]any{
			{"offerID": offerID.String(), "quantity": 2},
		},
		"name":        "My Draft Deal",
		"description": "This is a test draft",
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created types.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)
}

func TestCreateDraftUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	req := mustRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", mustJSONBody(t, map[string]any{"offers": []any{}}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDraftInvalidJSON(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDraftsEmpty(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	ids := mustGetDraftIDs(t, userID, nil)
	require.Empty(t, ids)
}

func TestGetDraftsUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/deals/drafts", nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDraftsCreatedByMeTrueReturnsOnlyAuthoredDrafts(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)

	draftByA := mustCreateDraft(t, userA, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})
	draftByBWithOfferA := mustCreateDraft(t, userB, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})

	ids := mustGetDraftIDs(t, userA, new(true))
	require.Contains(t, ids, draftByA)
	require.NotContains(t, ids, draftByBWithOfferA)
}

func TestGetDraftsCreatedByMeFalseReturnsParticipatingDrafts(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	offerB := mustCreateOffer(t, userB)

	draftByBWithOfferA := mustCreateDraft(t, userB, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})
	draftByBWithOfferB := mustCreateDraft(t, userB, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerB, Quantity: 1}})

	ids := mustGetDraftIDs(t, userA, new(false))
	require.Contains(t, ids, draftByBWithOfferA)
	require.NotContains(t, ids, draftByBWithOfferB)
}

func TestGetDraftsCreatedByMeDefaultFalse(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	_ = mustCreateDraft(t, userB, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})

	idsDefault := mustGetDraftIDs(t, userA, nil)
	idsFalse := mustGetDraftIDs(t, userA, new(false))
	require.ElementsMatch(t, idsFalse, idsDefault)
}

func TestGetDraftsInvalidCreatedByMeReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts?createdByMe=not-bool", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDraftsInvalidParticipatingReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts?participating=not-bool", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDraftByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	name := "Test Draft"
	description := "Test Description"
	offer1ID := mustCreateOffer(t, userID)
	offer2ID := mustCreateOffer(t, userID)

	draftID := mustCreateDraft(t, userID, &name, &description, []types.OfferIDAndQuantity{
		{OfferID: offer1ID, Quantity: 2},
		{OfferID: offer2ID, Quantity: 5},
	})

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var draft types.Draft
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&draft))
	require.Equal(t, draftID, draft.Id)
	require.Equal(t, userID, draft.AuthorId)
	require.NotNil(t, draft.Name)
	require.Equal(t, name, *draft.Name)
	require.NotNil(t, draft.Description)
	require.Equal(t, description, *draft.Description)
	require.Len(t, draft.Offers, 2)
	require.False(t, draft.CreatedAt.IsZero())
}

func TestGetDraftByIDWithOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{
		{OfferID: offerID, Quantity: 3},
	})

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var draft types.Draft
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&draft))
	require.Equal(t, draftID, draft.Id)
	require.Len(t, draft.Offers, 1)
	require.Equal(t, offerID, draft.Offers[0].Id)
}

func TestGetDraftByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+uuid.NewString(), userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDraftByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/not-a-uuid", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDraftByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+uuid.NewString(), nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestConfirmDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var confirmResp types.ConfirmDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&confirmResp))
	require.NotEmpty(t, confirmResp.Users)

	var found bool
	for _, u := range confirmResp.Users {
		if u.UserId == userID {
			found = true
			require.True(t, u.Confirmed)
		}
	}
	require.True(t, found)
}

func TestConfirmDraftUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := mustDo(t, mustRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString(), nil))
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestConfirmDraftInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/not-a-uuid", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestConfirmDraftForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	nonParticipantID := uuid.New()
	offerID := mustCreateOffer(t, ownerID)
	draftID := mustCreateDraft(t, ownerID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nonParticipantID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestConfirmDraftNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString(), userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCancelDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCancelDraftForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	nonParticipantID := uuid.New()
	offerID := mustCreateOffer(t, ownerID)
	draftID := mustCreateDraft(t, ownerID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", nonParticipantID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCancelDraftNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString()+"/cancel", userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCancelDraftNotFoundAfterAllConfirmed(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	offerB := mustCreateOffer(t, userB)

	draftID := mustCreateDraft(t, userA, nil, nil, []types.OfferIDAndQuantity{
		{OfferID: offerA, Quantity: 1},
		{OfferID: offerB, Quantity: 1},
	})
	mustConfirmDraft(t, userA, draftID)
	mustConfirmDraft(t, userB, draftID)

	cancelReq := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", userA, nil)
	cancelResp := mustDo(t, cancelReq)
	defer func() { _ = cancelResp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, cancelResp.StatusCode)
}

func TestDeleteDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	deleteReq := mustUserRequest(t, http.MethodDelete, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	deleteResp := mustDo(t, deleteReq)
	defer func() { _ = deleteResp.Body.Close() }()
	require.Equal(t, http.StatusOK, deleteResp.StatusCode)

	getReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	getResp := mustDo(t, getReq)
	defer func() { _ = getResp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestDeleteDraftForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	nonParticipantID := uuid.New()
	offerID := mustCreateOffer(t, ownerID)
	draftID := mustCreateDraft(t, ownerID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/deals/drafts/"+draftID.String(), nonParticipantID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteDraftNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/deals/drafts/"+uuid.NewString(), userID, nil)

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
