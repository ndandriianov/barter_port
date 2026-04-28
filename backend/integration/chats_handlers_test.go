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

func updateUserName(t *testing.T, fixture *Fixture, userID uuid.UUID, name string) {
	t.Helper()

	payload := []byte(`{"name":"` + name + `"}`)
	req, err := http.NewRequest(http.MethodPatch, fixture.UsersURL+"/users/me", bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func requireParticipantNames(t *testing.T, participants []chattypes.Participant, expected map[uuid.UUID]string) {
	t.Helper()

	require.Len(t, participants, len(expected))

	actual := make(map[uuid.UUID]string, len(participants))
	for _, participant := range participants {
		require.NotNil(t, participant.UserName)
		actual[participant.UserId] = *participant.UserName
	}

	for userID, expectedName := range expected {
		actualName, ok := actual[userID]
		require.True(t, ok, "participant %s not found in response", userID)
		require.Equal(t, expectedName, actualName)
	}
}

func listChats(t *testing.T, fixture *Fixture, userID uuid.UUID) []chattypes.Chat {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fixture.ChatsURL+"/chats/", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, userID))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var chats []chattypes.Chat
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&chats))
	return chats
}

func TestChatsCreateDirectChatCreated(t *testing.T) {
	t.Parallel()
	dumpChatsLogs(t)

	fixture := globalFixture

	requester := registerAuthUser(t, fixture)
	target := registerAuthUser(t, fixture)
	waitForUsersProjection(t, fixture, requester.UserID)
	waitForUsersProjection(t, fixture, target.UserID)

	requesterName := "Requester " + requester.UserID.String()[:8]
	targetName := "Target " + target.UserID.String()[:8]
	updateUserName(t, fixture, requester.UserID, requesterName)
	updateUserName(t, fixture, target.UserID, targetName)

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

	requesterChats := listChats(t, fixture, requester.UserID)
	require.Len(t, requesterChats, 1)
	require.Equal(t, chat.Id, requesterChats[0].Id)
	requireParticipantNames(t, requesterChats[0].Participants, map[uuid.UUID]string{
		requester.UserID: requesterName,
		target.UserID:    targetName,
	})
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
