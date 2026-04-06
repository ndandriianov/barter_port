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
	return mustCreateOfferWithAction(t, userID, dealstypes.Give)
}

func mustCreateOfferWithAction(t *testing.T, userID uuid.UUID, action dealstypes.OfferAction) uuid.UUID {
	t.Helper()

	body, err := json.Marshal(dealstypes.CreateOfferRequest{
		Name:        fmt.Sprintf("offer-%d", time.Now().UnixNano()),
		Description: "test offer",
		Type:        dealstypes.Good,
		Action:      action,
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

	draftByBWithOfferA := mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})
	draftByBWithOfferB := mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerB, Quantity: 1}})

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
	_ = mustCreateDraft(t, userB, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerA, Quantity: 1}})

	idsDefault := mustGetDraftIDs(t, userA, nil)
	idsFalse := mustGetDraftIDs(t, userA, new(false))

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

func TestGetDraftsInvalidParticipatingReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts?participating=not-bool", nil)
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

func TestConfirmDraftForbiddenForNonParticipant(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	nonParticipantID := uuid.New()
	offerID := mustCreateOffer(t, ownerID)
	draftID := mustCreateDraft(t, ownerID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, nonParticipantID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestConfirmDraftNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
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

func TestCancelDraftNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/drafts/"+uuid.NewString()+"/cancel", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
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
	require.NotEmpty(t, deal.Status)
	require.True(t, deal.Status.Valid())
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

func mustGetDealByID(t *testing.T, userID uuid.UUID, dealID uuid.UUID) dealstypes.Deal {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/"+dealID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal dealstypes.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	require.NotEmpty(t, deal.Status)
	require.True(t, deal.Status.Valid())
	return deal
}

func mustCreateDeal(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()

	before := mustGetDealIDs(t, userID, true)
	beforeSet := make(map[uuid.UUID]struct{}, len(before))
	for _, id := range before {
		beforeSet[id] = struct{}{}
	}

	offerID := mustCreateOffer(t, userID)
	draftID := mustCreateDraft(t, userID, nil, nil, []dealstypes.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
	mustConfirmDraft(t, userID, draftID)

	after := mustGetDealIDs(t, userID, true)
	require.NotEmpty(t, after)

	for _, id := range after {
		if _, ok := beforeSet[id]; !ok {
			return id
		}
	}

	return after[0]
}

func mustCreateTwoPartyDeal(t *testing.T, userA uuid.UUID, userB uuid.UUID) (uuid.UUID, uuid.UUID) {
	t.Helper()

	beforeA := mustGetDealIDs(t, userA, true)
	beforeSet := make(map[uuid.UUID]struct{}, len(beforeA))
	for _, id := range beforeA {
		beforeSet[id] = struct{}{}
	}

	offerA := mustCreateOfferWithAction(t, userA, dealstypes.Give)
	offerB := mustCreateOfferWithAction(t, userB, dealstypes.Give)
	draftID := mustCreateDraft(t, userA, nil, nil, []dealstypes.OfferIDAndQuantity{
		{OfferID: offerA, Quantity: 1},
		{OfferID: offerB, Quantity: 1},
	})
	mustConfirmDraft(t, userA, draftID)
	mustConfirmDraft(t, userB, draftID)

	afterA := mustGetDealIDs(t, userA, true)
	require.NotEmpty(t, afterA)

	dealID := afterA[0]
	for _, id := range afterA {
		if _, ok := beforeSet[id]; !ok {
			dealID = id
			break
		}
	}

	deal := mustGetDealByID(t, userA, dealID)
	for _, item := range deal.Items {
		if item.AuthorId == userA {
			return dealID, item.Id
		}
	}

	require.FailNow(t, "item authored by userA was not found in created deal")
	return uuid.Nil, uuid.Nil
}

func doChangeDealStatus(t *testing.T, dealID uuid.UUID, userID *uuid.UUID, rawBody []byte) *http.Response {
	t.Helper()

	if rawBody == nil {
		rawBody = []byte(`{}`)
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		dealsURL()+"/deals/"+dealID.String()+"/status",
		bytes.NewReader(rawBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if userID != nil {
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, *userID))
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func doGetDealStatusVotes(t *testing.T, dealID string, userID *uuid.UUID) *http.Response {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodGet,
		dealsURL()+"/deals/"+dealID+"/status",
		nil,
	)
	require.NoError(t, err)
	if userID != nil {
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, *userID))
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func doAddDealItem(t *testing.T, dealID string, userID *uuid.UUID, rawBody []byte) *http.Response {
	t.Helper()

	if rawBody == nil {
		rawBody = []byte(`{}`)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		dealsURL()+"/deals/"+dealID+"/items",
		bytes.NewReader(rawBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if userID != nil {
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, *userID))
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func mustChangeDealStatus(t *testing.T, dealID uuid.UUID, userID uuid.UUID, status dealstypes.DealStatus) dealstypes.Deal {
	t.Helper()

	body, err := json.Marshal(dealstypes.ChangeDealStatusRequest{ExpectedStatus: status})
	require.NoError(t, err)

	resp := doChangeDealStatus(t, dealID, &userID, body)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal dealstypes.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	return deal
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

// ----------------------------------------------------------------
// AddDealItem
// ----------------------------------------------------------------

func TestAddDealItemUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	body := []byte(`{"offerId":"` + uuid.NewString() + `","quantity":1}`)
	resp := doAddDealItem(t, uuid.NewString(), nil, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAddDealItemInvalidDealID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	body := []byte(`{"offerId":"` + uuid.NewString() + `","quantity":1}`)
	resp := doAddDealItem(t, "not-a-uuid", &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddDealItemSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)

	body := []byte(`{"offerId":"` + offerID.String() + `","quantity":2}`)
	resp := doAddDealItem(t, dealID.String(), &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal dealstypes.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	var found bool
	for _, item := range deal.Items {
		if item.AuthorId == userID && item.Quantity == 2 {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestAddDealItemNotParticipantForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	strangerID := uuid.New()
	dealID := mustCreateDeal(t, ownerID)
	offerID := mustCreateOffer(t, strangerID)

	body := []byte(`{"offerId":"` + offerID.String() + `","quantity":1}`)
	resp := doAddDealItem(t, dealID.String(), &strangerID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAddDealItemOfferNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	body := []byte(`{"offerId":"` + uuid.NewString() + `","quantity":1}`)
	resp := doAddDealItem(t, dealID.String(), &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAddDealItemForeignOfferForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)
	offerB := mustCreateOffer(t, userB)

	body := []byte(`{"offerId":"` + offerB.String() + `","quantity":1}`)
	resp := doAddDealItem(t, dealID.String(), &userA, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAddDealItemInvalidQuantity(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)

	body := []byte(`{"offerId":"` + offerID.String() + `","quantity":0}`)
	resp := doAddDealItem(t, dealID.String(), &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddDealItemClosedDealForbidden(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	offerID := mustCreateOffer(t, userID)

	mustChangeDealStatus(t, dealID, userID, dealstypes.Cancelled)

	body := []byte(`{"offerId":"` + offerID.String() + `","quantity":1}`)
	resp := doAddDealItem(t, dealID.String(), &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────
// UpdateDealItem
// ────────────────────────────────────────────────────────────────

func TestUpdateDealItemUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	body, err := json.Marshal(dealstypes.UpdateDealItemRequest{Name: func() *string { ; return new("x") }()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDealItemEmptyPatchReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	deal := mustGetDealByID(t, userID, dealID)
	require.NotEmpty(t, deal.Items)

	body, err := json.Marshal(dealstypes.UpdateDealItemRequest{})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+deal.Items[0].Id.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateDealItemAuthorCanEditContent(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)
	deal := mustGetDealByID(t, userID, dealID)
	require.NotEmpty(t, deal.Items)

	newName := "updated item"
	newDescription := "updated description"
	newQty := 7
	body, err := json.Marshal(dealstypes.UpdateDealItemRequest{
		Name:        &newName,
		Description: &newDescription,
		Quantity:    &newQty,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+deal.Items[0].Id.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var item dealstypes.Item
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	require.Equal(t, deal.Items[0].Id, item.Id)
	require.Equal(t, newName, item.Name)
	require.Equal(t, newDescription, item.Description)
	require.EqualValues(t, newQty, item.Quantity)
}

func TestUpdateDealItemNonAuthorCannotEditContent(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)

	body, err := json.Marshal(dealstypes.UpdateDealItemRequest{Name: new("forbidden update")})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userB))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateDealItemParticipantCanClaimAndReleaseReceiver(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, itemIDByA := mustCreateTwoPartyDeal(t, userA, userB)

	claimBody, err := json.Marshal(dealstypes.UpdateDealItemRequest{ClaimReceiver: new(true)})
	require.NoError(t, err)

	claimReq, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String(), bytes.NewReader(claimBody))
	require.NoError(t, err)
	claimReq.Header.Set("Content-Type", "application/json")
	claimReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userB))

	claimResp, err := http.DefaultClient.Do(claimReq)
	require.NoError(t, err)
	defer func() { _ = claimResp.Body.Close() }()
	require.Equal(t, http.StatusOK, claimResp.StatusCode)

	var claimed dealstypes.Item
	require.NoError(t, json.NewDecoder(claimResp.Body).Decode(&claimed))
	require.NotNil(t, claimed.ReceiverId)
	require.Equal(t, userB, *claimed.ReceiverId)

	releaseBody, err := json.Marshal(dealstypes.UpdateDealItemRequest{ReleaseReceiver: new(true)})
	require.NoError(t, err)

	releaseReq, err := http.NewRequest(http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/items/"+itemIDByA.String(), bytes.NewReader(releaseBody))
	require.NoError(t, err)
	releaseReq.Header.Set("Content-Type", "application/json")
	releaseReq.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userB))

	releaseResp, err := http.DefaultClient.Do(releaseReq)
	require.NoError(t, err)
	defer func() { _ = releaseResp.Body.Close() }()
	require.Equal(t, http.StatusOK, releaseResp.StatusCode)

	var released dealstypes.Item
	require.NoError(t, json.NewDecoder(releaseResp.Body).Decode(&released))
	require.Nil(t, released.ReceiverId)
}

// ────────────────────────────────────────────────────────────────
// ChangeDealStatus
// ────────────────────────────────────────────────────────────────

func TestChangeDealStatusUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	body := []byte(`{"expectedStatus":"Discussion"}`)
	resp := doChangeDealStatus(t, uuid.New(), nil, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestChangeDealStatusInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	body := []byte(`{"expectedStatus":"Discussion"}`)

	req, err := http.NewRequest(
		http.MethodPatch,
		dealsURL()+"/deals/not-a-uuid/status",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	body := []byte(`{"expectedStatus":"Discussion"}`)
	resp := doChangeDealStatus(t, uuid.New(), &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestChangeDealStatusInvalidJSONReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	resp := doChangeDealStatus(t, dealID, &userID, []byte(`not-json`))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusUnknownStatusReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	resp := doChangeDealStatus(t, dealID, &userID, []byte(`{"expectedStatus":"unknown"}`))
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusInvalidTransitionReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	dealID := mustCreateDeal(t, userID)

	body, err := json.Marshal(dealstypes.ChangeDealStatusRequest{ExpectedStatus: dealstypes.Confirmed})
	require.NoError(t, err)

	resp := doChangeDealStatus(t, dealID, &userID, body)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeDealStatusConsensusMovesToDiscussion(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	firstVote := mustChangeDealStatus(t, dealID, userA, dealstypes.Discussion)
	require.Equal(t, dealstypes.LookingForParticipants, firstVote.Status)

	secondVote := mustChangeDealStatus(t, dealID, userB, dealstypes.Discussion)
	require.Equal(t, dealstypes.Discussion, secondVote.Status)
}

func TestChangeDealStatusCancelledAppliesImmediately(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	updated := mustChangeDealStatus(t, dealID, userA, dealstypes.Cancelled)
	require.Equal(t, dealstypes.Cancelled, updated.Status)

	dealAfter := mustGetDealByID(t, userB, dealID)
	require.Equal(t, dealstypes.Cancelled, dealAfter.Status)
}

// ────────────────────────────────────────────────────────────────
// GetDealStatusVotes
// ────────────────────────────────────────────────────────────────

func TestGetDealStatusVotesUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp := doGetDealStatusVotes(t, uuid.NewString(), nil)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetDealStatusVotesInvalidUUID(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	resp := doGetDealStatusVotes(t, "not-a-uuid", &userID)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetDealStatusVotesNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	resp := doGetDealStatusVotes(t, uuid.NewString(), &userID)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDealStatusVotesEmptyWhenNoVotes(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	resp := doGetDealStatusVotes(t, dealID.String(), &userA)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var votes dealstypes.GetDealStatusVotesResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&votes))
	require.Empty(t, votes)
}

func TestGetDealStatusVotesReturnsRecordedVotes(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userA := uuid.New()
	userB := uuid.New()
	dealID, _ := mustCreateTwoPartyDeal(t, userA, userB)

	_ = mustChangeDealStatus(t, dealID, userA, dealstypes.Discussion)

	resp := doGetDealStatusVotes(t, dealID.String(), &userB)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var votes dealstypes.GetDealStatusVotesResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&votes))
	require.Len(t, votes, 1)
	require.Equal(t, userA, votes[0].UserId)
	require.Equal(t, dealstypes.Discussion, votes[0].Vote)
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
	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/offers/?sort=ByTime&my=not-bool", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
