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

func TestCreateDraftWithItemsAndNameAndDescription(t *testing.T) {
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

func TestCreateDraftWithOffers(t *testing.T) {
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
// GetMyDrafts
// ────────────────────────────────────────────────────────────────

func TestGetMyDraftsEmpty(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/deals/drafts/my", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var ids []uuid.UUID
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ids))
	require.Empty(t, ids)
}

func TestGetMyDraftsUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.Get(dealsURL() + "/deals/drafts/my")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
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
