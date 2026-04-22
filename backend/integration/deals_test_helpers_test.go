package integration

import (
	"barter-port/contracts/openapi/deals/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
	0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

type reviewedDealContext struct {
	DealID      uuid.UUID
	ProviderID  uuid.UUID
	ReceiverID  uuid.UUID
	ItemID      uuid.UUID
	OtherItemID uuid.UUID
	OfferID     uuid.UUID
	Review      types.Review
}

func dumpDealsLogs(t *testing.T) {
	t.Helper()
	DumpLogsOnFailure(t, globalFixture.Items, "deals")
}

func dealsURL() string {
	return globalFixture.DealsURL
}

func usersURL() string {
	return globalFixture.UsersURL
}

func mustJSONBody(t *testing.T, v any) io.Reader {
	t.Helper()

	body, err := json.Marshal(v)
	require.NoError(t, err)

	return bytes.NewReader(body)
}

func mustDo(t *testing.T, req *http.Request) *http.Response {
	t.Helper()

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func mustRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	return req
}

func mustBearerRequest(t *testing.T, method, url, token string, body io.Reader) *http.Request {
	t.Helper()

	req := mustRequest(t, method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func mustUserRequest(t *testing.T, method, url string, userID uuid.UUID, body io.Reader) *http.Request {
	t.Helper()

	return mustBearerRequest(t, method, url, mustAccessToken(t, userID), body)
}

func mustCreateOffer(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()
	return mustCreateOfferWithAction(t, userID, types.Give)
}

func mustCreateOfferWithAction(t *testing.T, userID uuid.UUID, action types.OfferAction) uuid.UUID {
	t.Helper()

	offer := mustCreateOfferDetails(t, userID, types.CreateOfferRequest{
		Name:        fmt.Sprintf("offer-%d", time.Now().UnixNano()),
		Description: "test offer",
		Type:        types.Good,
		Action:      action,
	})

	return offer.Id
}

func mustCreateOfferDetails(t *testing.T, userID uuid.UUID, body types.CreateOfferRequest) types.Offer {
	t.Helper()

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers", userID, mustJSONBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var offer types.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))
	require.NotEqual(t, uuid.Nil, offer.Id)

	return offer
}

func mustCreateOfferMultipartDetails(t *testing.T, userID uuid.UUID, body types.CreateOfferRequest, photos [][]byte) types.Offer {
	t.Helper()

	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	require.NoError(t, writer.WriteField("name", body.Name))
	require.NoError(t, writer.WriteField("description", body.Description))
	require.NoError(t, writer.WriteField("type", string(body.Type)))
	require.NoError(t, writer.WriteField("action", string(body.Action)))
	if body.Tags != nil {
		for _, tag := range *body.Tags {
			require.NoError(t, writer.WriteField("tags", string(tag)))
		}
	}

	for i, photo := range photos {
		part, err := writer.CreateFormFile("photos", fmt.Sprintf("photo-%d.png", i))
		require.NoError(t, err)
		_, err = part.Write(photo)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers", userID, &payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var offer types.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))
	require.NotEqual(t, uuid.Nil, offer.Id)

	return offer
}

func mustGetOfferByID(t *testing.T, userID uuid.UUID, offerID uuid.UUID) types.Offer {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/"+offerID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var offer types.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))

	return offer
}

func mustGetOffers(t *testing.T, userID uuid.UUID, my *bool) types.ListOffersResponse {
	t.Helper()

	return mustGetOffersBySort(t, userID, "ByTime", my)
}

func mustGetOffersBySort(t *testing.T, userID uuid.UUID, sort string, my *bool) types.ListOffersResponse {
	return mustGetOffersBySortAndTags(t, userID, sort, my, nil)
}

func mustGetOffersBySortAndTags(t *testing.T, userID uuid.UUID, sort string, my *bool, tags *[]types.TagName) types.ListOffersResponse {
	t.Helper()

	url := dealsURL() + "/offers?sort=" + sort + "&cursor_limit=100"
	if my != nil {
		url += fmt.Sprintf("&my=%t", *my)
	}

	var body io.Reader
	if tags != nil {
		body = mustJSONBody(t, map[string]any{"tags": tags})
	}

	req := mustUserRequest(t, http.MethodGet, url, userID, body)
	if tags != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.ListOffersResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func mustGetSubscribedOffers(t *testing.T, userID uuid.UUID) types.ListOffersResponse {
	t.Helper()

	return mustGetSubscribedOffersBySort(t, userID, "ByTime")
}

func mustGetSubscribedOffersBySort(t *testing.T, userID uuid.UUID, sort string) types.ListOffersResponse {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/offers/subscriptions?sort="+sort+"&cursor_limit=100", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.ListOffersResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func mustSubscribeToUser(t *testing.T, subscriberID, targetUserID uuid.UUID) {
	t.Helper()

	req := mustUserRequest(t, http.MethodPost, usersURL()+"/users/subscriptions", subscriberID, mustJSONBody(t, map[string]uuid.UUID{
		"targetUserId": targetUserID,
	}))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func mustViewOfferByID(t *testing.T, userID uuid.UUID, offerID uuid.UUID) {
	t.Helper()

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/offers/"+offerID.String()+"/view", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustUpdateOffer(t *testing.T, userID uuid.UUID, offerID uuid.UUID, body types.UpdateOfferRequest) types.Offer {
	t.Helper()

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), userID, mustJSONBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var offer types.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))

	return offer
}

func mustUpdateOfferMultipartDetails(
	t *testing.T,
	userID uuid.UUID,
	offerID uuid.UUID,
	body types.UpdateOfferRequest,
	photos [][]byte,
) types.Offer {
	t.Helper()

	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	if body.Name != nil {
		require.NoError(t, writer.WriteField("name", *body.Name))
	}
	if body.Description != nil {
		require.NoError(t, writer.WriteField("description", *body.Description))
	}
	if body.Type != nil {
		require.NoError(t, writer.WriteField("type", string(*body.Type)))
	}
	if body.Action != nil {
		require.NoError(t, writer.WriteField("action", string(*body.Action)))
	}
	if body.Tags != nil {
		for _, tag := range *body.Tags {
			require.NoError(t, writer.WriteField("tags", string(tag)))
		}
	}
	if body.DeletePhotoIds != nil {
		for _, photoID := range *body.DeletePhotoIds {
			require.NoError(t, writer.WriteField("deletePhotoIds", photoID.String()))
		}
	}

	for i, photo := range photos {
		part, err := writer.CreateFormFile("photos", fmt.Sprintf("photo-%d.png", i))
		require.NoError(t, err)
		_, err = part.Write(photo)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/offers/"+offerID.String(), userID, &payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var offer types.Offer
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&offer))

	return offer
}

func mustListTags(t *testing.T, userID uuid.UUID) types.ListTagsResponse {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/tags", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tags types.ListTagsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tags))

	return tags
}

func mustDeleteAdminTag(t *testing.T, adminToken string, name string) {
	t.Helper()

	req := mustBearerRequest(t, http.MethodDelete, dealsURL()+"/admin/tags?name="+url.QueryEscape(name), adminToken, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func mustDeleteOffer(t *testing.T, userID uuid.UUID, offerID uuid.UUID) {
	t.Helper()

	req := mustUserRequest(t, http.MethodDelete, dealsURL()+"/offers/"+offerID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func mustCreateDraft(
	t *testing.T,
	userID uuid.UUID,
	name *string,
	description *string,
	offers []types.OfferIDAndQuantity,
) uuid.UUID {
	t.Helper()

	reqBody := map[string]any{
		"offers":      offers,
		"name":        name,
		"description": description,
	}

	req := mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/drafts", userID, mustJSONBody(t, reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created types.CreateDraftDealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEqual(t, uuid.Nil, created.Id)

	return created.Id
}

func mustGetDraftIDs(t *testing.T, userID uuid.UUID, createdByMe *bool) []uuid.UUID {
	t.Helper()

	url := dealsURL() + "/deals/drafts"
	if createdByMe != nil {
		url += fmt.Sprintf("?createdByMe=%t", *createdByMe)
	}

	req := mustUserRequest(t, http.MethodGet, url, userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var drafts types.GetMyDraftDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&drafts))

	ids := make([]uuid.UUID, 0, len(drafts))
	for _, draft := range drafts {
		ids = append(ids, draft.Id)
	}

	return ids
}

func mustConfirmDraft(t *testing.T, userID uuid.UUID, draftID uuid.UUID) {
	t.Helper()

	req := mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/drafts/"+draftID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustGetDealIDs(t *testing.T, userID uuid.UUID, my bool) []uuid.UUID {
	t.Helper()

	result := mustGetDealsResponse(t, userID, my)
	ids := make([]uuid.UUID, 0, len(result))
	for _, item := range result {
		ids = append(ids, item.Id)
	}

	return ids
}

func mustGetDealsResponse(t *testing.T, userID uuid.UUID, my bool) types.GetDealsResponse {
	t.Helper()

	url := dealsURL() + "/deals"
	if my {
		url += "?my=true"
	}

	req := mustUserRequest(t, http.MethodGet, url, userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.GetDealsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func mustGetDealByID(t *testing.T, userID uuid.UUID, dealID uuid.UUID) types.Deal {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID.String(), userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal types.Deal
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
	draftID := mustCreateDraft(t, userID, nil, nil, []types.OfferIDAndQuantity{{OfferID: offerID, Quantity: 1}})
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

	offerA := mustCreateOfferWithAction(t, userA, types.Give)
	offerB := mustCreateOfferWithAction(t, userB, types.Give)
	draftID := mustCreateDraft(t, userA, nil, nil, []types.OfferIDAndQuantity{
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

func mustGetDealItemIDByAuthor(t *testing.T, viewerID uuid.UUID, dealID uuid.UUID, authorID uuid.UUID) uuid.UUID {
	t.Helper()

	deal := mustGetDealByID(t, viewerID, dealID)
	for _, item := range deal.Items {
		if item.AuthorId == authorID {
			return item.Id
		}
	}

	require.FailNow(t, "item authored by user was not found", "deal_id=%s author_id=%s", dealID, authorID)
	return uuid.Nil
}

func mustGetDealItemIDsByAuthor(t *testing.T, viewerID uuid.UUID, dealID uuid.UUID) map[uuid.UUID]uuid.UUID {
	t.Helper()

	deal := mustGetDealByID(t, viewerID, dealID)
	result := make(map[uuid.UUID]uuid.UUID, len(deal.Items))
	for _, item := range deal.Items {
		result[item.AuthorId] = item.Id
	}

	return result
}

func doChangeDealStatus(t *testing.T, dealID uuid.UUID, userID *uuid.UUID, rawBody []byte) *http.Response {
	t.Helper()

	if rawBody == nil {
		rawBody = []byte(`{}`)
	}

	var req *http.Request
	if userID == nil {
		req = mustRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/status", bytes.NewReader(rawBody))
	} else {
		req = mustUserRequest(t, http.MethodPatch, dealsURL()+"/deals/"+dealID.String()+"/status", *userID, bytes.NewReader(rawBody))
	}
	req.Header.Set("Content-Type", "application/json")

	return mustDo(t, req)
}

func doGetDealStatusVotes(t *testing.T, dealID string, userID *uuid.UUID) *http.Response {
	t.Helper()

	if userID == nil {
		return mustDo(t, mustRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID+"/status", nil))
	}

	return mustDo(t, mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID+"/status", *userID, nil))
}

func doAddDealItem(t *testing.T, dealID string, userID *uuid.UUID, rawBody []byte) *http.Response {
	t.Helper()

	if rawBody == nil {
		rawBody = []byte(`{}`)
	}

	var req *http.Request
	if userID == nil {
		req = mustRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID+"/items", bytes.NewReader(rawBody))
	} else {
		req = mustUserRequest(t, http.MethodPost, dealsURL()+"/deals/"+dealID+"/items", *userID, bytes.NewReader(rawBody))
	}
	req.Header.Set("Content-Type", "application/json")

	return mustDo(t, req)
}

func mustChangeDealStatus(t *testing.T, dealID uuid.UUID, userID uuid.UUID, status types.DealStatus) types.Deal {
	t.Helper()

	body, err := json.Marshal(types.ChangeDealStatusRequest{ExpectedStatus: status})
	require.NoError(t, err)

	resp := doChangeDealStatus(t, dealID, &userID, body)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var deal types.Deal
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	return deal
}

func mustUpdateDealItem(
	t *testing.T,
	userID uuid.UUID,
	dealID uuid.UUID,
	itemID uuid.UUID,
	body types.UpdateDealItemRequest,
) types.Item {
	t.Helper()

	req := mustUserRequest(
		t,
		http.MethodPatch,
		dealsURL()+"/deals/"+dealID.String()+"/items/"+itemID.String(),
		userID,
		mustJSONBody(t, body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var item types.Item
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))

	return item
}

func mustUpdateDealItemMultipartDetails(
	t *testing.T,
	userID uuid.UUID,
	dealID uuid.UUID,
	itemID uuid.UUID,
	body types.UpdateDealItemRequest,
	photos [][]byte,
) types.Item {
	t.Helper()

	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)

	if body.Name != nil {
		require.NoError(t, writer.WriteField("name", *body.Name))
	}
	if body.Description != nil {
		require.NoError(t, writer.WriteField("description", *body.Description))
	}
	if body.Quantity != nil {
		require.NoError(t, writer.WriteField("quantity", fmt.Sprintf("%d", *body.Quantity)))
	}
	if body.ClaimProvider != nil {
		require.NoError(t, writer.WriteField("claimProvider", fmt.Sprintf("%t", *body.ClaimProvider)))
	}
	if body.ReleaseProvider != nil {
		require.NoError(t, writer.WriteField("releaseProvider", fmt.Sprintf("%t", *body.ReleaseProvider)))
	}
	if body.ClaimReceiver != nil {
		require.NoError(t, writer.WriteField("claimReceiver", fmt.Sprintf("%t", *body.ClaimReceiver)))
	}
	if body.ReleaseReceiver != nil {
		require.NoError(t, writer.WriteField("releaseReceiver", fmt.Sprintf("%t", *body.ReleaseReceiver)))
	}
	if body.DeletePhotoIds != nil {
		for _, photoID := range *body.DeletePhotoIds {
			require.NoError(t, writer.WriteField("deletePhotoIds", photoID.String()))
		}
	}

	for i, photo := range photos {
		part, err := writer.CreateFormFile("photos", fmt.Sprintf("photo-%d.png", i))
		require.NoError(t, err)
		_, err = part.Write(photo)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := mustUserRequest(
		t,
		http.MethodPatch,
		dealsURL()+"/deals/"+dealID.String()+"/items/"+itemID.String(),
		userID,
		&payload,
	)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var item types.Item
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))

	return item
}

func mustCreateDiscussionDeal(t *testing.T, userIDs ...uuid.UUID) (uuid.UUID, map[uuid.UUID]uuid.UUID) {
	t.Helper()
	require.GreaterOrEqual(t, len(userIDs), 2)

	before := mustGetDealIDs(t, userIDs[0], true)
	beforeSet := make(map[uuid.UUID]struct{}, len(before))
	for _, id := range before {
		beforeSet[id] = struct{}{}
	}

	offers := make([]types.OfferIDAndQuantity, 0, len(userIDs))
	for _, userID := range userIDs {
		offers = append(offers, types.OfferIDAndQuantity{
			OfferID:  mustCreateOfferWithAction(t, userID, types.Give),
			Quantity: 1,
		})
	}

	draftID := mustCreateDraft(t, userIDs[0], nil, nil, offers)
	for _, userID := range userIDs {
		mustConfirmDraft(t, userID, draftID)
	}

	after := mustGetDealIDs(t, userIDs[0], true)
	require.NotEmpty(t, after)

	dealID := after[0]
	for _, id := range after {
		if _, ok := beforeSet[id]; !ok {
			dealID = id
			break
		}
	}

	itemIDsByAuthor := mustGetDealItemIDsByAuthor(t, userIDs[0], dealID)
	for idx, authorID := range userIDs {
		receiverID := userIDs[(idx+1)%len(userIDs)]
		mustUpdateDealItem(t, receiverID, dealID, itemIDsByAuthor[authorID], types.UpdateDealItemRequest{
			ClaimReceiver: new(true),
		})
	}

	for _, userID := range userIDs {
		_ = mustChangeDealStatus(t, dealID, userID, types.Discussion)
	}

	return dealID, itemIDsByAuthor
}

func mustCreateCompletedReviewableTwoPartyDeal(t *testing.T, userA uuid.UUID, userB uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	dealID, itemIDsByAuthor := mustCreateDiscussionDeal(t, userA, userB)
	itemIDByA := itemIDsByAuthor[userA]
	itemIDByB := itemIDsByAuthor[userB]

	_ = mustUpdateDealItem(t, userA, dealID, itemIDByA, types.UpdateDealItemRequest{
		Name: new(fmt.Sprintf("updated-item-%d", time.Now().UnixNano())),
	})

	_ = mustChangeDealStatus(t, dealID, userA, types.Confirmed)
	_ = mustChangeDealStatus(t, dealID, userB, types.Confirmed)
	_ = mustChangeDealStatus(t, dealID, userA, types.Completed)
	_ = mustChangeDealStatus(t, dealID, userB, types.Completed)

	return dealID, itemIDByA, itemIDByB
}

func mustCreateDealItemReview(
	t *testing.T,
	userID uuid.UUID,
	dealID uuid.UUID,
	itemID uuid.UUID,
	rating int,
	comment *string,
) types.Review {
	t.Helper()

	req := mustUserRequest(
		t,
		http.MethodPost,
		dealsURL()+"/deals/"+dealID.String()+"/items/"+itemID.String()+"/reviews",
		userID,
		mustJSONBody(t, types.CreateReviewRequest{
			Rating:  rating,
			Comment: comment,
		}),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var review types.Review
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&review))

	return review
}

func mustCreateReviewedOfferItemContext(t *testing.T) reviewedDealContext {
	t.Helper()

	return mustCreateReviewedOfferItemContextWithUsers(t, uuid.New(), uuid.New())
}

func mustCreateReviewedOfferItemContextWithRegisteredUsers(t *testing.T) reviewedDealContext {
	t.Helper()

	fixture := globalFixture
	providerID := mustRegisterProjectedUser(t, fixture)
	receiverID := mustRegisterProjectedUser(t, fixture)

	return mustCreateReviewedOfferItemContextWithUsers(t, providerID, receiverID)
}

func mustCreateReviewedOfferItemContextWithUsers(t *testing.T, providerID uuid.UUID, receiverID uuid.UUID) reviewedDealContext {
	t.Helper()

	dealID, itemID, otherItemID := mustCreateCompletedReviewableTwoPartyDeal(t, providerID, receiverID)
	review := mustCreateDealItemReview(t, receiverID, dealID, itemID, 5, new("excellent"))

	require.NotNil(t, review.OfferRef)

	return reviewedDealContext{
		DealID:      dealID,
		ProviderID:  providerID,
		ReceiverID:  receiverID,
		ItemID:      itemID,
		OtherItemID: otherItemID,
		OfferID:     review.OfferRef.OfferId,
		Review:      review,
	}
}

func mustGetJoinRequests(t *testing.T, userID uuid.UUID, dealID uuid.UUID) types.GetDealJoinRequestsResponse {
	t.Helper()

	req := mustUserRequest(t, http.MethodGet, dealsURL()+"/deals/"+dealID.String()+"/joins", userID, nil)
	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result types.GetDealJoinRequestsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func mustVoteForFailure(t *testing.T, userID uuid.UUID, dealID uuid.UUID, votedForID uuid.UUID) {
	t.Helper()

	req := mustUserRequest(
		t,
		http.MethodPost,
		dealsURL()+"/deals/failures/"+dealID.String()+"/votes",
		userID,
		mustJSONBody(t, types.VoteForFailureRequest{UserId: votedForID}),
	)
	req.Header.Set("Content-Type", "application/json")

	resp := mustDo(t, req)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
