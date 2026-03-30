package integration

import (
	dealstypes "barter-port/contracts/openapi/deals/types"
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
	return globalFixture.ItemsURL
}

func mustCreateItem(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()

	body, err := json.Marshal(dealstypes.CreateItemRequest{
		Name:        fmt.Sprintf("item-%d", time.Now().UnixNano()),
		Description: "test item",
		Type:        dealstypes.Good,
		Action:      dealstypes.Give,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/items/", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var item dealstypes.Item
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	require.NotEqual(t, uuid.Nil, item.Id)

	return item.Id
}

func mustCreateDraft(t *testing.T, userID uuid.UUID, name *string, description *string, items []struct {
	ItemID   uuid.UUID `json:"itemID"`
	Quantity int       `json:"quantity"`
}) uuid.UUID {
	t.Helper()

	reqBody := map[string]any{
		"items":       items,
		"name":        name,
		"description": description,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/drafts", bytes.NewReader(body))
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

func TestCreateDraftSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	body, err := json.Marshal(map[string]any{
		"items": []any{},
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/drafts", bytes.NewReader(body))
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

func TestCreateDraftWithNameAndDescription(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	name := "My Draft Deal"
	description := "This is a test draft"

	body, err := json.Marshal(map[string]any{
		"items":       []any{},
		"name":        name,
		"description": description,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/drafts", bytes.NewReader(body))
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

func TestCreateDraftWithItems(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	itemID := mustCreateItem(t, userID)

	body, err := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"itemID": itemID.String(), "quantity": 2},
		},
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/drafts", bytes.NewReader(body))
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

	body, err := json.Marshal(map[string]any{"items": []any{}})
	require.NoError(t, err)

	resp, err := http.Post(dealsURL()+"/drafts", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDraftInvalidJSON(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodPost, dealsURL()+"/drafts", bytes.NewReader([]byte(`not-json`)))
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

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/my-drafts", nil)
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

func TestGetMyDraftsReturnsOwnDrafts(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	otherUserID := uuid.New()

	id1 := mustCreateDraft(t, userID, nil, nil, nil)
	id2 := mustCreateDraft(t, userID, nil, nil, nil)
	// другой пользователь — не должен появляться в списке
	mustCreateDraft(t, otherUserID, nil, nil, nil)

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/my-drafts", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var ids []uuid.UUID
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ids))
	require.Len(t, ids, 2)
	require.ElementsMatch(t, []uuid.UUID{id1, id2}, ids)
}

func TestGetMyDraftsUnauthorized(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	resp, err := http.Get(dealsURL() + "/my-drafts")
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
	draftID := mustCreateDraft(t, userID, &name, &description, nil)

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/drafts/"+draftID.String(), nil)
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
	require.Empty(t, draft.Items)
	require.False(t, draft.CreatedAt.IsZero())
}

func TestGetDraftByIDWithItems(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	itemID := mustCreateItem(t, userID)

	draftID := mustCreateDraft(t, userID, nil, nil, []struct {
		ItemID   uuid.UUID `json:"itemID"`
		Quantity int       `json:"quantity"`
	}{
		{ItemID: itemID, Quantity: 3},
	})

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/drafts/"+draftID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var draft dealstypes.Draft
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&draft))
	require.Equal(t, draftID, draft.Id)
	require.Len(t, draft.Items, 1)
	require.Equal(t, itemID, draft.Items[0].Id)
}

func TestGetDraftByIDNotFound(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/drafts/"+uuid.NewString(), nil)
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

	req, err := http.NewRequest(http.MethodGet, dealsURL()+"/drafts/not-a-uuid", nil)
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

	resp, err := http.Get(dealsURL() + "/drafts/" + uuid.NewString())
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
