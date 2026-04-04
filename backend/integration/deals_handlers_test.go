package integration

import (
	dealstypes "barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────
// Вспомогательные функции
// ────────────────────────────────────────────────────────────────

func dumpDealsLogs(t *testing.T) {
	t.Helper()
	DumpLogsOnFailure(t, globalFixture.Items, "deals")
}

func dealsURL() string {
	return globalFixture.DealsURL
}

func mustCreateOffer(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()

	body, err := json.Marshal(dealstypes.CreateOfferRequest{
		Name:        fmt.Sprintf("offer-%d", time.Now().UnixNano()),
		Description: "test offer",
		Type:        dealstypes.Good,
		Action:      dealstypes.Give,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/offers/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var offer dealstypes.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))
	require.NotEqual(t, uuid.Nil, offer.Id)

	return offer.Id
}

func mustCreateDraft(
	t *testing.T,
	userID uuid.UUID,
	name *string,
	description *string,
	offers []dealstypes.OfferIDAndQuantity,
) uuid.UUID {
	t.Helper()

	reqBody := map[string]any{
		"offers":      offers,
		"name":        name,
		"description": description,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created dealstypes.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)

	return created.Id
}

func mustGetOffers(t *testing.T, userID uuid.UUID, my *bool) dealstypes.ListOffersResponse {
	t.Helper()

	url := dealsURL() + "/offers/?sort=ByTime&cursor_limit=100"
	if my != nil {
		url += fmt.Sprintf("&my=%t", *my)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result dealstypes.ListOffersResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func mustGetDraftIDs(t *testing.T, userID uuid.UUID, createdByMe *bool) []uuid.UUID {
	t.Helper()

	url := dealsURL() + "/deals/drafts"
	if createdByMe != nil {
		url += fmt.Sprintf("?createdByMe=%t", *createdByMe)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var drafts dealstypes.GetMyDraftDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&drafts))

	ids := make([]uuid.UUID, 0, len(drafts))
	for _, draft := range drafts {
		ids = append(ids, draft.Id)
	}

	return ids
}

// ────────────────────────────────────────────────────────────────
// CreateDraft
// ────────────────────────────────────────────────────────────────

func TestCreateDraftNoOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	body, err := json.Marshal(map[string]any{
		"offers": []any{},
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr dealstypes.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.Equal(t, domain.ErrNoOffers.Error(), *apiErr.Message)
}

func TestCreateDraftNoItemsWithNameAndDescription(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	name := "My Draft Deal"
	description := "This is a test draft"

	body, err := json.Marshal(map[string]any{
		"offers":      []any{},
		"name":        name,
		"description": description,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr dealstypes.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.Equal(t, domain.ErrNoOffers.Error(), *apiErr.Message)
}

func TestCreateDraftWithOffersOnly(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	body, err := json.Marshal(map[string]any{
		"offers": []map[string]any{
			{"offerID": offerID.String(), "quantity": 2},
		},
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created dealstypes.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)
}

func TestCreateDraftWithOffersAndNameAndDescription(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	name := "My Draft Deal"
	description := "This is a test draft"

	body, err := json.Marshal(map[string]any{
		"offers": []map[string]any{
			{"offerID": offerID.String(), "quantity": 2},
		},
		"name":        name,
		"description": description,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created dealstypes.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)
}

func TestCreateDraftUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	body, err := json.Marshal(map[string]any{"offers": []any{}})
	require.NoError(t, err)

	resp, err := http.Post(dealsURL()+"/deals/drafts", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDraftInvalidJSON(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/deals/drafts", bytes.NewReader([]byte(`not-json`)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// GetDrafts
// ────────────────────────────────────────────────────────────────

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

	resp, err := http.Get(dealsURL() + "/deals/drafts")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDraftsCreatedByMeTrueReturnsOnlyAuthoredDrafts(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)

	draftByA := mustCreateDraft(t, userA, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})
	draftByBWithOfferA := mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})

	createdByMe := true
	ids := mustGetDraftIDs(t, userA, &createdByMe)

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

	draftByBWithOfferA := mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})
	draftByBWithOfferB := mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerB, Quantity: 1}})

	createdByMe := false
	ids := mustGetDraftIDs(t, userA, &createdByMe)

	require.Contains(t, ids, draftByBWithOfferA)
	require.NotContains(t, ids, draftByBWithOfferB)
}

func TestGetDraftsCreatedByMeDefaultFalse(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	_ = mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})

	idsDefault := mustGetDraftIDs(t, userA, nil)
	createdByMe := false
	idsFalse := mustGetDraftIDs(t, userA, &createdByMe)

	require.ElementsMatch(t, idsFalse, idsDefault)
}

func TestGetDraftsInvalidCreatedByMeReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts?createdByMe=not-bool", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// GetDraftByID
// ────────────────────────────────────────────────────────────────

func TestGetDraftByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	name := "Test Draft"
	description := "Test Description"
	offer1ID := mustCreateOffer(t, userID)
	offer2ID := mustCreateOffer(t, userID)

	offers := []dealstypes.OfferIDAndQuantity{
		{OfferID: offer1ID, Quantity: 2},
		{OfferID: offer2ID, Quantity: 5},
	}

	draftID := mustCreateDraft(t, userID, &name, &description, offers)

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var draft dealstypes.Draft
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&draft))
	require.Equal(t, draftID, draft.Id)
	require.Equal(t, userID, draft.AuthorId)
	require.NotNil(t, draft.Name)
	require.Equal(t, name, *draft.Name)
	require.NotNil(t, draft.Description)
	require.Equal(t, description, *draft.Description)

	require.Len(t, draft.Offers, 2)

	var foundOffer1, foundOffer2 bool
	for _, it := range draft.Offers {
		switch it.Id {
		case offer1ID:
			foundOffer1 = true
			require.EqualValues(t, 2, it.Quantity)
		case offer2ID:
			foundOffer2 = true
			require.EqualValues(t, 5, it.Quantity)
		}
	}

	require.True(t, foundOffer1)
	require.True(t, foundOffer2)

	require.False(t, draft.CreatedAt.IsZero())
}

func TestGetDraftByIDWithOffers(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)

	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{
		{OfferID: offerID, Quantity: 3},
	})

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var draft dealstypes.Draft
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&draft))
	require.Equal(t, draftID, draft.Id)
	require.Len(t, draft.Offers, 1)
	require.Equal(t, offerID, draft.Offers[0].Id)
}

func TestGetDraftByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts/"+uuid.NewString(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDraftByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts/not-a-uuid", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDraftByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.Get(dealsURL() + "/deals/drafts/" + uuid.NewString())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// ConfirmDraft
// ────────────────────────────────────────────────────────────────

func TestConfirmDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var confirmResp dealstypes.ConfirmDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&confirmResp))
	require.NotEmpty(t, confirmResp.Users)
	require.Equal(t, userID, confirmResp.Users[0].UserId)
	require.True(t, confirmResp.Users[0].Confirmed)
}

func TestConfirmDraftUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.DefaultClient.Do(mustRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString(), nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestConfirmDraftInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/not-a-uuid", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// CancelDraft
// ────────────────────────────────────────────────────────────────

func TestCancelDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCancelDraftForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	nonParticipantID := uuid.New()
	offerID := mustCreateOffer(t, ownerID)
	draftID := mustCreateDraft(t, ownerID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, nonParticipantID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCancelDraftForbiddenAfterAllConfirmed(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	offerA := mustCreateOffer(t, userA)
	offerB := mustCreateOffer(t, userB)

	draftID := mustCreateDraft(t, userA, nil, nil, []dealstypes.OfferIDAndQuantity{
		{OfferID: offerA, Quantity: 1},
		{OfferID: offerB, Quantity: 1},
	})

	confirmA, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	confirmA.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userA))

	respA, err := http.DefaultClient.Do(confirmA)
	require.NoError(t, err)
	defer func() { _ = respA.Body.Close() }()
	require.Equal(t, http.StatusOK, respA.StatusCode)

	confirmB, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	confirmB.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userB))

	respB, err := http.DefaultClient.Do(confirmB)
	require.NoError(t, err)
	defer func() { _ = respB.Body.Close() }()
	require.Equal(t, http.StatusOK, respB.StatusCode)

	cancelReq, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String()+"/cancel", nil)
	require.NoError(t, err)
	cancelReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userA))

	cancelResp, err := http.DefaultClient.Do(cancelReq)
	require.NoError(t, err)
	defer func() { _ = cancelResp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, cancelResp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// GetDeals
// ────────────────────────────────────────────────────────────────

func TestGetDealsUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.Get(dealsURL() + "/deals/")
	require.NoError(t, err)
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

// ────────────────────────────────────────────────────────────────
// GetDealByID
// ────────────────────────────────────────────────────────────────

func TestGetDealByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	dealIDs := mustGetDealIDs(t, userID, true)
	require.Len(t, dealIDs, 1)

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/"+dealIDs[0].String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal dealstypes.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	require.Equal(t, dealIDs[0], deal.Id)
	require.False(t, deal.CreatedAt.IsZero())
	require.Len(t, deal.Items, 1)
	require.NotEqual(t, uuid.Nil, deal.Items[0].Id)
	require.Equal(t, userID, deal.Items[0].AuthorId)
}

func TestGetDealByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/"+uuid.NewString(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDealByIDInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/not-a-uuid", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDealByIDUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.Get(dealsURL() + "/deals/" + uuid.NewString())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// Вспомогательные функции (helpers for deals)
// ────────────────────────────────────────────────────────────────

func mustConfirmDraft(t *testing.T, userID uuid.UUID, draftID uuid.UUID) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustGetDealIDs(t *testing.T, userID uuid.UUID, my bool) []uuid.UUID {
	t.Helper()

	url := dealsURL() + "/deals/"
	if my {
		url += "?my=true"
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result dealstypes.GetDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	ids := make([]uuid.UUID, 0, len(result))
	for _, item := range result {
		ids = append(ids, item.Id)
	}

	return ids
}

func mustCreateDeal(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()

	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	ids := mustGetDealIDs(t, userID, true)
	require.NotEmpty(t, ids)

	return ids[0]
}

func mustRequest(t *testing.T, method, url string, body *bytes.Reader) *http.Request {
	t.Helper()

	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		reqBody = body
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	return req
}

// ────────────────────────────────────────────────────────────────
// GetOffers
// ────────────────────────────────────────────────────────────────

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

	my := true
	result := mustGetOffers(t, userA, &my)
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
	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/offers/?sort=ByTime&my=not-bool", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
