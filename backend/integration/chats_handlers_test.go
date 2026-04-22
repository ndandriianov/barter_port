package integration

import (
	chattypes "barter-port/contracts/openapi/chats/types"
	usertypes "barter-port/contracts/openapi/users/types"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func dumpChatsLogs(t *testing.T) {
	t.Helper()
	DumpLogsOnFailure(t, globalFixture.Chats, "chats")
}

func createDirectChatRequest(t *testing.T, fixture *Fixture, requesterID, participantID uuid.UUID) *http.Response {
	t.Helper()

	payload, err := json.Marshal(chattypes.CreateChatRequest{ParticipantId: participantID})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.ChatsURL+"/chats/", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, requesterID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func subscribeUser(t *testing.T, fixture *Fixture, subscriberID, targetUserID uuid.UUID) *http.Response {
	t.Helper()

	payload, err := json.Marshal(usertypes.SubscribeRequest{TargetUserId: targetUserID})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.UsersURL+"/users/subscriptions", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, subscriberID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getSubscriptions(t *testing.T, fixture *Fixture, userID uuid.UUID) usertypes.GetSubscriptionsResponse {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fixture.UsersURL+"/users/subscriptions", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body usertypes.GetSubscriptionsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func hasUser(subscriptions usertypes.GetSubscriptionsResponse, userID uuid.UUID) bool {
	for _, u := range subscriptions {
		if u.Id == userID {
			return true
		}
	}
	return false
}

func TestChatsCreateDirectChatCreated(t *testing.T) {
	t.Parallel()
	dumpChatsLogs(t)

	fixture := globalFixture

	requester := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, requester.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	// Для успешного createChat по контракту target должен быть подписан на requester.
	subResp := subscribeUser(t, fixture, target.UserID, requester.UserID)
	t.Cleanup(func() { _ = subResp.Body.Close() })
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	resp := createDirectChatRequest(t, fixture, requester.UserID, target.UserID)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var chat chattypes.Chat
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&chat))
	require.NotEqual(t, uuid.Nil, chat.Id)
	require.Len(t, chat.Participants, 2)

	// CheckSubscription должен автоматически создать обратную подписку requester -> target.
	subs := getSubscriptions(t, fixture, requester.UserID)
	require.True(t, hasUser(subs, target.UserID))
}

func TestChatsCreateDirectChatForbiddenWhenNoIncomingSubscription(t *testing.T) {
	t.Parallel()
	dumpChatsLogs(t)

	fixture := globalFixture

	requester := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, requester.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	resp := createDirectChatRequest(t, fixture, requester.UserID, target.UserID)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestChatsCreateDirectChatConflictWhenAlreadyExists(t *testing.T) {
	t.Parallel()
	dumpChatsLogs(t)

	fixture := globalFixture

	requester := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, requester.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	subResp := subscribeUser(t, fixture, target.UserID, requester.UserID)
	t.Cleanup(func() { _ = subResp.Body.Close() })
	require.Equal(t, http.StatusCreated, subResp.StatusCode)

	first := createDirectChatRequest(t, fixture, requester.UserID, target.UserID)
	t.Cleanup(func() { _ = first.Body.Close() })
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := createDirectChatRequest(t, fixture, requester.UserID, target.UserID)
	t.Cleanup(func() { _ = second.Body.Close() })
	require.Equal(t, http.StatusConflict, second.StatusCode)
}

func TestChatsCreateDirectChatUnauthorized(t *testing.T) {
	t.Parallel()
	dumpChatsLogs(t)

	fixture := globalFixture
	participantID := uuid.New()

	payload, err := json.Marshal(chattypes.CreateChatRequest{ParticipantId: participantID})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fixture.ChatsURL+"/chats/", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
