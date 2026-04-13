package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"barter-port/internal/deals/domain"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type offerGroupResponse struct {
	Id          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description"`
	Units       []offerGroupUnitResponse `json:"units"`
}

type offerGroupUnitResponse struct {
	Id     uuid.UUID     `json:"id"`
	Offers []types.Offer `json:"offers"`
}

func TestCreateOfferGroupSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerA := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Велосипед",
		Description: "test offer",
		Type:        types.Good,
		Action:      types.Give,
	}).Id
	offerB := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Шлем",
		Description: "test offer",
		Type:        types.Good,
		Action:      types.Give,
	}).Id
	offerC := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        "Ремонт",
		Description: "test offer",
		Type:        types.Service,
		Action:      types.Give,
	}).Id
	description := "group description"

	group := mustCreateOfferGroup(t, userID, map[string]any{
		"description": description,
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerA.String()},
					{"offerId": offerB.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerC.String()},
				},
			},
		},
	})

	require.NotEqual(t, uuid.Nil, group.Id)
	require.Equal(t, "Велосипед и Шлем, Ремонт", group.Name)
	require.NotNil(t, group.Description)
	require.Equal(t, description, *group.Description)
	require.Len(t, group.Units, 2)
	require.Len(t, group.Units[0].Offers, 2)
	require.Len(t, group.Units[1].Offers, 1)
}

func TestListOfferGroupsReturnsCreatedGroup(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	group := mustCreateOfferGroup(t, userID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": mustCreateOffer(t, userID).String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offer-groups", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var items []offerGroupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&items))

	found := false
	for _, item := range items {
		if item.Id == group.Id {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestCreateOfferGroupWithMixedActionsInUnitReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	offerGive := mustCreateOfferWithAction(t, userID, types.Give)
	offerTake := mustCreateOfferWithAction(t, userID, types.Take)

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups", userID, mustJSONBody(t, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerGive.String()},
					{"offerId": offerTake.String()},
				},
			},
		},
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.NotNil(t, apiErr.Message)
	require.Equal(t, domain.ErrMixedOfferActionsInUnit.Error(), *apiErr.Message)
}

func TestGetOfferGroupByIDSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	userID := uuid.New()
	group := mustCreateOfferGroup(t, userID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": mustCreateOffer(t, userID).String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offer-groups/"+group.Id.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var item offerGroupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	require.Equal(t, group.Id, item.Id)
	require.Equal(t, group.Name, item.Name)
}

func TestCreateDraftFromOfferGroupSuccess(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	responderID := uuid.New()
	offerA := mustCreateOffer(t, ownerID)
	offerB := mustCreateOffer(t, ownerID)
	offerC := mustCreateOffer(t, ownerID)
	responderOffer := mustCreateOfferWithAction(t, responderID, types.Give)
	group := mustCreateOfferGroup(t, ownerID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerA.String()},
					{"offerId": offerB.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerC.String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups/"+group.Id.String()+"/drafts", responderID, mustJSONBody(t, map[string]any{
		"selectedOfferIds": []string{offerB.String(), offerC.String()},
		"responderOfferId": responderOffer.String(),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created types.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)

	draftReq := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/drafts/"+created.Id.String(), responderID, nil)
	draftResp := mustDo(t, draftReq)
	defer func() { _ = draftResp.Body.Close() }()
	require.Equal(t, http.StatusOK, draftResp.StatusCode)

	var draft types.Draft
	require.NoError(t, json.NewDecoder(draftResp.Body).Decode(&draft))
	require.Equal(t, responderID, draft.AuthorId)
	require.Len(t, draft.Offers, 3)

	got := map[uuid.UUID]int{}
	for _, offer := range draft.Offers {
		got[offer.Id] = offer.Quantity
	}
	require.Equal(t, 1, got[offerB])
	require.Equal(t, 1, got[offerC])
	require.Equal(t, 1, got[responderOffer])
}

func TestCreateDraftFromUniformOfferGroupWithoutResponderOfferReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	responderID := uuid.New()
	offerA := mustCreateOffer(t, ownerID)
	offerB := mustCreateOffer(t, ownerID)
	group := mustCreateOfferGroup(t, ownerID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerA.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerB.String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups/"+group.Id.String()+"/drafts", responderID, mustJSONBody(t, map[string]any{
		"selectedOfferIds": []string{offerA.String(), offerB.String()},
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.NotNil(t, apiErr.Message)
	require.Equal(t, domain.ErrOfferGroupResponderOfferRequired.Error(), *apiErr.Message)
}

func TestCreateDraftFromMixedActionOfferGroupWithoutResponderOfferReturnsCreated(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	responderID := uuid.New()
	offerGive := mustCreateOfferWithAction(t, ownerID, types.Give)
	offerTake := mustCreateOfferWithAction(t, ownerID, types.Take)
	group := mustCreateOfferGroup(t, ownerID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerGive.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerTake.String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups/"+group.Id.String()+"/drafts", responderID, mustJSONBody(t, map[string]any{
		"selectedOfferIds": []string{offerGive.String(), offerTake.String()},
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateDraftFromUniformOfferGroupWithDifferentActionResponderOfferReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	responderID := uuid.New()
	offerA := mustCreateOffer(t, ownerID)
	offerB := mustCreateOffer(t, ownerID)
	offerC := mustCreateOffer(t, ownerID)
	responderOffer := mustCreateOfferWithAction(t, responderID, types.Take)
	group := mustCreateOfferGroup(t, ownerID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerA.String()},
					{"offerId": offerB.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerC.String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups/"+group.Id.String()+"/drafts", responderID, mustJSONBody(t, map[string]any{
		"selectedOfferIds": []string{offerA.String(), offerC.String()},
		"responderOfferId": responderOffer.String(),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.NotNil(t, apiErr.Message)
	require.Equal(t, domain.ErrOfferGroupResponderOfferAction.Error(), *apiErr.Message)
}

func TestCreateDraftFromOfferGroupInvalidSelectionReturnsBadRequest(t *testing.T) {
	t.Parallel()
	dumpDealsLogs(t)

	ownerID := uuid.New()
	responderID := uuid.New()
	offerA := mustCreateOfferWithAction(t, ownerID, types.Give)
	offerB := mustCreateOfferWithAction(t, ownerID, types.Give)
	offerC := mustCreateOfferWithAction(t, ownerID, types.Take)
	responderOffer := mustCreateOfferWithAction(t, responderID, types.Take)
	group := mustCreateOfferGroup(t, ownerID, map[string]any{
		"units": []map[string]any{
			{
				"offers": []map[string]any{
					{"offerId": offerA.String()},
					{"offerId": offerB.String()},
				},
			},
			{
				"offers": []map[string]any{
					{"offerId": offerC.String()},
				},
			},
		},
	})

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups/"+group.Id.String()+"/drafts", responderID, mustJSONBody(t, map[string]any{
		"selectedOfferIds": []string{offerA.String(), offerB.String()},
		"responderOfferId": responderOffer.String(),
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var apiErr types.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&apiErr))
	require.NotNil(t, apiErr.Message)
	require.Equal(t, domain.ErrInvalidOfferGroupSelect.Error(), *apiErr.Message)
}

func mustCreateOfferGroup(t *testing.T, userID uuid.UUID, body map[string]any) offerGroupResponse {
	t.Helper()

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offer-groups", userID, mustJSONBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var item offerGroupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	require.NotEqual(t, uuid.Nil, item.Id)

	return item
}
